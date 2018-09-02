package huskytm

import (
	. "github.com/and-hom/wwmap/cron/catalog-sync/common"
	wp "github.com/and-hom/go-wordpress"
	"net/http"
	"fmt"
	"time"
	log "github.com/Sirupsen/logrus"
	"strconv"
	"github.com/and-hom/wwmap/lib/dao"
	"regexp"
	"html"
	"github.com/and-hom/wwmap/lib/util"
)

const REPORT_CATEGORY string = "8"
const IMG_RE_1 string = "\\<img.*?title=\"(.*?)\".*?src=\"(.*?)\".*?\\>"
const IMG_RE_2 string = "\\<img.*?src=\"(.*?)\".*?title=\"(.*?)\".*?\\>"
const YEAR_RE string = "(?:\\D|:^)(2\\d{3})(?:\\D|$)"

var yearRe = regexp.MustCompile(YEAR_RE)

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

	s1 := ImgSearcher{expr:regexp.MustCompile(IMG_RE_1), titleIndex:1, urlIndex:2}
	s2 := ImgSearcher{expr:regexp.MustCompile(IMG_RE_2), titleIndex:2, urlIndex:1}
	return &HuskytmReportProvider{
		client:client,
		tags:tagsById,
		images:make([]dao.Img, 0),
		imgExprs:[]ImgSearcher{s1, s2},
		cachedImgSearchResults:make(map[int][]ImgSearchResult),
	}, nil
}

type ImgSearchResult struct {
	title string
	url   string
}
type ImgSearcher struct {
	expr       *regexp.Regexp
	urlIndex   int
	titleIndex int
}

func (this *ImgSearcher) find(str string) []ImgSearchResult {
	found := this.expr.FindAllStringSubmatch(str, -1)
	result := make([]ImgSearchResult, len(found))
	log.Info("Found ", len(found))
	for i := 0; i < len(found); i++ {
		result[i] = ImgSearchResult{
			url:found[i][this.urlIndex],
			title:html.UnescapeString(found[i][this.titleIndex]),
		}
	}
	return result
}

type HuskytmReportProvider struct {
	client                 *wp.Client
	tags                   map[int]string
	images                 []dao.Img
	cachedImgSearchResults map[int][]ImgSearchResult
	imgExprs               []ImgSearcher
}

func (this *HuskytmReportProvider) SourceId() string {
	return SOURCE
}

func (this *HuskytmReportProvider) Close() error {
	return nil
}

func (this *HuskytmReportProvider) ReportsSince(key time.Time) ([]dao.VoyageReport, time.Time, error) {
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
		if ! dateModified.After(key) {
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

		pageBody := post.Content.Rendered
		foundImgs := []ImgSearchResult{}
		for _, s := range this.imgExprs {
			foundImgs = append(foundImgs, s.find(pageBody)...)
		}
		this.cachedImgSearchResults[post.ID] = foundImgs

		datePublished, err := time.Parse(TIME_FORMAT, post.Date)

		title := post.Title.Rendered

		ids = append(ids, dao.VoyageReport{
			RemoteId:fmt.Sprintf("%d", post.ID),
			Title: title,
			Author: "Husky Team",
			Url:post.Link,
			DatePublished: datePublished,
			DateModified: dateModified,
			DateOfTrip:getYear(title),
			Source:SOURCE,
			Tags: tags,
		})
	}
	return ids, latest, nil
}

func getYear(title string) time.Time {
	yearFound := yearRe.FindAllStringSubmatch(title, -1)
	year := int64(0)
	for i := 0; i < len(yearFound); i++ {
		y, err := strconv.ParseInt(yearFound[i][1], 10, 32)
		if err != nil {
			log.Errorf("Can not parse year: %s", yearFound[1])
			continue
		}
		if y > year {
			year = y
		}
	}
	if year > 0 {
		return time.Date(int(year), time.January, 2, 0, 0, 0, 0, time.UTC)
	}
	return util.ZeroDateUTC()
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
	this.images = make([]dao.Img, len(media))

	for i := 0; i < len(media); i++ {
		m := media[i].(wp.Media)

		tm, err := time.Parse(TIME_FORMAT, m.Date)
		if err != nil {
			return err
		}

		this.images[i] = dao.Img{
			Source:SOURCE,
			RemoteId: fmt.Sprintf("%d", m.ID),
			RawUrl:m.SourceURL,
			Url: m.MediaDetails.Sizes.Large.SourceURL,
			PreviewUrl: m.MediaDetails.Sizes.Thumbnail.SourceURL,
			DatePublished: tm,
			LabelsForSearch: []string{
				m.Title.Rendered,
				//m.Description.Rendered,
				//m.Caption.Rendered,
				//m.AltText,
			},
			Type:dao.IMAGE_TYPE_IMAGE,
		}
	}
	//b,_ := json.Marshal(this.images)
	//fmt.Printf("Cached images are: \n%s\n", string(b))
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

	imgsFound := []dao.Img{}
	for _, imgSearchResult := range this.cachedImgSearchResults[int(reportIdInt)] {
		for i := 0; i < len(this.images); i++ {
			img := this.images[i]
			if imgSearchResult.url == img.RawUrl {
				img.LabelsForSearch = append([]string{imgSearchResult.title}, img.LabelsForSearch...)
				imgsFound = append(imgsFound, img)
			}
		}
	}

	return imgsFound, nil
}