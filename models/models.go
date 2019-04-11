package models

import "math"

// MediaItem struct is
type MediaItem struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Duration    int      `json:"duration"`
	Countries   []string `json:"countries"`
	AgeValue    int      `json:"age_value"`
	Year        string   `json:"year"`
	Logo        string   `json:"logo"`
	Rating      float64  `json:"rating"`
	Description string   `json:"description"`
	Genres      []string `json:"genres"`
	Persons     []struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"persons"`
	Packages   []int `json:"packages"`
	AssetTypes []int `json:"asset_types"`
}

func eqStrArr(lhs []string, rhs []string) bool {
	if len(lhs) != len(rhs) {
		return false
	}
	for i := range lhs {
		if lhs[i] != rhs[i] {
			return false
		}
	}
	return true
}

func eqIntArr(lhs []int, rhs []int) bool {
	if len(lhs) != len(rhs) {
		return false
	}
	for i := range lhs {
		if lhs[i] != rhs[i] {
			return false
		}
	}
	return true
}

func (this *MediaItem) EQ(other MediaItem) bool {
	ret := this.ID == other.ID &&
		this.Name == other.Name &&
		this.Type == other.Type &&
		this.Duration == other.Duration &&
		eqStrArr(this.Countries, other.Countries) &&
		this.AgeValue == other.AgeValue &&
		this.Year == other.Year &&
		this.Logo == other.Logo &&
		math.Abs(this.Rating-other.Rating) < math.E &&
		this.Description == other.Description &&
		eqStrArr(this.Genres, other.Genres) &&
		eqIntArr(this.Packages, other.Packages) &&
		eqIntArr(this.AssetTypes, other.AssetTypes)
	return ret
}

type EPGItem struct {
	ID          int         `json:"id"`
	Name        string      `json:"name"`
	AgeValue    int         `json:"age_value"`
	StartTime   int         `json:"start_time"`
	EndTime     int         `json:"end_time"`
	Genre       string      `json:"genre"`
	Description string      `json:"description"`
	Logo        string      `json:"logo"`
	Channel     ChannelInfo `json:"channel"`
	LocationID  int         `json:"location_id"`
}

func (this *EPGItem) EQ(other EPGItem) bool {
	return this.ID == other.ID &&
		this.Name == other.Name &&
		this.AgeValue == other.AgeValue &&
		this.StartTime == other.StartTime &&
		this.EndTime == other.EndTime &&
		this.Genre == other.Genre &&
		this.Description == other.Description &&
		this.Logo == other.Logo &&
		this.Channel.ID == other.Channel.ID &&
		this.Channel.Logo == other.Channel.Logo &&
		this.Channel.Name == other.Channel.Name &&
		this.LocationID == other.LocationID

}

type ChannelInfo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Logo string `json:"logo"`
}

type ItemUnion struct {
	Type      string     `json:"type"`
	MediaItem *MediaItem `json:"media_item,omit_empty"`
	EPGItem   *EPGItem   `json:"epg,omit_empty"`
}
