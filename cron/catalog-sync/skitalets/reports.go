package skitalets

import (
	"crypto/md5"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
	"github.com/and-hom/wwmap/cron/catalog-sync/common"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/lib/util"
	"golang.org/x/net/html/charset"
	"hash"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const SOURCE string = "skitalets"
const SITE_URL = "http://www.skitalets.ru/tourism-types/?materials=0&turizm=18290"
const SITE_URL_BASE = "http://www.skitalets.ru"
const DEFAULT_AUTHOR = "skitalets.ru"
const ENCODING string = "utf-8"
const RIVER_RE = "[^ео](р\\.\\s?|рек[аеу]\\s+)([А-Яа-яA-Za-z\\s\\.-]+)"

var DATE_FORMAT = []common.DateExtractor{
	common.CreateDateExtractor("\\d{2}\\.\\d{2}\\.\\d{4}", "02.01.2006"),
	common.CreateDateExtractor("\\d{2}\\.\\d{2}\\s+\\d{4}", "02.01 2006"),
	common.CreateDateExtractor("\\d{2}\\.\\d{2}\\.\\d{2}", "02.01.06"),
	common.CreateDateExtractor("\\d{4}", "2006"),
}
var separatorRe = regexp.MustCompile("[\\s,]")
var riverRe = regexp.MustCompile(RIVER_RE)

var zero = util.ZeroDateUTC()

func GetReportProvider() (common.ReportProvider, error) {
	return &SkitaletsReportsProvider{
		client:    http.Client{},
		rateLimit: util.NewRateLimit(100 * time.Millisecond),
		hash:      md5.New(),
	}, nil
}

type SkitaletsReportsProvider struct {
	client    http.Client
	rateLimit util.RateLimit
	hash      hash.Hash
}

func (this *SkitaletsReportsProvider) SourceId() string {
	return SOURCE
}

func (this SkitaletsReportsProvider) ReportsSince(t time.Time) ([]dao.VoyageReport, time.Time, error) {
	document, err := this.doGet(SITE_URL)
	if err != nil {
		return []dao.VoyageReport{}, t, err
	}

	reports := make([]dao.VoyageReport, 0)
	maxTime := t

	rows := document.Find(".card")
	rows.Each(func(i int, row *goquery.Selection) {
		if row.Find("thead").Length() > 0 {
			return
		}

		datePublished := this.parseDateOfTrip(row.Find(".card-date").Text())
		if datePublished == nil || datePublished.Before(t) {
			return
		}

		link := row.Find(".card-title a")
		reportUrl, found := link.Attr("href")
		if !found {
			log.Warnf("Can not find href for %s", link.Text())
			return
		}
		title := link.Text()
		author := row.Find(".author span a").Text()
		if strings.TrimSpace(author) == "" {
			author = DEFAULT_AUTHOR
		}

		url := SITE_URL_BASE + reportUrl

		tags, err := this.getTags(url)
		if err != nil {
			log.Error("Can't get tags: ", err)
			return
		}

		_, err = io.WriteString(this.hash, reportUrl)
		if err != nil {
			log.Error("Can not get md5 sum for url ", reportUrl)
			return
		}
		reports = append(reports, dao.VoyageReport{
			RemoteId:      fmt.Sprintf("%x", this.hash.Sum(nil)),
			LinkedEntity: dao.LinkedEntity{IdTitle: dao.IdTitle{Title: title}},
			Author:        author,
			Source:        SOURCE,
			Url:           url,
			Tags:          tags,
			DateOfTrip:    nil,
			DatePublished: datePublished,
			DateModified:  util.PtrToTime(datePublished),
		})
	})

	return reports, maxTime, nil
}

func (this SkitaletsReportsProvider) doGet(url string) (*goquery.Document, error) {
	getPageReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	getPageResp, err := this.client.Do(getPageReq)
	if err != nil {
		return nil, err
	}

	encodedReader, err := charset.NewReaderLabel(ENCODING, getPageResp.Body)
	if err != nil {
		return nil, err
	}
	document, err := goquery.NewDocumentFromReader(encodedReader)
	if err != nil {
		return nil, err
	}
	return document, nil
}

func (this SkitaletsReportsProvider) getTags(url string) ([]string, error) {
	log.Infof("Get tags for %s", url)
	tags := make([]string, 0)
	document, err := this.doGet(url)
	if err != nil {
		return []string{}, err
	}

	b := document.Find("b")
	if b != nil {
		b.Each(collectTags(boldHeader(func(tag string) {
			tags = append(tags, tag)
		})))
	}

	strong := document.Find("strong")
	if strong != nil {
		strong.Each(collectTags(boldHeader(func(tag string) {
			tags = append(tags, tag)
		})))
	}

	p := document.Find("p")
	if p != nil {
		p.Each(collectTags(containsHeader(func(tag string) {
			tags = append(tags, tag)
		})))
	}

	log.Infof("Tags are: %v for %s", tags, url)
	return tags, nil
}

func collectTags(onTagElement func(selection *goquery.Selection)) func(int, *goquery.Selection) {
	return func(i int, selection *goquery.Selection) {
		if strings.Contains(selection.Text(), "Маршрут:") ||
			strings.Contains(strings.ToLower(selection.Text()), "нитка маршрута:") ||
			strings.Contains(strings.ToLower(selection.Text()), "маршрут похода:") {

			onTagElement(selection)
		}
	}
}

func boldHeader(onTag func(tag string)) func(*goquery.Selection) {
	return func(selection *goquery.Selection) {
		if !selection.Parent().Is("p") {
			return
		}

		text := selection.Parent().Text()
		getTagsFromString(text, onTag)
	}
}

func containsHeader(onTag func(tag string)) func(*goquery.Selection) {
	return func(selection *goquery.Selection) {
		if selection.Children().Size() > 0 {
			return
		}
		text := selection.Text()
		getTagsFromString(text, onTag)
	}
}

func getTagsFromString(text string, onTag func(tag string)) {
	text = strings.Replace(text, "ё", "е", -1)
	text = strings.Replace(text, " ", " ", -1) // replace &nbsp;
	log.Debug("Text is: ", text)

	for _, river := range riverRe.FindAllStringSubmatch(text, -1) {
		onTag(strings.TrimSpace(river[2]))
	}
}

func (this SkitaletsReportsProvider) parseDateOfTrip(dateStr string) *time.Time {
	for _, df := range DATE_FORMAT {
		d, found := df.GetDate(this.replaceMonth(dateStr))
		if found {
			return &d
		}
	}
	log.Debugf("Can not parse date: %s", dateStr)
	return nil
}

func (this SkitaletsReportsProvider) Images(reportId string) ([]dao.Img, error) {
	return []dao.Img{}, nil
}

func (this *SkitaletsReportsProvider) Close() error {
	return nil
}

func (this *SkitaletsReportsProvider) replaceMonth(str string) string {
	str = strings.Replace(str, " января ", ".01.", -1)
	str = strings.Replace(str, " февраля ", ".02.", -1)
	str = strings.Replace(str, " марта ", ".03.", -1)
	str = strings.Replace(str, " апреля ", ".04.", -1)
	str = strings.Replace(str, " мая ", ".05.", -1)
	str = strings.Replace(str, " июня ", ".06.", -1)
	str = strings.Replace(str, " июля ", ".07.", -1)
	str = strings.Replace(str, " августа ", ".08.", -1)
	str = strings.Replace(str, " сентября ", ".09.", -1)
	str = strings.Replace(str, " октября ", ".10.", -1)
	str = strings.Replace(str, " ноября ", ".11.", -1)
	str = strings.Replace(str, " декабря ", ".12.", -1)
	return str
}
