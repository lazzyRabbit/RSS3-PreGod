package handler

import (
	"fmt"
	"net/http"

	"github.com/NaturalSelectionLabs/RSS3-PreGod/hub/internal/api"
	"github.com/NaturalSelectionLabs/RSS3-PreGod/hub/internal/middleware"
	"github.com/NaturalSelectionLabs/RSS3-PreGod/hub/internal/protocol"
	"github.com/NaturalSelectionLabs/RSS3-PreGod/shared/database"
	"github.com/NaturalSelectionLabs/RSS3-PreGod/shared/pkg/constants"
	"github.com/NaturalSelectionLabs/RSS3-PreGod/shared/pkg/rss3uri"
	"github.com/gin-gonic/gin"
)

func GetInstanceHandlerFunc(c *gin.Context) {
	var instance rss3uri.Instance

	if platformInstance, err := middleware.GetPlatformInstance(c); err == nil {
		instance = platformInstance
	} else if networkInstance, err := middleware.GetNetworkInstance(c); err == nil {
		instance = networkInstance
	} else {
		api.SetError(c, api.ErrorIndexer, err)

		return
	}

	if err := database.QueryInstance(
		database.DB,
		instance.GetIdentity(),
		constants.ProfileSourceIDCrossbell.Int(),
	); err != nil {
		api.SetError(c, api.ErrorDatabase, err)

		return
	}

	instanceList := protocol.NewInstanceList(instance)

	c.JSON(http.StatusOK, protocol.File{
		Identifier: fmt.Sprintf("%s?%s", rss3uri.New(instance).String(), c.Request.URL.Query().Encode()),
		Total:      int64(len(instanceList)),
		List:       instanceList,
	})
}
