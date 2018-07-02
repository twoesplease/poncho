package main

import (
	"./checkfinished"
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/Jeffail/gabs"
	"github.com/joho/godotenv"
	. "github.com/logrusorgru/aurora"
	"github.com/matryer/try"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"
)

func Greeting() {
	fmt.Println(Cyan("\nHey there! 👋 Let's get you some weather data."))
}

func GetCity() string {
	checkinput := bufio.NewReader(os.Stdin)
	fmt.Println(Cyan("What's the name of the city you live in?"))
	cityname, _ := checkinput.ReadString('\n')
	cityname = strings.TrimSuffix(cityname, "\n")
	return cityname
}

func GetState() string {
	checkinput := bufio.NewReader(os.Stdin)
	fmt.Println(Cyan("\nAnd what's the 2-letter all-caps abbreviation for the state?"))
	stateabbrev, _ := checkinput.ReadString('\n')
	stateabbrev = strings.TrimSuffix(stateabbrev, "\n")
	return stateabbrev
}

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

var GeoURLResponseBody []byte

func MakeGeolocationCall(url string) ([]byte, error) {
	subin := GeolocationSubValues{GetCity(), GetState(), geoApiKey}
	tmpl, err := template.New("url").Parse(url)
	// Create a variable that implements io.Writer so that I don't have to write the output to standard output
	var parsedGeoUrl bytes.Buffer
	err = tmpl.Execute(&parsedGeoUrl, subin)

	if err != nil {
		log.Fatal(err)
	}

	stringifiedParsedGeourl := fmt.Sprint(&parsedGeoUrl)

	latLongClient := http.Client{
		Timeout: time.Second * 2,
	}

	req, reqErr := http.NewRequest(http.MethodPost, stringifiedParsedGeourl, nil)
	if reqErr != nil {
		log.Fatal(reqErr)
	}

	req.Header.Set("User-Agent", "hobby-weather-app")

	Res, getErr := latLongClient.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
	}

	Output, readErr := ioutil.ReadAll(Res.Body)
	stringifiedBody := string(Output)
	if strings.Contains(stringifiedBody, "ZERO_RESULTS") {
		return nil, errors.New("Hmm, it seems that's not a valid location.Let's try again.")
	}

	if readErr != nil {
		log.Fatal(readErr)
	}
	GeoURLResponseBody = Output
	return Output, nil
}

var value []byte

func geoCallRepeater() {
	err := try.Do(func(attempt int) (bool, error) {
		var err error
		value, err = MakeGeolocationCall(preparsedGeoUrl)
		if err != nil {
			fmt.Println(Red("Hmm, it seems that's not a valid location.  Let's try again."))
			time.Sleep(2 * time.Second) // wait 2 seconds before retrying
		}
		return attempt < 3, err // try 3 times
	})
	if err != nil {
		log.Fatalln("error: ", err)
	}
}

var preparsedGeoUrl = "https://maps.googleapis.com/maps/api/geocode/json?address={{.Cityname}},+{{.Statename}}&key={{.Apikey}}"
var preparsedWeatherUrl = "https://api.darksky.net/forecast/{{.DarkskyKey}}/{{.Latitude}},{{.Longitude}}"

type GeolocationSubValues struct {
	Cityname  string
	Statename string
	Apikey    string
}

func GetLatitude() (latitude string) {
	geoCallRepeater()
	parsedJson, err := gabs.ParseJSON([]byte(GeoURLResponseBody))
	if err != nil {
	}
	lat := parsedJson.Path("results.geometry.location.lat").Data()
	// Convert latitude and longitude to strings so they can be interpolated into weatherurl
	// as part of the Latlong struct
	stringifiedLatitude := fmt.Sprint(lat)
	latWithoutLeftBracket := strings.TrimPrefix(stringifiedLatitude, "[")
	latWithoutBrackets := strings.TrimSuffix(latWithoutLeftBracket, "]")
	return latWithoutBrackets
}

func GetLongitude() (longitude string) {
	parsedJson, err := gabs.ParseJSON([]byte(GeoURLResponseBody))
	if err != nil {
	}
	long := parsedJson.Path("results.geometry.location.lng").Data()
	stringifiedLongitude := fmt.Sprint(long)
	longWithoutLeftBracket := strings.TrimPrefix(stringifiedLongitude, "[")
	longWithoutBrackets := strings.TrimSuffix(longWithoutLeftBracket, "]")
	return longWithoutBrackets
}

var userRequest string

