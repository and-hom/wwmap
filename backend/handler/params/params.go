package params

import (
	"fmt"
	"net/http"
	"strconv"
)

type Params struct {
	// Skip some spot id
	Skip int64
	// Show only some spot
	Only int64
	// Show only river
	River int64
	// Show only region
	Region int64
	// Show only country
	Country int64
	// Type of spot and river hyperlinks
	LinkType string
	// Maximum category of displayed rivers
	MaximumCategory int
}

func Parse(req *http.Request) (Params,  *http.Request, error) {
	params := Params{}
	var err error

	skipIdStr := req.FormValue("skip")
	if skipIdStr != "" {
		params.Skip, err = strconv.ParseInt(skipIdStr, 10, 64)
		if err != nil {
			return params, req, fmt.Errorf("Can not parse skip id %s %s", skipIdStr, err)
		}
	}

	onlyIdStr := req.FormValue("only")
	if onlyIdStr != "" {
		params.Only, err = strconv.ParseInt(onlyIdStr, 10, 64)
		if err != nil {
			return params, req, fmt.Errorf("Can not parse only id %s %s", onlyIdStr, err)
		}
	}

	riverIdStr := req.FormValue("river")
	if riverIdStr != "" {
		params.River, err = strconv.ParseInt(riverIdStr, 10, 64)
		if err != nil {
			return params, req, fmt.Errorf("Can not parse river id %s %s", riverIdStr, err)
		}
	}

	regionIdStr := req.FormValue("region")
	if regionIdStr != "" {
		params.Region, err = strconv.ParseInt(regionIdStr, 10, 64)
		if err != nil {
			return params, req, fmt.Errorf("Can not parse region id %s %s", regionIdStr, err)
		}
	}

	countryIdStr := req.FormValue("country")
	if countryIdStr != "" {
		params.Country, err = strconv.ParseInt(countryIdStr, 10, 64)
		if err != nil {
			return params, req, fmt.Errorf("Can not parse country id %s %s", countryIdStr, err)
		}
	}

	return params, req, nil
}
