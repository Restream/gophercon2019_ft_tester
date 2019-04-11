package models

type GetSearchResponce struct {
	TotalItems int         `json:"total_items"`
	Items      []ItemUnion `json:"items"`
}

type GetMediaItemsResponce struct {
	TotalItems int          `json:"total_items"`
	Items      []*MediaItem `json:"items"`
}

type GetEPGItemsResponce struct {
	TotalItems int        `json:"total_items"`
	Items      []*EPGItem `json:"items"`
}
