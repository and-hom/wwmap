package huskytm

//go:generate go-bindata -pkg $GOPACKAGE -o templates.go -prefix templates/ ./templates

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"net/http"
	"strconv"
)

const SOURCE string = "huskytm"
const API_BASE string = "https://huskytm.ru/wp-json/wp/v2"
const CUSTOM_FIELDS_API_BASE string = "https://huskytm.ru/wp-json/acf/v3"
const TIME_FORMAT string = "2006-01-02T15:04:05"
const TOTAL_PAGES_HEADER = "X-WP-TotalPages"

func emptyMap() map[string]interface{} {
	return make(map[string]interface{})
}

func paginate(get func(interface{}) ([]interface{}, *http.Response, []byte, error), params map[string]interface{}) ([]interface{}, error) {
	result := []interface{}{}
	for page := 1; page < 100000; page++ {
		params["page"] = fmt.Sprintf("%d", page)

		res, resp, b, err := get(params)
		if err != nil {
			log.Errorf("Can not paginate: %s", string(b))
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			log.Errorf("Can not paginate: %s", string(b))
			return nil, errors.New("Can not paginate")
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
