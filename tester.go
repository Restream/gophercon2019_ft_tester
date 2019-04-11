package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	"./models"
	"./repo"

	"github.com/valyala/fasthttp"
)

var client *fasthttp.HostClient
var jobQueue chan Job
var stats []*Stat

var hostAddr, dataDir, ammoDir *string
var httpConnections, httpTimeout, reqCount *int
var waitGroup sync.WaitGroup
var dataHolder repo.DataHolder

type Stat struct {
	Method         string
	Lck            sync.Mutex
	TotalLatency   time.Duration
	TotalElapsed   time.Duration
	RequestsCount  int
	ConnErrors     int
	ContentErrors  int
	HTTPCodeErrors int
}

type Job struct {
	URL  string
	Ammo models.Ammo
	Stat *Stat
}

func main() {
	hostAddr = flag.String("host", "127.0.0.1:8080", "Base URL of search application server")
	dataDir = flag.String("datadir", "data", "Dir, where datafiles are located")
	ammoDir = flag.String("ammodir", "data", "Dir, where ammos are located")
	httpConnections = flag.Int("conn", 2, "Number of simulatenous connections")
	httpTimeout = flag.Int("timeout", 10, "Requests timeout in seconds")
	reqCount = flag.Int("count", 10, "Requests count per each method")
	flag.Parse()

	if err := dataHolder.Init(*dataDir, *ammoDir); err != nil {
		log.Fatal(err)
	}

	initWorkers()
	log.Printf("Starting test API server at %s\n", *hostAddr)
	for method, ammos := range dataHolder.Ammos {
		if len(ammos) != 0 {
			makeJobs(method, ammos)
		}
	}

	log.Println("Finish ")
	teardownWorkers()
	printStats()
}

func makeJobs(method string, ammos []models.Ammo) {
	stat := &Stat{Method: method}
	stats = append(stats, stat)
	tStart := time.Now()

	for i := 0; i < 30000; i++ {
		ammo := ammos[i%len(ammos)]
		jobQueue <- Job{
			fmt.Sprintf("http://%s/%s?%s", *hostAddr, method, ammo.Args),
			ammo,
			stat,
		}
	}
	stat.Lck.Lock()
	defer stat.Lck.Unlock()
	stat.TotalElapsed = time.Now().Sub(tStart)
}

func respValidator(err error, latency time.Duration, respCode int, bodyData []byte, job *Job) {
	job.Stat.Lck.Lock()
	job.Stat.RequestsCount++
	job.Stat.TotalLatency += latency

	if err != nil {
		job.Stat.ConnErrors++
		job.Stat.Lck.Unlock()
		return
	}
	if respCode != job.Ammo.HTTPCode {
		job.Stat.HTTPCodeErrors++
		job.Stat.Lck.Unlock()
		return
	}
	job.Stat.Lck.Unlock()

	score := 0
	switch job.Stat.Method {
	case "/api/v1/media_items":
		err = validateMediaItemsResp(bodyData, job.Ammo)
	case "/api/v1/epg":
		err = validateEPGItemsResp(bodyData, job.Ammo)
	case "/api/v1/search":
		score, err = validateSearchResp(bodyData, job.Ammo)
	}

	job.Stat.Lck.Lock()
	if err != nil {
		job.Stat.ContentErrors++
	}
	_ = score
	job.Stat.Lck.Unlock()
}

func validateSearchResp(body []byte, ammo models.Ammo) (int, error) {
	resp := models.GetSearchResponce{}

	if err := json.Unmarshal(body, &resp); err != nil {
		return 0, err
	}
	foundMediaItems := make(map[int]int)
	foundEPGItems := make(map[int]int)
	for _, item := range resp.Items {
		switch item.Type {
		case "media_item":
			foundMediaItems[item.MediaItem.ID] = 1
		case "epg":
			foundEPGItems[item.EPGItem.ID] = 1
		}
	}

	foundCount := 0
	for _, expID := range ammo.Ids {
		if expID.Type == models.ContentTypeEPG {
			if _, ok := foundEPGItems[expID.ID]; ok {
				foundCount++
			}
		}
		if expID.Type == models.ContentTypeMediaItem {
			if _, ok := foundMediaItems[expID.ID]; ok {

				foundCount++
			}
		}
	}
	if foundCount < 100*len(ammo.Ids)/70 {
		return 0, fmt.Errorf("At least 70%% of answers should match reference")
	}

	return 1, nil
}

