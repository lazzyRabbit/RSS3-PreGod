package model

import (
	"github.com/NaturalSelectionLabs/RSS3-PreGod/shared/pkg/constants"
	"gorm.io/datatypes"
)

// `link_list` model.
type LinkList struct {
	Base BaseModel `gorm:"embedded"`

	LinkListID string `gorm:"primaryKey;type:text;column:link_list_id"`

	RSS3ID string `gorm:"type:text;column:rss3_id"` // owner id

	LinkType constants.LinkTypeID `gorm:"type:int"`

	Metadata datatypes.JSON `gorm:"type:jsonb"`
}