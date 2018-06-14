package tlib

import (
	"time"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/PuerkitoBio/goquery"
	"github.com/and-hom/wwmap/cron/catalog-sync/common"
	"net/url"
	"net/http"
	"github.com/pkg/errors"
	"strings"
	log "github.com/Sirupsen/logrus"
	"regexp"
	"github.com/goodsign/monday"
	"github.com/and-hom/wwmap/lib/util"
)

const SOURCE string = "tlib"
const SITE_URL = "http://www.tlib.ru"
const TIME_FORMAT string = "02.01.2006 15:04:05"
const DATE_FORMAT string = "02.01.2006"

const URL_ID_RE = "id=(\\d+)"
const TITILE_RE = "Маршрут:[\\s\n]*?(.*?)[\\s\n]*?Тип:"
const YEAR_MONTH_RE = "год:[\\s\n]*?(\\d{4})[\\s\n]*?(;[\\s\n]*?месяц:[\\s\n]*?(январь|февраль|март|апрель|май|июнь|июль|август|сентябрь|октябрь|ноябрь|декабрь)[\\s\n]*?)?"
const RIVER_RE = "([рp]\\.\\s?|река\\s+)(.*?)\\s?[=$\\(]"
const PUBLISHED_DATE_RE = "Загружено:\\s*(\\d{1,2}\\.\\d{1,2}\\.\\d{4} \\d{1,2}:\\d{1,2}:\\d{1,2})"

var zero = util.ZeroDateUTC()

var urlIdRe = regexp.MustCompile(URL_ID_RE)
var titleRe = regexp.MustCompile(TITILE_RE)
var yearMonthRe = regexp.MustCompile(YEAR_MONTH_RE)
var riverRe = regexp.MustCompile(RIVER_RE)
var publishedDateRe = regexp.MustCompile(PUBLISHED_DATE_RE)

func GetReportProvider() (common.ReportProvider, error) {
	return &TlibReportsProvider{
		client:http.Client{},
	}, nil
}

type TlibReportsProvider struct {
	client http.Client
}

type ViewState struct {
	ViewState       string
	EventValidation string
	Cookie          string
	NextPage        bool
}

func getViewState(document *goquery.Document, resp *http.Response) (ViewState, error) {
	eventValidation, exists := document.Find("form #__EVENTVALIDATION").Attr("value")
	if !exists {
		return ViewState{}, errors.New("Can not find value of field __EVENTVALIDATION")
	}

	viewState, exists := document.Find("form #__VIEWSTATE").Attr("value")
	if !exists {
		return ViewState{}, errors.New("Can not find value of field __VIEWSTATE")
	}
	return ViewState{
		ViewState:viewState,
		EventValidation:eventValidation,
		Cookie:resp.Header.Get("Set-Cookie"),
		NextPage:false,
	}, nil
}

func (this TlibReportsProvider) ReportsSince(t time.Time) ([]dao.VoyageReport, time.Time, error) {
	getPageReq, err := http.NewRequest("GET", SITE_URL, nil)
	if err != nil {
		return []dao.VoyageReport{}, t, err
	}
	getPageResp, err := this.client.Do(getPageReq)
	if err != nil {
		return []dao.VoyageReport{}, t, err
	}

	document, err := goquery.NewDocumentFromReader(getPageResp.Body)
	if err != nil {
		return []dao.VoyageReport{}, t, err
	}

	viewState, err := getViewState(document, getPageResp)
	if err != nil {
		return []dao.VoyageReport{}, t, err
	}

	reports := make([]dao.VoyageReport, 0)
	maxTime := t

	for page, hasNext := 0, true; hasNext; page++ {
		hasNext, viewState, err = this.queryData(page, viewState, t, func(r dao.VoyageReport) {
			reports = append(reports, r)
			if r.DateModified.After(maxTime) {
				maxTime = r.DateModified
			}
		})
		if err != nil {
			return []dao.VoyageReport{}, t, err
		}
	}
	return reports, maxTime, nil
}

