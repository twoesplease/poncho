package main

import (
	"./checkfinished"
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

type GeolocationSubValues struct {
	Cityname  string
	Statename string
	Apikey    string
}

func GetCity() string {
	checkinput := bufio.NewReader(os.Stdin)
	fmt.Println("\nHey there! 👋 Let's get you some weather data.\n\nWhat's the name of the city you live in?")
	cityname, _ := checkinput.ReadString('\n')
	cityname = strings.TrimSuffix(cityname, "\n")
	return cityname
}

var cityname = GetCity()

func GetState() string {
	checkinput := bufio.NewReader(os.Stdin)
	fmt.Println("\nAnd what's the 2-letter all-caps abbreviation for the state?")
	stateabbrev, _ := checkinput.ReadString('\n')
	stateabbrev = strings.TrimSuffix(stateabbrev, "\n")
	fmt.Println("\nGot it.  Now, what kind of weather data would you like?")
	return stateabbrev
}

var statename = GetState()

func GetUserRequest() string {
	checkinput := bufio.NewReader(os.Stdin)
	fmt.Println(`Here are your choices:
	* Text summary of the next hour's weather. --> Enter "minutely"
	* Percent chance of precipitation in the next hour. --> Enter "hprecip"
	* Temperature that it currently feels like. --> Enter "feelslike"
	* Exit without getting weather data. --> Enter "exit"`)
	userRequest, _ := checkinput.ReadString('\n')
	userRequest = strings.TrimSuffix(userRequest, "\n")
	return userRequest
}

var userRequest = GetUserRequest()

func LoadEnvVars() {
	err := godotenv.Load(".gitignore/key.txt")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func GetGeoApiKey() string {
	LoadEnvVars()
	getkey := os.Getenv("GEOLOCATION_KEY")
	apikey := strings.TrimSuffix(getkey, "\n")
	return apikey
}

var geoApiKey = GetGeoApiKey()

func MakeGeolocationCall(url string) []byte {
	subin := GeolocationSubValues{cityname, statename, geoApiKey}
	tmpl, err := template.New("url").Parse(url)
	// Create a variable that implements io.Writer so that I don't have to write the output to standard output
	var parsedGeoUrl bytes.Buffer
	err = tmpl.Execute(&parsedGeoUrl, subin)

	if err != nil {
		fmt.Println(err)
	}

	stringifiedParsedGeourl := fmt.Sprint(&parsedGeoUrl)

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
	return output
}

var preparsedGeourl = "https://maps.googleapis.com/maps/api/geocode/json?address={{.Cityname}},+{{.Statename}}&key={{.Apikey}}"
var GeoUrlResponseBody = MakeGeolocationCall(preparsedGeourl)
var preparsedWeatherUrl = "https://api.darksky.net/forecast/{{.DarkskyKey}}/{{.Latitude}},{{.Longitude}}"

func GetLatitude() string {
	parsedJson, err := gabs.ParseJSON([]byte(GeoUrlResponseBody))
	if err != nil {
		fmt.Println(err)
	}
	latitude := parsedJson.Path("results.geometry.location.lat").Data()
	// Convert latitude and longitude to strings so they can be interpolated into weatherurl
	// as part of the Latlong struct
	stringifiedLatitude := fmt.Sprint(latitude)
	latWithoutLeftBracket := strings.TrimPrefix(stringifiedLatitude, "[")
	latWithoutBrackets := strings.TrimSuffix(latWithoutLeftBracket, "]")
	return latWithoutBrackets
}

var latitude = GetLatitude()

func GetLongitude() string {
	parsedJson, err := gabs.ParseJSON([]byte(GeoUrlResponseBody))
	if err != nil {
		fmt.Println(err)
	}
	longitude := parsedJson.Path("results.geometry.location.lng").Data()
	stringifiedLongitude := fmt.Sprint(longitude)
	longWithoutLeftBracket := strings.TrimPrefix(stringifiedLongitude, "[")
	longWithoutBrackets := strings.TrimSuffix(longWithoutLeftBracket, "]")
	return longWithoutBrackets
}

var longitude = GetLongitude()

type WeatherSubValues struct {
	DarkskyKey string
	Latitude   string
	Longitude  string
}

func GetWeatherApiKey() string {
	LoadEnvVars()
	getWeatherKey := os.Getenv("DARKSKY_KEY")
	darkskyKey := strings.TrimSuffix(getWeatherKey, "\n")
	return darkskyKey
}

var weatherApiKey = GetWeatherApiKey()

func MakeWeatherApiCall() {
	substitute := WeatherSubValues{weatherApiKey, latitude, longitude}
	tmpl2, err2 := template.New("preparsedWeatherUrl").Parse(preparsedWeatherUrl)
	// Create a variable that implements io.Writer so that I don't have to write the output to standard output
	var parsed_weatherurl bytes.Buffer
	err2 = tmpl2.Execute(&parsed_weatherurl, substitute)

	if err2 != nil {
		fmt.Println(err2)
	}

	stringifiedParsedWeatherUrl := fmt.Sprint(&parsed_weatherurl)

	weatherClient := http.Client{
		Timeout: time.Second * 2,
	}

	weatherreq, err := http.NewRequest(http.MethodGet, stringifiedParsedWeatherUrl, nil)
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

	parsedWeatherJson, err := gabs.ParseJSON([]byte(weatheroutput))

	minutecast := parsedWeatherJson.Path("minutely.summary").Data()

	rawHPrecipData := parsedWeatherJson.Path("hourly.data.0.precipProbability").Data()
	if rawHPrecipData == nil {
		rawHPrecipData = 0
	}
	precipInNextHour := rawHPrecipData.(int) * 100

	feelsLike := parsedWeatherJson.Path("currently.apparentTemperature").Data()
	feelsLike = fmt.Sprint(feelsLike)

	switch userRequest {
	case "minutely":
		fmt.Println("\nMinutecast: ", minutecast)
	case "hprecip":
		fmt.Println("\nChance of precipitation in next hour: ", precipInNextHour, "%")
	case "feelslike":
		fmt.Println("\nIt feels like it's", feelsLike, "°F")
	case "exit":
		fmt.Println("\nOk, bye! 🤙")
		os.Exit(0)
	default:
		fmt.Println("Sorry, I didn't quite catch that.")
		GetUserRequest()
	}

	var stayOrExit = checkfinished.IsUserDone()

	switch stayOrExit {
	case "more please":
		GetUserRequest()
	case "exit":
		fmt.Println("\nOk, bye! 🤙")
		os.Exit(0)
	default:
		fmt.Println("\nSorry, didn't quite get that.  Can you try again?")
		checkfinished.IsUserDone()
	}

}

func main() {
	MakeWeatherApiCall()
}
