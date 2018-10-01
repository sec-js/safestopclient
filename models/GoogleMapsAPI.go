package models

import (
	"fmt"
	"net/http"
	"io/ioutil"
	url2 "net/url"
	"encoding/json"
)


type GoogleGeocodeResponse struct {
	Result []struct {
		AddressComponents []struct {
			LongName  string   `json:"long_name"`
			ShortName string   `json:"short_name"`
			Types     []string `json:"types"`
		} `json:"address_components"`

		Geometry struct {
			Location struct {
				Latitude  float64 `json:"lat"`
				Longitude float64 `json:"lng"`
			} `json:"location"`

			LocationType string `json:"location_type"`
		} `json:"geometry"`

		FormattedAddress string   `json:"formatted_address"`
		PlaceId          string   `json:"place_id"`
		Types            []string `json:"types"`
	} `json:"results"`
	Status string `json"status"`
}

type Coordinate struct {
	Latitude float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func GoogleServer() string {
	return "https://maps.googleapis.com/maps/api"
}

func GoogleAPIKey() string {
	return "&key=AIzaSyDiObDov0rg4zTixsC4E1bFaxjMf3gSwRQ"
}

func Geocode(address string, postal_code string) *Coordinate {

	url := fmt.Sprintf("%s/geocode/json?address=%s,%s%s", GoogleServer(), url2.QueryEscape(address), url2.QueryEscape(postal_code), GoogleAPIKey())

	rs, err := http.Get(url)
	// Process response
	if err != nil {
		return nil
	}
	defer rs.Body.Close()

	bodyBytes, err := ioutil.ReadAll(rs.Body)
	if err != nil {
		return nil
	}

	gr := GoogleGeocodeResponse{}
	err = json.Unmarshal(bodyBytes, &gr)
	if (err != nil){
		return nil
	}

	if gr.Status == "OK" {
		if len(gr.Result) > 0 {
			if gr.Result[0].Geometry.LocationType == "ROOFTOP" || gr.Result[0].Geometry.LocationType == "RANGE_INTERPOLATED" {
				return &Coordinate{
					gr.Result[0].Geometry.Location.Latitude,
					gr.Result[0].Geometry.Location.Longitude,
				}
			}
		}
	}

	return nil
}
