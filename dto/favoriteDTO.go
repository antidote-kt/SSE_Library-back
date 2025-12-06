package dto

type FavoriteDTO struct {
	SourceID uint64 `json:"sourceId"`
	UserID   uint64 `json:"userId"`
	Type     string `json:"type"`
}
