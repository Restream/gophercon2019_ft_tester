package models

const (
	ContentTypeMediaItem = 1
	ContentTypeEPG       = 2
)

type ContentID struct {
	Type int `json:"type"`
	ID   int `json:"id"`
}

type Ammo struct {
	Args       string      `json:"args"`
	HTTPCode   int         `json:"code"`
	TotalItems int         `json:"total_items"`
	Ids        []ContentID `json:"ids"`
}
