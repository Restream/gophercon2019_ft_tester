package repo

import (
	"encoding/json"
	"log"
	"os"
	"path"

	"github.com/restream/gophercon2019_ft_tester/models"
)

type DataHolder struct {
	EpgItems   map[int]models.EPGItem
	MediaItems map[int]models.MediaItem
	Ammos      map[string][]models.Ammo
}

func (dh *DataHolder) loadEPG(jsonPath string) error {
	file, err := os.Open(jsonPath)
	if err != nil {
		return err
	}
	defer file.Close()

	dec := json.NewDecoder(file)

	epgItems := make([]models.EPGItem, 0, 150000)
	if err = dec.Decode(&epgItems); err != nil {
		return err
	}
	dh.EpgItems = make(map[int]models.EPGItem)
	for _, epg := range epgItems {
		dh.EpgItems[epg.ID] = epg
	}

	log.Printf("Loaded %d epgs from %s", len(dh.EpgItems), jsonPath)
	return nil
}

func (dh *DataHolder) loadMediaItems(jsonPath string) error {
	file, err := os.Open(jsonPath)
	if err != nil {
		return err
	}
	defer file.Close()

	dec := json.NewDecoder(file)

	mediaItems := make([]models.MediaItem, 0, 50000)
	if err = dec.Decode(&mediaItems); err != nil {
		return err
	}

	dh.MediaItems = make(map[int]models.MediaItem)
	for _, mediaItem := range mediaItems {
		dh.MediaItems[mediaItem.ID] = mediaItem
	}

	log.Printf("Loaded %d media items from %s", len(dh.MediaItems), jsonPath)
	return nil
}

func (dh *DataHolder) loadAmmos(jsonPath string) error {
	file, err := os.Open(jsonPath)
	if err != nil {
		return err
	}
	defer file.Close()

	dec := json.NewDecoder(file)

	dh.Ammos = make(map[string][]models.Ammo)
	if err = dec.Decode(&dh.Ammos); err != nil {
		return err
	}

	return nil
}

func (dh *DataHolder) Init(dataDir, ammosDir string) error {
	if err := dh.loadEPG(path.Join(dataDir, "epg.json")); err != nil {
		return err
	}

	if err := dh.loadMediaItems(path.Join(dataDir, "media_items.json")); err != nil {
		return err
	}

	if ammosDir == "" {
		return nil
	}

	if err := dh.loadAmmos(path.Join(ammosDir, "ammos.json")); err != nil {
		return err
	}
	return nil
}