func validateEPGItemsResp(body []byte, ammo models.Ammo) error {
	resp := models.GetEPGItemsResponce{}

	if err := json.Unmarshal(body, &resp); err != nil {
		return err
	}
	if len(resp.Items) != len(ammo.Ids) {
		return fmt.Errorf("Expected %d items, but got %d", len(ammo.Ids), len(resp.Items))
	}
	if resp.TotalItems != ammo.TotalItems {
		return fmt.Errorf("Expected %d total_items, but got %d", len(ammo.Ids), len(resp.Items))
	}

	for i, epg := range resp.Items {
		if ammo.Ids[i].ID != epg.ID {
			return fmt.Errorf("Expected item with ID=%d , but got ID=%d in position %d", ammo.Ids[i], epg.ID, i)
		}
		if !epg.EQ(dataHolder.EpgItems[epg.ID]) {
			return fmt.Errorf("Item with ID %d content mismatch", epg.ID)
		}
	}

	return nil
}

func validateMediaItemsResp(body []byte, ammo models.Ammo) error {
	resp := models.GetMediaItemsResponce{}

	if err := json.Unmarshal(body, &resp); err != nil {
		return err
	}
	if len(resp.Items) != len(ammo.Ids) {
		return fmt.Errorf("Expected %d items, but got %d", len(ammo.Ids), len(resp.Items))
	}
	if resp.TotalItems != ammo.TotalItems {
		return fmt.Errorf("Expected %d total_items, but got %d", len(ammo.Ids), len(resp.Items))
	}

	for i, mediaItem := range resp.Items {
		if ammo.Ids[i].ID != mediaItem.ID {
			return fmt.Errorf("Expected item with ID=%d , but got ID=%d in position %d", ammo.Ids[i], mediaItem.ID, i)
		}
		if !mediaItem.EQ(dataHolder.MediaItems[mediaItem.ID]) {
			return fmt.Errorf("Item with ID %d content mismatch", mediaItem.ID)
		}
	}
	return nil
}

func initWorkers() {
	jobQueue = make(chan Job, 100)

	client = &fasthttp.HostClient{Addr: *hostAddr}

	for i := 0; i < *httpConnections; i++ {
		waitGroup.Add(1)
		go func() {
			for job := range jobQueue {
				doRequest(job)
			}
			waitGroup.Done()
		}()
	}
}

func teardownWorkers() {
	close(jobQueue)
	waitGroup.Wait()
}

func doRequest(job Job) {
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(job.URL)

	resp := fasthttp.AcquireResponse()
	t := time.Now()
	err := client.DoTimeout(req, resp, time.Duration(*httpTimeout)*time.Second)

	respValidator(err, time.Now().Sub(t), resp.StatusCode(), resp.Body(), &job)

	fasthttp.ReleaseRequest(req)
	fasthttp.ReleaseResponse(resp)
}

func printStat(stat *Stat) {
	fmt.Printf(
		" %-30s%10d%10d%16v%16d%16d%16d\n",
		stat.Method,
		stat.RequestsCount,
		int(float64(stat.RequestsCount)/stat.TotalElapsed.Seconds()),
		stat.TotalLatency/time.Duration(stat.RequestsCount),
		stat.ConnErrors,
		stat.HTTPCodeErrors,
		stat.ContentErrors,
	)
}

func printStats() {
	fmt.Printf(" %-30s%10s%10s%16s%16s%16s%16s\n", "Method", "Requests", "RPS", "Avg Latency", "Socket errors", "Wrong http code", "Wrong content")
	fmt.Printf(" %-30s%10s%10s%16s%16s%16s%16s\n", "------", "--------", "---", "-----------", "-------------", "--------------", "--------------")
	for _, s := range stats {
		printStat(s)
	}
}
