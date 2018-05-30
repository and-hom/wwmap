package huskytm

import (
	. "github.com/and-hom/wwmap/cron/catalog-sync/common"
	wp "github.com/and-hom/go-wordpress"
	"net/http"
	"fmt"
	"time"
	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"strconv"
	"github.com/and-hom/wwmap/lib/dao"
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

	return &HuskytmReportProvider{client:client, tags:tagsById, images:make(map[int][]dao.Img)}, nil
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
	images map[int][]dao.Img
}

func (this *HuskytmReportProvider) Close() error {
	return nil
}

func (this *HuskytmReportProvider) ReportsSince(key string) ([]dao.VoyageReport, string, error) {
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
		return []dao.VoyageReport{}, key, err
	}

	ids := []dao.VoyageReport{}
	latest := time.Unix(0, 0)
	for i := 0; i < len(posts); i++ {
		post := &posts[i]
		dateModified, err := time.Parse(TIME_FORMAT, post.Modified)
		if err != nil {
			return []dao.VoyageReport{}, key, err
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

		ids = append(ids, dao.VoyageReport{
			RemoteId:fmt.Sprintf("%d", post.ID),
			Title: post.Title.Rendered,
			Url:post.Link,
			DatePublished: datePublished,
			DateModified: dateModified,
			Source:SOURCE,
			Tags: tags,
		})
	}
	return ids, latest.Format(TIME_FORMAT), nil
}

func (this *HuskytmReportProvider) cacheImages() error {
	params := emptyMap()
	params["media_type"] = "image"
	media, err := paginate(func(p interface{}) ([]interface{}, *http.Response, []byte, error) {
		m, r, b, e := this.client.Media().List(params)
		res := make([]interface{}, len(m))
		for i := 0; i < len(m); i++ {
			res[i] = m[i]
		}
		return res, r, b, e
	}, params)
	if err != nil {
		return err
	}

	for i := 0; i < len(media); i++ {
		m := media[i].(wp.Media)

		if m.Post <= 0 {
			continue
		}

		tm, err := time.Parse(TIME_FORMAT, m.Date)
		if err != nil {
			return err
		}
		fmt.Println("Title: ", m.Title.Rendered)
		fmt.Println("Description: ", m.Description.Rendered)
		fmt.Println("Caption: ", m.Caption.Rendered)
		fmt.Println("AltText: ", m.AltText)

		this.images[m.Post] = append(this.images[m.Post], dao.Img{
			Source:SOURCE,
			RemoteId: fmt.Sprintf("%d", m.ID),
			Url: m.MediaDetails.Sizes.Large.SourceURL,
			PreviewUrl: m.MediaDetails.Sizes.Thumbnail.SourceURL,
			DatePublished: tm,
			LabelsForSearch: []string{
				m.Title.Rendered,
				m.Description.Rendered,
				m.Caption.Rendered,
				m.AltText,
			},
		})
	}
	//fmt.Printf("Cached images are: %v+\n", this.images)
	return nil
}

func (this *HuskytmReportProvider) Images(reportId string) ([]dao.Img, error) {
	if len(this.images) == 0 {
		this.cacheImages()
	}
	reportIdInt, err := strconv.ParseInt(reportId, 10, 32)
	if err != nil {
		log.Error("Report id should be integer: ", reportId)
		return []dao.Img{}, err
	}
	imgs, found := this.images[int(reportIdInt)]
	if found {
		return imgs, nil
	}
	return []dao.Img{}, nil
}

func emptyMap() map[string]interface{} {
	return make(map[string]interface{})
}