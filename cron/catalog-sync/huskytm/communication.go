package huskytm

import (
	. "github.com/and-hom/wwmap/cron/catalog-sync/common"
	wp "github.com/and-hom/go-wordpress"
	"net/http"
	"fmt"
	"time"
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/lib/model"
	"github.com/pkg/errors"
	"strconv"
)

const SOURCE string = "huskytm"
const API_BASE string = "https://huskytm.ru/wp-json/wp/v2"
const TIME_FORMAT string = "2006-01-02T15:04:05"
const REPORT_CATEGORY string = "8"
const TOTAL_PAGES_HEADER = "X-WP-TotalPages"

func GetReportProvider(login, password string) (ReportProvider, error) {
	client := wp.NewClient(&wp.Options{
		BaseAPIURL: API_BASE, // example: `http://192.168.99.100:32777/wp-json/wp/v2`
		Username:   login,
		Password:   password,
	})

	tags, err := paginate(func(p interface{}) ([]interface{}, *http.Response, []byte, error) {
		t, r, b, e := client.Tags().List(p)
		res := make([]interface{}, len(t))
		for i := 0; i < len(t); i++ {
			res[i] = t[i]
		}
		return res, r, b, e

	}, emptyMap())

	if err != nil {
		return nil, fmt.Errorf("Connection failed: %s", err.Error())
	}
	tagsById := make(map[int]string)
	for _, t := range tags {
		tag := t.(wp.Tag)
		tagsById[tag.Id] = tag.Name
	}

	return &HuskytmReportProvider{client:client, tags:tagsById}, nil
}

func paginate(get func(interface{}) ([]interface{}, *http.Response, []byte, error), params map[string]interface{}) ([]interface{}, error) {
	result := []interface{}{}
	for page := 1; page < 100000; page++ {
		params["page"] = fmt.Sprintf("%d", page)

		res, resp, b, err := get(params)
		if resp.StatusCode != http.StatusOK {
			log.Errorf("Can not paginate: %s", string(b))
			return nil, errors.New("Can not paginate")
		}
		if err != nil {
			log.Errorf("Can not paginate: %s", string(b))
			return nil, err
		}
		result = append(result, res...)

		totalPagesStr := resp.Header.Get(TOTAL_PAGES_HEADER)
		totalPages, err := strconv.ParseInt(totalPagesStr, 10, 32)
		if err != nil {
			log.Errorf("Can not parse header \"%s\" = %s", TOTAL_PAGES_HEADER, totalPagesStr)
		}
		if page >= int(totalPages) {
			break
		}
	}
	return result, nil
}

type HuskytmReportProvider struct {
	client *wp.Client
	tags   map[int]string
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

		tags := []string{}
		for _, tagId := range post.Tags {
			tag, found := this.tags[tagId]
			if found {
				tags = append(tags, tag)
			} else {
				log.Fatalf("Unknown tag with id %d in %v+", tagId, this.tags)
			}
		}

		datePublished, err := time.Parse(TIME_FORMAT, post.Date)

		ids = append(ids, model.VoyageReport{
			RemoteId:fmt.Sprintf("%d", post.ID),
			Url:post.Link,
			DatePublished: datePublished,
			DateModified: dateModified,
			Source:SOURCE,
			Tags: tags,
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
		if err != nil {
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