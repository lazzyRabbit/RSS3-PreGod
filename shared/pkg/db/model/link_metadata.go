package model

import (
	"github.com/NaturalSelectionLabs/RSS3-PreGod/shared/pkg/constants"
	"gorm.io/datatypes"
)

// `link_metadata` model.
type Link_Metadata struct {
	Base BaseModel `gorm:"embedded"`

	LinkMetadataID string `gorm:"primaryKey;type:text;column:link_metadata_id"`

	RSS3ID       string `gorm:"type:text"`
	TargetRSS3ID string `gorm:"type:text"`

	LinkType constants.LinkTypeID `gorm:"type:int"`

	Metadata datatypes.JSON `gorm:"type:jsonb"`
}