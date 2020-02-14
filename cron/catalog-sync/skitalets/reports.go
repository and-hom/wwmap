package skitalets

import (
	"crypto/md5"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	log "github.com/Sirupsen/logrus"
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
const SITE_URL = "http://www.skitalets.ru/water/"
const SITE_URL_BASE = "http://www.skitalets.ru"
const DEFAULT_AUTHOR = "skitalets.ru"
const ENCODING string = "cp1251"

var DATE_FORMAT = []common.DateExtractor{
	common.CreateDateExtractor("\\d{2}\\.\\d{2}\\.\\d{4}", "02.01.2006"),
	common.CreateDateExtractor("\\d{2}\\.\\d{2}\\s+\\d{4}", "02.01 2006"),
	common.CreateDateExtractor("\\d{2}\\.\\d{2}\\.\\d{2}", "02.01.06"),
	common.CreateDateExtractor("\\d{4}", "2006"),
}
var separatorRe = regexp.MustCompile("[\\s,]")

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

	rows := document.Find(".stable tr")
	rows.Each(func(i int, row *goquery.Selection) {
		if row.Find("thead").Length() > 0 {
			return
		}

		dateOfTrip := this.parseDateOfTrip(row.Find("td:nth-of-type(5)").Text())
		if dateOfTrip.Before(t) {
			return
		}

		link := row.Find("td:nth-of-type(4) a")
		reportUrl, found := link.Attr("href")
		if !found {
			log.Warnf("Can not find href for %s", link.Text())
			return
		}
		title := link.Text()
		author := row.Find("td:nth-of-type(6)").Text()
		if strings.TrimSpace(author) == "" {
			author = DEFAULT_AUTHOR
		}

		tags := this.getTags(row.Find("td:nth-of-type(3)").Text())

		_, err := io.WriteString(this.hash, reportUrl)
		if err != nil {
			log.Error("Can not get md5 sum for url ", reportUrl)
			return
		}
		reports = append(reports, dao.VoyageReport{
			RemoteId:      fmt.Sprintf("%x", this.hash.Sum(nil)),
			Title:         title,
			Author:        author,
			Source:        SOURCE,
			Url:           SITE_URL_BASE + reportUrl,
			Tags:          tags,
			DateOfTrip:    dateOfTrip,
			DatePublished: zero,
			DateModified:  zero,
		})
	})

	return reports, maxTime, nil
}

func (this SkitaletsReportsProvider) getTags(desc string) []string {
	tags := make([]string, 0)
	for _, river := range separatorRe.Split(desc, -1) {
		tag := strings.ToLower(river)
		if tag != "" && tag != "-" {
			tags = append(tags, strings.TrimSpace(tag))
		}
	}
	return tags
}

func (this SkitaletsReportsProvider) parseDateOfTrip(dateStr string) time.Time {
	for _, df := range DATE_FORMAT {
		d, found := df.GetDate(dateStr)
		if found {
			return d
		}
	}
	log.Debug("Can not parse date: %s\n", dateStr)
	return zero
}

func (this SkitaletsReportsProvider) Images(reportId string) ([]dao.Img, error) {
	return []dao.Img{}, nil
}

func (this *SkitaletsReportsProvider) Close() error {
	return nil
}
