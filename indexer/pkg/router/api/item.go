package api

import (
	"net/http"

	"github.com/NaturalSelectionLabs/RSS3-PreGod/shared/pkg/constants"
	"github.com/gin-gonic/gin"
)

type GetItemRequest struct {
	Identity   string               `form:"proof" binding:"required"`
	PlatformID constants.PlatformID `form:"platform_id" binding:"required"`
	NetworkID  constants.NetworkID  `form:"network_id"`
	ItemType   constants.ItemType   `form:"item_type"`
}

func GetItemHandlerFunc(c *gin.Context) {
	request := GetItemRequest{}
	if err := c.ShouldBind(&request); err != nil {
		return
	}

	// TODO Query data

	c.JSON(http.StatusOK, request)
}
