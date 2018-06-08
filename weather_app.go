package main

import (
	"encoding/json"
	"fmt"
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

	convertJson := []byte(output)

	var holdresponsebody interface{}

	jsonErr := json.Unmarshal(convertJson, &holdresponsebody)
	if jsonErr != nil {
		fmt.Println(jsonErr)
		return
	}
	latlong := holdresponsebody.(map[string]interface{})
	fmt.Print("Latitude: ")
	// latlong2  isn't working because I'm not getting my type assertion right
	latlong1 := latlong["results"].(interface{})
	latlong2 := latlong1["geometry"].(map[string]interface{})
	fmt.Println(latlong2)
	// This doesn't work, only fmt.Println(latlong["results"]
	// fmt.Println(latlong["results"]["geometry"]["location"]["lat"])

}