func (this TlibReportsProvider) queryData(pageNum int, viewState ViewState, dateFrom time.Time, onReport func(r dao.VoyageReport)) (bool, ViewState, error) {
	log.Infof("Querying page %d", pageNum)
	form := url.Values{}
	if viewState.NextPage {
		form.Add("__EVENTTARGET", "LinkNext")
		form.Add("__EVENTARGUMENT", "")
	} else {
		form.Add("ctl22", "Найти")
	}
	form.Add("__VIEWSTATE", viewState.ViewState)
	form.Add("__EVENTVALIDATION", viewState.EventValidation)
	form.Add("ctl00", "")
	form.Add("ctl01", "")
	form.Add("ctl03", "")
	form.Add("ctl04", "")
	form.Add("ctl05", "")
	form.Add("ctl06", "")
	form.Add("ctl07", "водный")
	form.Add("ctl09", "")
	form.Add("ctl11", "")
	form.Add("ctl13", "")
	form.Add("ctl15", "")
	form.Add("ctl17", "")
	form.Add("ctl19", "")
	form.Add("DatepickerAfter", time.Now().Format(DATE_FORMAT))
	form.Add("DatepickerBefore", dateFrom.Format(DATE_FORMAT))
	form.Add("SortedByList", "Shifr")
	form.Add("SortedByDest", "ASC")

	req, err := http.NewRequest("POST", SITE_URL, strings.NewReader(form.Encode()))
	if err != nil {
		return false, viewState, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Upgrade-Insecure-Requests", "1")
	req.Header.Add("Referer", "http://www.tlib.ru/")
	req.Header.Add("Origin", "http://www.tlib.ru/")
	req.Header.Add("Host", "www.tlib.ru")
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Ubuntu Chromium/66.0.3359.181 Chrome/66.0.3359.181 Safari/537.36")
	req.Header.Add("Cookie", viewState.Cookie)

	resp, err := this.client.Do(req)
	if err != nil {
		return false, viewState, err
	}
	document, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return false, viewState, err
	}
	rows := document.Find("#DataGrid1 tr")
	log.Infof("Found %d rows", rows.Length())
	rows.Each(func(i int, row *goquery.Selection) {
		// skip table header
		if row.Find("td:nth-of-type(1)").Text() == "N" {
			return
		}

		reportUrl, found := row.Find("td:nth-of-type(2) a").Attr("href")
		if !found {
			log.Warn("Can not find href for %s", row.Text())
			return
		}
		reportUrl = strings.Replace(reportUrl, "..", SITE_URL, 1)

		remoteId := urlIdRe.FindStringSubmatch(reportUrl)[1]

		datePublished, err := this.getDatePublished(reportUrl)
		if err != nil {
			log.Errorf("Can not find date published for %s %s", reportUrl, err.Error())
			datePublished = zero
		}

		descritionText := row.Find("td:nth-of-type(2)").Text()

		submatch := titleRe.FindStringSubmatch(descritionText)
		if len(submatch) < 2 {
			log.Fatalf("Illegal title: %s\n%s", descritionText, row.Text())
		}
		title := submatch[1]
		tags := make([]string, 0)
		for _, river := range riverRe.FindAllStringSubmatch(title, -1) {
			tags = append(tags, river[2])
		}

		dateOfTrip, err := parseDateOfTrip(descritionText)
		if err != nil {
			log.Error("Can not parse date of trip ", err)
			dateOfTrip = zero
		}

		report := dao.VoyageReport{
			RemoteId: remoteId,
			Title: title,
			Source: SOURCE,
			Url: reportUrl,
			Tags: tags,
			DateOfTrip: dateOfTrip,
			DatePublished: datePublished,
			DateModified: datePublished,
		}
		onReport(report)
	})

	_, exists := document.Find("#LinkNext").Attr("href")
	if exists {
		nextViewState, err := getViewState(document, resp)
		if err != nil {
			return false, viewState, err
		}
		if nextViewState.Cookie == "" {
			nextViewState.Cookie = viewState.Cookie
		}
		nextViewState.NextPage = true
		return true, nextViewState, nil
	}
	return false, viewState, nil
}

func parseDateOfTrip(desc string) (time.Time, error) {
	dateOfTheTripSubmatch := yearMonthRe.FindStringSubmatch(strings.ToLower(desc))
	if len(dateOfTheTripSubmatch) < 2 {
		log.Warn("Date of the trip not dound: ", desc)
		return zero, nil
	}

	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		return zero, err
	}

	monthStr := "Январь"
	if len(dateOfTheTripSubmatch) >= 4 && dateOfTheTripSubmatch[3] != "" {
		monthStr = strings.Title(dateOfTheTripSubmatch[3])
	}

	dateOfTheTripStr := strings.ToLower(dateOfTheTripSubmatch[1] + " " + monthStr)
	dateOfTheTrip, err := monday.ParseInLocation("2006 January", dateOfTheTripStr, loc, monday.LocaleRuRU)
	if err != nil {
		return zero, err
	}

	return dateOfTheTrip, nil
}

func (this TlibReportsProvider) getDatePublished(reportUrl string) (time.Time, error) {
	req, err := http.NewRequest("GET", reportUrl, nil)
	if err != nil {
		return zero, err
	}
	resp, err := this.client.Do(req)
	if err != nil {
		return zero, err
	}
	document, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return zero, err
	}
	descriptionBodyText := document.Find("#Label1").Text()
	found := publishedDateRe.FindStringSubmatch(descriptionBodyText)
	if len(found) < 2 {
		log.Warnf("Date not found for description %s %s:\n%s", reportUrl, found, descriptionBodyText)
		return zero, nil
	}
	dateStr := found[1]
	return time.Parse(TIME_FORMAT, dateStr)
}

func (this TlibReportsProvider) Images(reportId string) ([]dao.Img, error) {
	return []dao.Img{}, nil
}

func (this *TlibReportsProvider) Close() error {
	return nil
}

