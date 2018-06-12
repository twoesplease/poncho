package main

import (
	"fmt"
	"github.com/Jeffail/gabs"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func main() {

	url := "https://maps.googleapis.com/maps/api/geocode/json?address=Atlanta,+GA&key=MY_API_KEY"

	latLongClient := http.Client{
		Timeout: time.Second * 2,
	}

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("User-Agent", "hobby-weather-app")

	res, getErr := latLongClient.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
	}

	output, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	parsedJson, err := gabs.ParseJSON([]byte(output))

	latitude, didlatwork := parsedJson.Path("results.geometry.location.lat").Data().(interface{})
	longitude, didlongwork := parsedJson.Path("results.geometry.location.lng").Data().(interface{})
	fmt.Println("\n*** Did you remember to add the API key back? ***\n")
	fmt.Println("Latitude: ", latitude, didlatwork)
	fmt.Println("Longitude: ", longitude, didlongwork)
}
