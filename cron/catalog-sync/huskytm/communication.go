package huskytm

import (
	. "github.com/and-hom/wwmap/cron/catalog-sync/common"
	wp "github.com/and-hom/go-wordpress"
	"net/http"
	"fmt"
	"time"
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/lib/model"
)

const SOURCE string = "huskytm"
const API_BASE string = "https://huskytm.ru/wp-json/wp/v2"
const TIME_FORMAT string = "2006-01-02T15:04:05"
const REPORT_CATEGORY string = "8"

func GetReportProvider(login, password string) (ReportProvider, error) {
	client := wp.NewClient(&wp.Options{
		BaseAPIURL: API_BASE, // example: `http://192.168.99.100:32777/wp-json/wp/v2`
		Username:   login,
		Password:   password,
	})
	_, resp, _, _ := client.Users().Me(emptyMap())
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Connection failed: %v+", resp)
	}

	return &HuskytmReportProvider{client:client}, nil
}

type HuskytmReportProvider struct {
	client *wp.Client
}

func (this *HuskytmReportProvider) Close() error {
	return nil
}

func (this *HuskytmReportProvider) ReportsSince(key string) ([]model.VoyageReport, string, error) {
	startTs, err := time.Parse(TIME_FORMAT, key)
	if err != nil {
		log.Warnf("Can not parse start key %s as time: use time(0) as start", key)
		startTs = time.Unix(0, 0)
	}

	params := emptyMap()
	params["context"] = "view"
	params["orderby"] = "modified"
	params["categories"] = REPORT_CATEGORY

	posts, _, responseBytes, err := this.client.Posts().List(params)
	if err != nil {
		log.Error(string(responseBytes))
		return []model.VoyageReport{}, key, err
	}

	ids := []model.VoyageReport{}
	latest := time.Unix(0, 0)
	for i := 0; i < len(posts); i++ {
		post := &posts[i]
		dateModified, err := time.Parse(TIME_FORMAT, post.Modified)
		if err != nil {
			return []model.VoyageReport{}, key, err
		}
		if ! dateModified.After(startTs) {
			continue
		}
		if latest.Before(dateModified) {
			latest = dateModified
		}

		datePublished, err := time.Parse(TIME_FORMAT, post.Date)

		ids = append(ids, model.VoyageReport{
			RemoteId:fmt.Sprintf("%d", post.ID),
			Url:post.Link,
			DatePublished: datePublished,
			DateModified: dateModified,
			Source:SOURCE,
		})
	}
	return ids, latest.Format(TIME_FORMAT), nil
}

func (this *HuskytmReportProvider) Images(reportId int) ([]model.Img, error) {
	params := emptyMap()
	params["media_type"] = "image"
	media, _, responseBytes, err := this.client.Media().List(params)
	if err != nil {
		log.Error(string(responseBytes))
		return []model.Img{}, err
	}

	imgs := make([]model.Img, len(media))
	for i := 0; i < len(media); i++ {
		tm, err := time.Parse(TIME_FORMAT, media[i].Modified)
		if err!=nil {
			return []model.Img{}, err
		}
		imgs[i] = model.Img{
			Source:SOURCE,
			Url: media[i].MediaDetails.Sizes.Large.SourceURL,
			PreviewUrl: media[i].MediaDetails.Sizes.Thumbnail.SourceURL,
			DateTaken: tm,
		}
	}
	return imgs, nil
}

func emptyMap() map[string]interface{} {
	return make(map[string]interface{})
}