package twitter

import (
	"fmt"

	"github.com/NaturalSelectionLabs/RSS3-PreGod/indexer/pkg/util"
	"github.com/NaturalSelectionLabs/RSS3-PreGod/shared/database/datatype"
	"github.com/NaturalSelectionLabs/RSS3-PreGod/shared/pkg/config"
	"github.com/NaturalSelectionLabs/RSS3-PreGod/shared/pkg/httpx"
	lop "github.com/samber/lo/parallel"
	"github.com/valyala/fastjson"
)

const endpoint = "https://api.twitter.com/1.1"

func GetUserShow(name string) (*UserShow, error) {
	key := util.GotKey("round-robin", "Twitter", config.Config.Indexer.Twitter.Tokens)
	authorization := fmt.Sprintf("Bearer %s", key)

	var headers = map[string]string{
		"Authorization": authorization,
	}

	url := fmt.Sprintf("%s/users/show.json?screen_name=%s", endpoint, name)

	response, err := httpx.Get(url, headers)
	if err != nil {
		return nil, err
	}

	var parser fastjson.Parser
	parsedJson, err := parser.Parse(string(response))

	if err != nil {
		return nil, err
	}

	userShow := new(UserShow)

	userShow.Name = string(parsedJson.GetStringBytes("name"))
	userShow.ScreenName = string(parsedJson.GetStringBytes("screen_name"))
	userShow.Description = string(parsedJson.GetStringBytes("description"))
	userShow.Entities = string(parsedJson.GetObject("entities").MarshalTo(nil))

	return userShow, nil
}

// TODO: offset?
// See https://developer.twitter.com/en/docs/twitter-api/v1/tweets/timelines/api-reference/get-statuses-user_timeline
func GetTimeline(name string, count uint32) ([]*ContentInfo, error) {
	key := util.GotKey("round-robin", "Twitter", config.Config.Indexer.Twitter.Tokens)
	authorization := fmt.Sprintf("Bearer %s", key)

	var headers = map[string]string{
		"Authorization": authorization,
	}

	url := fmt.Sprintf("%s/statuses/user_timeline.json?screen_name=%s&count=%d&exclude_replies=true", endpoint, name, count)

	response, err := httpx.Get(url, headers)

	if err != nil {
		return nil, err
	}

	contentInfos := make([]*ContentInfo, 0, 100)

	var parser fastjson.Parser

	parsedJson, err := parser.Parse(string(response))
	if err != nil {
		return nil, err
	}

	contentArray, err := parsedJson.Array()
	if err != nil {
		return contentInfos, err
	}

	cs := lop.Map(contentArray, func(contentValue *fastjson.Value, _ int) *ContentInfo {
		contentInfo := new(ContentInfo)

		contentInfo.PreContent = formatTweetText(contentValue)
		contentInfo.Timestamp = string(contentValue.GetStringBytes("created_at"))
		contentInfo.Hash = string(contentValue.GetStringBytes("id_str"))
		contentInfo.Link = fmt.Sprintf("https://twitter.com/%s/status/%s", name, contentInfo.Hash)
		contentInfo.Attachments = getTweetAttachments(contentValue)
		contentInfo.ScreenName = string(contentValue.GetStringBytes("user", "screen_name"))

		return contentInfo
	})

	contentInfos = append(contentInfos, cs...)

	return contentInfos, nil
}

func getTweetAttachments(contentInfo *fastjson.Value) datatype.Attachments {
	attachments := datatype.Attachments{}

	// media
	extendedEntitiesValue := contentInfo.Get("extended_entities")
	if extendedEntitiesValue != nil {
		medias := extendedEntitiesValue.GetArray("media")

		as := lop.Map(medias, func(media *fastjson.Value, _ int) datatype.Attachment {
			// TODO: video
			mediaUrl := string(media.GetStringBytes("media_url_https"))

			contentHeader, _ := httpx.GetContentHeader(mediaUrl)

			a := datatype.Attachment{
				Type:        "media",
				Address:     mediaUrl,
				MimeType:    contentHeader.MIMEType,
				SizeInBytes: contentHeader.SizeInByte,
			}

			return a
		})

		attachments = append(attachments, as...)
	}

	// quote address
	quotedStatusValue := contentInfo.Get("quoted_status")
	if quotedStatusValue != nil {
		quotedStatusId := string(quotedStatusValue.GetStringBytes("id_str"))
		quotedStatusLink := fmt.Sprintf("https://twitter.com/%s/status/%s", string(quotedStatusValue.GetStringBytes("user", "screen_name")), quotedStatusId)

		qa := datatype.Attachment{
			Type:     "quote_address",
			Content:  quotedStatusLink,
			MimeType: "text/uri-list",
		}

		attachments = append(attachments, qa)

		text := string(quotedStatusValue.GetStringBytes("text"))
		qc := datatype.Attachment{
			Type:     "quote_content",
			Content:  text,
			MimeType: "text/plain",
		}

		attachments = append(attachments, qc)
	}

	return attachments
}

func formatTweetText(contentValue *fastjson.Value) string {
	text := contentValue.GetStringBytes("text")

	return string(text)
}
