package holidays

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jame-developer/aeontrac/aeontrac"
	"io"
	"net/http"
	url2 "net/url"
	"strings"
	"time"
)

type (
	// holidayResponse represents the response from the public holidays API
	holidayResponse struct {
		Id        string `json:"id"`
		StartDate string `json:"startDate"`
		EndDate   string `json:"endDate"`
		Type      string `json:"type"`
		Name      []struct {
			Language string `json:"language"`
			Text     string `json:"text"`
		} `json:"name"`
		Nationwide   bool `json:"nationwide"`
		Subdivisions []struct {
			Code      string `json:"code"`
			ShortName string `json:"shortName"`
		} `json:"subdivisions"`
	}
	holidayItem struct {
		Name    string
		Regions []string
	}
)

// LoadHolidays loads the public holidays from the API and saves them to the database
// It returns the list of public holidays and an error if any.
// It uses the provided configuration to load the public holidays.
// It uses the Open Holidays API to load the public holidays. See https://openholidaysapi.org/
func LoadHolidays(config aeontrac.PublicHolidaysConfig, year int) (map[string]holidayItem, error) {
	u, _ := url2.Parse(config.APIURL)
	u.RawQuery = url2.Values{"countryIsoCode": {config.Country}, "validFrom": {fmt.Sprintf("%d-01-01", year)}, "validTo": {fmt.Sprintf("%d-12-31", year)}}.Encode()
	var res, err = http.Get(u.String())
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(res.Body)
	_ = res.Body.Close()
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		fmt.Printf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
		return nil, errors.New("failed to load public holidays")
	}

	var holidays []holidayResponse
	err = json.Unmarshal(body, &holidays)
	if err != nil {
		return nil, err
	}
	holidayDays := map[string]holidayItem{}

	for _, holiday := range holidays {
		start, err := time.Parse(time.DateOnly, holiday.StartDate)
		if err != nil {
			return nil, err
		}
		end, err := time.Parse(time.DateOnly, holiday.EndDate)
		if err != nil {
			return nil, err
		}

		for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
			dateKey := d.Format(time.DateOnly)
			var holidayNames []string
			for _, name := range holiday.Name {
				holidayNames = append(holidayNames, name.Text)
			}
			var holidayRegions []string
			for _, region := range holiday.Subdivisions {
				holidayRegions = append(holidayRegions, region.Code)
			}
			holidayDays[dateKey] = holidayItem{
				Name:    strings.Join(holidayNames, ", "),
				Regions: holidayRegions,
			}
		}
	}

	return holidayDays, nil
}
