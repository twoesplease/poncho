package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/Jeffail/gabs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"
)

func main() {

	type Citystate struct {
		City  string
		State string
	}

	checkinput := bufio.NewReader(os.Stdin)
	fmt.Println("What's the name of the city you live in?")
	cityname, _ := checkinput.ReadString('\n')
	cityname = strings.TrimSuffix(cityname, "\n")

	fmt.Println("And what's the 2-letter all-caps abbreviation for the state?")
	stateabbrev, _ := checkinput.ReadString('\n')

	// msg := "Got it.  You live in {{.City}}, {{.State}}."
	geourl := "https://maps.googleapis.com/maps/api/geocode/json?address={{.City}},+{{.State}}&key=MY_API_KEY"

	subin := Citystate{cityname, stateabbrev}
	tmpl, err := template.New("geourl").Parse(geourl)
	// Create a variable that implements io.Writer so that I don't have to write the output to standard output
	var blah bytes.Buffer
	err = tmpl.Execute(&blah, subin)

	if err != nil {
		fmt.Println(err)
	}

	latLongClient := http.Client{
		Timeout: time.Second * 2,
	}

	req, err := http.NewRequest(http.MethodPost, geourl, nil)
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

	type Latlong struct {
		Latitude  string
		Longitude string
	}

	parsedJson, err := gabs.ParseJSON([]byte(output))

	latitude := parsedJson.Path("results.geometry.location.lat").Data().(interface{})
	longitude := parsedJson.Path("results.geometry.location.lng").Data().(interface{})
	// Convert latitude and longitude to strings so they can be interpolated into weatherurl
	// as part of the Latlong struct
	stringifiedLatitude := fmt.Sprint(latitude)
	latWithoutLeftBracket := strings.TrimPrefix(stringifiedLatitude, "[")
	latWithoutBrackets := strings.TrimSuffix(latWithoutLeftBracket, "]")
	stringifiedLongitude := fmt.Sprint(longitude)
	longWithoutLeftBracket := strings.TrimPrefix(stringifiedLongitude, "[")
	longWithoutBrackets := strings.TrimSuffix(longWithoutLeftBracket, "]")

	preparsed_weatherurl := "https://api.darksky.net/forecast/MY_API_KEY/{{.Latitude}},{{.Longitude}}"

	substitute := Latlong{latWithoutBrackets, longWithoutBrackets}
	tmpl2, err2 := template.New("preparsed_weatherurl").Parse(preparsed_weatherurl)
	// Create a variable that implements io.Writer so that I don't have to write the output to standard output
	var parsed_weatherurl bytes.Buffer
	err = tmpl2.Execute(&parsed_weatherurl, substitute)

	if err2 != nil {
		fmt.Println(err)
	}

	stringified_parsed_weatherurl := fmt.Sprint(&parsed_weatherurl)

	weatherClient := http.Client{
		Timeout: time.Second * 2,
	}

	weatherreq, err := http.NewRequest(http.MethodGet, stringified_parsed_weatherurl, nil)
	if err != nil {
		log.Fatal(err)
	}

	weatherreq.Header.Set("User-Agent", "go-weather-app")

	weatherres, errGet := weatherClient.Do(weatherreq)
	if errGet != nil {
		log.Fatal(errGet)
	}

	weatheroutput, errRead := ioutil.ReadAll(weatherres.Body)
	if errRead != nil {
		log.Fatal(errRead)
	}

	parsedweatherJson, err := gabs.ParseJSON([]byte(weatheroutput))

	minutecast := parsedweatherJson.Path("minutely.summary").Data() //.(interface{})
	fmt.Println("Minutecast: ", minutecast)

}
