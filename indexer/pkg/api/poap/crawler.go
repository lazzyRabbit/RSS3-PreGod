package poap

import (
	"fmt"
	"time"

	"github.com/NaturalSelectionLabs/RSS3-PreGod/indexer/pkg/api/nft_utils"
	"github.com/NaturalSelectionLabs/RSS3-PreGod/indexer/pkg/crawler"
	"github.com/NaturalSelectionLabs/RSS3-PreGod/shared/database"
	"github.com/NaturalSelectionLabs/RSS3-PreGod/shared/database/datatype"
	"github.com/NaturalSelectionLabs/RSS3-PreGod/shared/database/model"
	"github.com/NaturalSelectionLabs/RSS3-PreGod/shared/pkg/constants"
	"github.com/NaturalSelectionLabs/RSS3-PreGod/shared/pkg/logger"
	"github.com/NaturalSelectionLabs/RSS3-PreGod/shared/pkg/rss3uri"
)

type poapCrawler struct {
	crawler.DefaultCrawler
}

func NewPoapCrawler() crawler.Crawler {
	return &poapCrawler{
		crawler.DefaultCrawler{
			Assets: []model.Asset{},
			Notes:  []model.Note{},
		},
	}
}

func (pc *poapCrawler) Work(param crawler.WorkParam) error {
	if param.NetworkID != constants.NetworkIDGnosisMainnet {
		return fmt.Errorf("network is not gnosis")
	}

	poapResps, err := GetActions(param.Identity)
	if err != nil {
		return fmt.Errorf("poap [%s] get actions error:", err)
	}

	owner := rss3uri.NewAccountInstance(param.OwnerID, param.OwnerPlatformID.Symbol()).UriString()
	author := rss3uri.NewAccountInstance(param.Identity, constants.PlatformSymbolEthereum).UriString()

	for _, item := range poapResps {
		tsp, err := item.GetTsp()
		if err != nil {
			logger.Error(tsp, err)
			tsp = time.Now()
		}

		id := ContractAddress + "-" + item.TokenId
		note := model.Note{
			Identifier:  rss3uri.NewNoteInstance(id, constants.NetworkSymbolGnosisMainnet).UriString(),
			Owner:       owner,
			RelatedURLs: []string{fmt.Sprintf("https://app.poap.xyz/token/%s", item.TokenId)},
			Tags:        constants.ItemTagsNFTPOAP.ToPqStringArray(),
			Authors:     []string{author},
			Title:       item.PoapEventInfo.Name,
			Summary:     item.PoapEventInfo.Description,
			Attachments: database.MustWrapJSON(datatype.Attachments{
				{
					Type:     "preview",
					Address:  item.PoapEventInfo.ImageUrl,
					MimeType: "image/png",
				},
				{
					Type:     "external_url",
					Content:  item.PoapEventInfo.EventUrl,
					MimeType: "text/uri-list",
				},
				{
					Type:     "start_date",
					Content:  item.PoapEventInfo.StartDate,
					MimeType: "text/plain",
				},
				{
					Type:     "end_date",
					Content:  item.PoapEventInfo.EndDate,
					MimeType: "text/plain",
				},
				{
					Type:     "expiry_date",
					Content:  item.PoapEventInfo.ExpiryDate,
					MimeType: "text/plain",
				},
			}),
			Source:          constants.NoteSourceNameEthereumNFT.String(),
			MetadataNetwork: constants.NetworkSymbolGnosisMainnet.String(),
			MetadataProof:   id, // TODO: this should be the tx hash in note actually?
			Metadata: database.MustWrapJSON(map[string]interface{}{
				"from": "0x0",
				"to":   item.Owner,
			}),
			DateCreated: tsp,
			DateUpdated: tsp,
		}

		pc.Notes = append(pc.Notes, note)

		asset := note
		asset.Identifier = rss3uri.NewAssetInstance(id, constants.NetworkSymbolGnosisMainnet).UriString()
		asset.Source = constants.AssetSourceNameEthereumNFT.String()

		pc.Assets = append(pc.Assets, model.Asset(asset))

		if err := nft_utils.CompleteMimeTypesForItems(pc.Notes, pc.Assets, pc.Profiles); err != nil {
			logger.Error("poap complete mime types error:", err)
		}
	}

	return nil
}
