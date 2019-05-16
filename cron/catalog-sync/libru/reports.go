package libru

import (
	"github.com/PuerkitoBio/goquery"
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/cron/catalog-sync/common"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/lib/util"
	"golang.org/x/net/html/charset"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const SOURCE string = "libru"
const SITE_URL = "http://lib.ru/TURIZM/"
const DEFAULT_AUTHOR string = "lib.ru"
const ENCODING string = "KOI8-R"

const YEAR_RE string = "(\\D|^)(\\d{4})(\\D|$)"
const RIVER_RE = "[A-ZА-Яa-zа-я0-9\\-/]{3,}"

var yearRe = regexp.MustCompile(YEAR_RE)
var riverRe = regexp.MustCompile(RIVER_RE)

var zero = util.ZeroDateUTC()

func GetReportProvider() (common.ReportProvider, error) {
	return &LibRuReportsProvider{
		client:    http.Client{},
		rateLimit: util.NewRateLimit(100 * time.Millisecond),
	}, nil
}

type LibRuReportsProvider struct {
	client    http.Client
	rateLimit util.RateLimit
}

func (this *LibRuReportsProvider) SourceId() string {
	return SOURCE
}

func (this LibRuReportsProvider) ReportsSince(t time.Time) ([]dao.VoyageReport, time.Time, error) {
	getPageReq, err := http.NewRequest("GET", SITE_URL, nil)
	if err != nil {
		return []dao.VoyageReport{}, t, err
	}
	getPageResp, err := this.client.Do(getPageReq)
	if err != nil {
		return []dao.VoyageReport{}, t, err
	}

	encodedReader, err := charset.NewReaderLabel(ENCODING, getPageResp.Body)
	if err != nil {
		return []dao.VoyageReport{}, t, err
	}
	document, err := goquery.NewDocumentFromReader(encodedReader)
	if err != nil {
		return []dao.VoyageReport{}, t, err
	}

	reports := make([]dao.VoyageReport, 0)
	maxTime := t

	rows := document.Find("li")
	rows.Each(func(i int, row *goquery.Selection) {
		if !strings.Contains(row.Find("tt > small > a > b").Text(), "огл") {
			return
		}
		link := row.Find("li > a")
		reportUrl, found := link.Attr("href")
		if !found {
			log.Warn("Can not find href for %s", link.Text())
			return
		}
		author := DEFAULT_AUTHOR
		title := link.Text()
		titleLower := strings.ToLower(title)
		if !strings.Contains(titleLower, "отчёт") && !strings.Contains(titleLower, "отчет") {
			return
		}
		dateOfTrip := this.parseDateOfTrip(title)
		if dateOfTrip == zero {
			return
		}

		tags := this.getTags(title)

		reports = append(reports, dao.VoyageReport{
			RemoteId:      reportUrl,
			Title:         title,
			Author:        author,
			Source:        SOURCE,
			Url:           SITE_URL + reportUrl,
			Tags:          tags,
			DateOfTrip:    dateOfTrip,
			DatePublished: zero,
			DateModified:  zero,
		})
	})

	return reports, maxTime, nil
}

func (this LibRuReportsProvider) getTags(title string) []string {
	tags := make([]string, 0)
	for _, river := range riverRe.FindAllStringSubmatch(title, -1) {
		tag := strings.ToLower(river[0])
		if tag != "" && tag != "отчёт" && tag != "отчет" {
			tags = append(tags, strings.TrimSpace(tag))
			splitted := strings.Split(tag, "-")
			if len(splitted) > 1 {
				for _, p := range splitted {
					tags = append(tags, strings.TrimSpace(p))
				}
			}
		}
	}
	return tags
}

func (this LibRuReportsProvider) parseDateOfTrip(title string) time.Time {
	yearFound := yearRe.FindAllStringSubmatch(title, -1)
	if len(yearFound) > 0 {
		log.Debug(yearFound)
		y, err := strconv.ParseInt(yearFound[0][2], 10, 32)
		if err != nil {
			log.Errorf("Can not parse year: %s", yearFound[0])
		} else if y > 0 {
			return time.Date(int(y), time.January, 2, 0, 0, 0, 0, time.UTC)
		}
	}
	return zero
}

func (this LibRuReportsProvider) Images(reportId string) ([]dao.Img, error) {
	return []dao.Img{}, nil
}

func (this *LibRuReportsProvider) Close() error {
	return nil
}
