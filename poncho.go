package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/Jeffail/gabs"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"
)

type SubValues struct {
	Cityname  string
	Statename string
	Apikey    string
}

func GetCity() string {
	checkinput := bufio.NewReader(os.Stdin)
	fmt.Println("What's the name of the city you live in?")
	cityname, _ := checkinput.ReadString('\n')
	cityname = strings.TrimSuffix(cityname, "\n")
	return cityname
}

var cityname = GetCity()

func GetState() string {
	checkinput := bufio.NewReader(os.Stdin)
	fmt.Println("And what's the 2-letter all-caps abbreviation for the state?")
	stateabbrev, _ := checkinput.ReadString('\n')
	stateabbrev = strings.TrimSuffix(stateabbrev, "\n")
	return stateabbrev
}

var statename = GetState()

func LoadEnvVars() {
	err := godotenv.Load(".gitignore/key.txt")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func GetApiKey() string {
	LoadEnvVars()
	getkey := os.Getenv("GEOLOCATION_KEY")
	apikey := strings.TrimSuffix(getkey, "\n")
	return apikey
	// msg := "Got it.  You live in {{.City}}, {{.State}}."
}

var apikey = GetApiKey()

func MakeGeolocationCall(url string) string {
	subin := SubValues{cityname, statename, apikey}
	tmpl, err := template.New("url").Parse(url)
	// Create a variable that implements io.Writer so that I don't have to write the output to standard output
	var parsedGeoUrl bytes.Buffer
	err = tmpl.Execute(&parsedGeoUrl, subin)

	if err != nil {
		fmt.Println(err)
	}

	stringifiedParsedGeourl := fmt.Sprint(&parsedGeoUrl)
	fmt.Println("Parsed geourl: ")
	fmt.Println(&parsedGeoUrl)

	latLongClient := http.Client{
		Timeout: time.Second * 2,
	}

	req, err := http.NewRequest(http.MethodPost, stringifiedParsedGeourl, nil)
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
	return fmt.Sprint(output)
}

func main() {

	preparsedGeourl := "https://maps.googleapis.com/maps/api/geocode/json?address={{.Cityname}},+{{.Statename}}&key={{.Apikey}}"
	preparsedWeatherUrl := "https://api.darksky.net/forecast/{{.DarkskyKey}}/{{.Latitude}},{{.Longitude}}"
	var GeoUrlResponseBody = MakeGeolocationCall(preparsedGeourl)

	type Latlong struct {
		DarkskyKey string
		Latitude   string
		Longitude  string
	}

	getWeatherKey := os.Getenv("DARKSKY_KEY")
	DarkskyKey := strings.TrimSuffix(getWeatherKey, "\n")

	parsedJson, err := gabs.ParseJSON([]byte(GeoUrlResponseBody))

	latitude := parsedJson.Path("results.geometry.location.lat").Data()
	longitude := parsedJson.Path("results.geometry.location.lng").Data()
	// Convert latitude and longitude to strings so they can be interpolated into weatherurl
	// as part of the Latlong struct
	stringifiedLatitude := fmt.Sprint(latitude)
	latWithoutLeftBracket := strings.TrimPrefix(stringifiedLatitude, "[")
	latWithoutBrackets := strings.TrimSuffix(latWithoutLeftBracket, "]")
	stringifiedLongitude := fmt.Sprint(longitude)
	longWithoutLeftBracket := strings.TrimPrefix(stringifiedLongitude, "[")
	longWithoutBrackets := strings.TrimSuffix(longWithoutLeftBracket, "]")

	substitute := Latlong{DarkskyKey, latWithoutBrackets, longWithoutBrackets}
	tmpl2, err2 := template.New("preparsedWeatherUrl").Parse(preparsedWeatherUrl)
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

	minutecast := parsedweatherJson.Path("minutely.summary").Data()

	fmt.Println("Minutecast: ", minutecast)

}
