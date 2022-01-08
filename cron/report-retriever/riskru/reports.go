package riskru

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
	"github.com/and-hom/wwmap/cron/report-retriever/common"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/lib/util"
	"golang.org/x/net/html/charset"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const SOURCE string = "riskru"
const SITE_URL = "https://www.risk.ru/blog/activity/water"
const SITE_URL_BASE = "https://www.risk.ru"
const DEFAULT_AUTHOR = "risk.ru"
const ENCODING string = "utf-8"

var DATE_FORMAT = common.CreateDateExtractor("\\d{2}\\.\\d{2}\\.\\d{4}", "02.01.2006")

var zero = util.ZeroDateUTC()

func GetReportProvider() (common.ReportProvider, error) {
	return &RiskReportsProvider{
		client: http.Client{},
	}, nil
}

type RiskReportsProvider struct {
	client http.Client
}

func (this *RiskReportsProvider) SourceId() string {
	return SOURCE
}

func (this *RiskReportsProvider) ReportsSince(t time.Time) ([]dao.VoyageReport, time.Time, error) {
	reports := make([]dao.VoyageReport, 0)
	maxTime := zero

	for page, hasNext := 1, true; hasNext; page++ {
		_hasNext, err := this.ReportsSinceForPage(t, page, func(r dao.VoyageReport) {
			reports = append(reports, r)
			if r.DateModified.After(maxTime) {
				maxTime = r.DateModified
			}
		})
		if t.After(maxTime) {
			break
		}
		hasNext = _hasNext
		if err != nil {
			return []dao.VoyageReport{}, t, err
		}
	}

	return reports, maxTime, nil
}

func (this *RiskReportsProvider) ReportsSinceForPage(t time.Time, page int, onReport func(r dao.VoyageReport)) (bool, error) {
	url := fmt.Sprintf("%s?Topic_page=%d", SITE_URL, page)
	document, err := this.doGet(url)
	if err != nil {
		return false, err
	}

	rows := document.Find(".commonPost .rightPart .header")
	rows.Each(func(i int, row *goquery.Selection) {
		titleBlock := row.Find("h2 a")
		if titleBlock == nil {
			return
		}

		title := titleBlock.Text()
		reportUrl, exists := titleBlock.Attr("href")
		remoteId := "0"
		if !exists {
			reportUrl = ""
		} else {
			urlParts := strings.Split(reportUrl, "/")
			remoteId = urlParts[len(urlParts)-1]
		}

		authorAndTimeBlock := row.Find(".userInfo")
		var datePublished *time.Time = nil
		if authorAndTimeBlock != nil {
			timeStr := authorAndTimeBlock.Text()
			d, ok := DATE_FORMAT.GetDate(timeStr)
			if ok {
				datePublished = &d
			} else {
				datePublished = nil
			}
		}

		// skip pages before startDate
		if datePublished==nil || datePublished.Before(t) {
			return
		}

		authorBlock := row.Find(".userInfo a")
		author := DEFAULT_AUTHOR
		if authorBlock != nil {
			author = authorBlock.Text()
		}

		articleUrl := SITE_URL_BASE + reportUrl
		tags := this.getTags(articleUrl)

		onReport(dao.VoyageReport{
			RemoteId:      remoteId,
			LinkedEntity: dao.LinkedEntity{IdTitle: dao.IdTitle{Title: title}},
			Author:        author,
			Source:        SOURCE,
			Url:           articleUrl,
			Tags:          tags,
			DateOfTrip:    nil,
			DatePublished: datePublished,
			DateModified:  util.PtrToTime(datePublished),
		})
	})

	hasNext := false
	lastBtn := document.Find(".yiiPager .last a")
	if lastBtn != nil {
		lastHref, exists := lastBtn.Attr("href")
		if exists {
			parts := strings.Split(lastHref, "Topic_page=")
			if len(parts) == 2 {
				i, err := strconv.ParseInt(parts[1], 10, 64)
				if err == nil {
					hasNext = (page < int(i))
				}
			}
		}
	}

	return hasNext, nil
}

func (this *RiskReportsProvider) getTags(articleUrl string) []string {
	doc, err := this.doGet(articleUrl)
	if err != nil {
		logrus.Error(err)
		return []string{}
	}
	tags := []string{}

	tagLinks := doc.Find(".tags .tagsIcon a")
	tagLinks.Each(func(i int, selection *goquery.Selection) {
		tags = append(tags, selection.Text())
	})
	return tags
}

func (this *RiskReportsProvider) Images(reportId string) ([]dao.Img, error) {
	return []dao.Img{}, nil
}

func (this *RiskReportsProvider) Close() error {
	return nil
}

func (this *RiskReportsProvider) doGet(url string) (*goquery.Document, error) {
	logrus.Info("Fetch ", url)
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
