package arweave

import (
	"errors"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/NaturalSelectionLabs/RSS3-PreGod/shared/database"
	"github.com/NaturalSelectionLabs/RSS3-PreGod/shared/database/datatype"
	"github.com/NaturalSelectionLabs/RSS3-PreGod/shared/database/model"
	"github.com/NaturalSelectionLabs/RSS3-PreGod/shared/pkg/constants"
	"github.com/NaturalSelectionLabs/RSS3-PreGod/shared/pkg/logger"
	"github.com/NaturalSelectionLabs/RSS3-PreGod/shared/pkg/rss3uri"
)

var ErrTimeout = errors.New("received timeout")
var ErrInterrupt = errors.New("received interrupt")

type crawlConfig struct {
	fromHeight    int64
	confirmations int64
	step          int64
	minStep       int64
	sleepInterval time.Duration
	nextRoundTime time.Time
}

type crawler struct {
	identity  ArAccount
	interrupt chan os.Signal
	cfg       *crawlConfig
}

func NewCrawler(identity ArAccount, crawlCfg *crawlConfig) *crawler {
	return &crawler{
		identity,
		make(chan os.Signal, 1),
		crawlCfg,
	}
}

func (ar *crawler) run() error {
	for {
		// handle interrupt
		if ar.gotInterrupt() {
			return ErrInterrupt
		}

		if ar.cfg.nextRoundTime.After(time.Now()) {
			time.Sleep(1 * time.Second)

			continue
		}

		startBlockHeight := ar.cfg.fromHeight
		endBlockHeight := ar.cfg.fromHeight + ar.cfg.step

		// check latest confirmed block height
		latestConfirmedBlockHeight, err := GetLatestBlockHeightWithConfirmations(ar.cfg.confirmations)
		if err != nil {
			logger.Errorf("get latest block error: %v", err)

			return err
		}

		if latestConfirmedBlockHeight <= endBlockHeight {
			logger.Info("catch up with the latest block height...")

			ar.cfg.nextRoundTime = ar.cfg.nextRoundTime.Add(ar.cfg.sleepInterval)

			// use minStep if we are at the end of the chain
			ar.cfg.step = ar.cfg.minStep

			logger.Info("arweave catch up with the latest block height")

			continue
		}

		logger.Infof("Getting articles from [%d] to [%d], with step [%d] and latest confirmed block height [%d]",
			startBlockHeight, endBlockHeight, ar.cfg.step, latestConfirmedBlockHeight)
		ar.parseMirrorArticles(startBlockHeight, endBlockHeight, ar.identity)

		// set new fromHeight for next round
		ar.cfg.fromHeight = endBlockHeight

		// sleep 0.5 second per round
		time.Sleep(500 * time.Millisecond)
	}
}

//TODO: I think it will be the same as other crawler formats in the future,
// and it will return to an abstract and unified crawler
func (ar *crawler) parseMirrorArticles(from, to int64, owner ArAccount) error {
	articles, err := GetMirrorContents(from, to, owner)
	if err != nil {
		logger.Errorf("GetMirrorContents error: [%v]", err)

		return err
	}

	items := make([]model.Note, 0, len(articles))

	for _, article := range articles {
		attachment := datatype.Attachments{
			{
				Type:     "body",
				Content:  article.Content,
				MimeType: "text/markdown",
			},
		}

		tsp := time.Unix(article.Timestamp, 0)

		// ignore empty item
		if article.Author == "" {
			continue
		}

		// ellipsis the content as summary

		summary := article.Content
		if maxSummaryLength := 400; len(summary) > maxSummaryLength { // TODO: define the max length specifically in protocol?
			summary = string([]rune(summary)[:maxSummaryLength]) + "..."
		}

		author := rss3uri.NewAccountInstance(article.Author, constants.PlatformSymbolEthereum).UriString()
		note := model.Note{
			Identifier: rss3uri.NewNoteInstance(article.TxHash, constants.NetworkSymbolArweaveMainnet).UriString(),
			Owner:      author,
			RelatedURLs: []string{
				"https://arweave.net/" + article.TxHash,
				"https://mirror.xyz/" + article.Author + "/" + article.OriginalDigest,
			},
			Tags:            constants.ItemTagsMirrorEntry.ToPqStringArray(),
			Authors:         []string{author},
			Title:           article.Title,
			Summary:         summary,
			Attachments:     database.MustWrapJSON(attachment),
			Source:          constants.NoteSourceNameMirrorEntry.String(),
			MetadataNetwork: constants.NetworkSymbolArweaveMainnet.String(),
			MetadataProof:   article.TxHash,
			Metadata:        database.MustWrapJSON(map[string]interface{}{}),
			DateCreated:     tsp,
			DateUpdated:     tsp,
		}

		items = append(items, note)
	}

	tx := database.DB.Begin()
	defer tx.Rollback()

	if items != nil && len(items) > 0 {
		if _, dbErr := database.CreateNotes(tx, items, true); dbErr != nil {
			return dbErr
		}
	}

	if err = tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

func (ar *crawler) Start() error {
	signal.Notify(ar.interrupt, os.Interrupt)

	log.Println("Starting Arweave crawler...")

	if err := ar.run(); err != nil {
		logger.Errorf("arweave crawler errro [%v]", err)

		return err
	}

	return nil
}

func (ar *crawler) gotInterrupt() bool {
	select {
	case <-ar.interrupt:
		signal.Stop(ar.interrupt)

		return true
	default:
		return false
	}
}
