package models

type DocumentTagMap struct {
	DocumentID uint64
	TagID      uint64
}

func (DocumentTagMap) TableName() string {
	return "document_tag_map"
}