func GetUserRequest() string {
	checkinput := bufio.NewReader(os.Stdin)
	fmt.Println(Cyan(`Here are your choices:
	* Text summary of the next hour's weather. --> Enter "minutely"
	* Percent chance of precipitation in the next hour. --> Enter "hprecip"
	* Temperature that it currently feels like. --> Enter "feelslike"
	* Current humidity level. --> Enter "humidity"
	* Current wind speed. --> Enter "windspeed"
	* Current visibility in miles. --> Enter "visibility"
	* Exit without getting weather data. --> Enter "exit"`))
	request, _ := checkinput.ReadString('\n')
	request = strings.TrimSuffix(request, "\n")
	userRequest = request
	return request
}

func GetWeatherApiKey() string {
	LoadEnvVars()
	getWeatherKey := os.Getenv("DARKSKY_KEY")
	darkskyKey := strings.TrimSuffix(getWeatherKey, "\n")
	return darkskyKey
}

var weatherApiKey = GetWeatherApiKey()

type WeatherSubValues struct {
	DarkskyKey string
	Latitude   string
	Longitude  string
}

var parsedWeatherJSON *gabs.Container

func MakeWeatherApiCall() {
	substitute := WeatherSubValues{weatherApiKey, GetLatitude(), GetLongitude()}
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
	if weatherres.StatusCode != 200 {
		fmt.Println(Red("Oops, that request didn't work.  Let's try again."))
		retryWeatherCall()
	}

	weatheroutput, errRead := ioutil.ReadAll(weatherres.Body)
	if errRead != nil {
		log.Fatal(errRead)
	}

	JSONoutput, err := gabs.ParseJSON([]byte(weatheroutput))
	parsedWeatherJSON = JSONoutput

	GetUserRequest()
	userRequestSwitch()
	stayOrExitSwitch()
}

func userRequestSwitch() {
	minutecast := parsedWeatherJSON.Path("minutely.summary").Data()

	rawHPrecipData := parsedWeatherJSON.Path("hourly.data.0.precipProbability").Data()
	if rawHPrecipData == nil {
		rawHPrecipData = 0
	}
	precipInNextHour := rawHPrecipData.(int) * 100

	feelsLike := parsedWeatherJSON.Path("currently.apparentTemperature").Data()
	feelsLike = fmt.Sprint(feelsLike)

	rawHumidityData := parsedWeatherJSON.Path("currently.humidity").Data()
	if rawHumidityData == nil {
		rawHumidityData = 0
	}
	humidityPercent := rawHumidityData.(float64) * 100

	windSpeed := parsedWeatherJSON.Path("currently.windSpeed").Data()

	visibility := parsedWeatherJSON.Path("currently.visibility").Data()

	switch userRequest {
	case "minutely":
		if minutecast != nil {
			fmt.Println(Green("\nMinutecast: "), Green(minutecast))
		} else {
			fmt.Println(Red("Darn!  I couldn't get that data."))
		}
	case "hprecip":
		fmt.Println(Green("\nChance of precipitation in next hour: "), Green(precipInNextHour), Green("%."))
	case "feelslike":
		fmt.Println(Green("\nIt feels like it's"), Green(feelsLike), Green("°F."))
	case "humidity":
		fmt.Println(Green("\nRight now the humidity's at"), Green(humidityPercent), Green("%."))
	case "windspeed":
		fmt.Println(Green("\nThe wind's blowing"), Green(windSpeed), Green("miles per hour."))
	case "visibility":
		fmt.Println(Green("\nRight now you can see about"), Green(visibility), Green("miles."))
	case "exit":
		fmt.Println(Green("\nOk, bye! 🤙"))
		os.Exit(0)
	default:
		fmt.Println(Red("Sorry, I didn't quite catch that."))
		GetUserRequest()
		userRequestSwitch()
	}

}

func stayOrExitSwitch() {
	var stayOrExit = checkfinished.IsUserDone()

	switch stayOrExit {
	case "more please":
		GetUserRequest()
		userRequestSwitch()
		stayOrExitSwitch()
	case "exit":
		fmt.Println(Cyan("\nOk, bye! 🤙"))
		os.Exit(0)
	default:
		fmt.Println(Red("\nSorry, didn't quite get that.  Can you try again?"))
		stayOrExitSwitch()
	}
}

func retryWeatherCall() {
	//*** Adding GetCity() & GetState() function calls below prevent this from an infinite loop
	// when nil latitude and longitude data are returned ***//
	GetCity()
	GetState()

	MakeWeatherApiCall()
}

func main() {
	Greeting()
	MakeWeatherApiCall()
}
