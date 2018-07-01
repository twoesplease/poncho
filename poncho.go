package main

import (
	"./checkfinished"
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/Jeffail/gabs"
	"github.com/joho/godotenv"
	"github.com/matryer/try"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"
)

// func Greeting() {
// fmt.Println("\nHey there! 👋 Let's get you some weather data.")
// }

func GetCity() string {
	checkinput := bufio.NewReader(os.Stdin)
	fmt.Println("\nHey there! 👋 Let's get you some weather data.\nWhat's the name of the city you live in?")
	cityname, _ := checkinput.ReadString('\n')
	cityname = strings.TrimSuffix(cityname, "\n")
	fmt.Println("City name: ", cityname)
	return cityname
}

// var cityname = GetCity()

func GetState() string {
	checkinput := bufio.NewReader(os.Stdin)
	fmt.Println("\nAnd what's the 2-letter all-caps abbreviation for the state?")
	stateabbrev, _ := checkinput.ReadString('\n')
	stateabbrev = strings.TrimSuffix(stateabbrev, "\n")
	fmt.Println("State name: ", stateabbrev)
	return stateabbrev
}

// var statename = GetState()

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
		// fmt.Println("Hmm, it seems that's not a valid location.  Let's try again.")
		return nil, errors.New("Hmm, it seems that's not a valid location.Let's try again.")
	}

	if readErr != nil {
		log.Fatal(readErr)
		fmt.Println("Geo Response code: ", Res.StatusCode)
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
			fmt.Println("Hmm, it seems that's not a valid location.  Let's try again.")
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
		fmt.Print("Getlat json parse error: ")
		fmt.Println(err)
	}
	lat := parsedJson.Path("results.geometry.location.lat").Data()
	// Convert latitude and longitude to strings so they can be interpolated into weatherurl
	// as part of the Latlong struct
	stringifiedLatitude := fmt.Sprint(lat)
	latWithoutLeftBracket := strings.TrimPrefix(stringifiedLatitude, "[")
	latWithoutBrackets := strings.TrimSuffix(latWithoutLeftBracket, "]")
	fmt.Println("Latitude: ", latWithoutBrackets)
	return latWithoutBrackets
}

// var latitude = GetLatitude()

func GetLongitude() (longitude string) {
	parsedJson, err := gabs.ParseJSON([]byte(GeoURLResponseBody))
	if err != nil {
		fmt.Print("Getlong JSON parse error: ")
		fmt.Println(err)
	}
	long := parsedJson.Path("results.geometry.location.lng").Data()
	stringifiedLongitude := fmt.Sprint(long)
	longWithoutLeftBracket := strings.TrimPrefix(stringifiedLongitude, "[")
	longWithoutBrackets := strings.TrimSuffix(longWithoutLeftBracket, "]")
	fmt.Println("Longitude: ", longWithoutBrackets)
	return longWithoutBrackets
}

// var longitude = GetLongitude()

func introduceWeatherRequest() {
	fmt.Println("\nGot it.  Now, what kind of weather data would you like?")
}

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
	fmt.Println("Parsed weather URL: ", stringifiedParsedWeatherUrl)

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
	fmt.Println("Status code: ", weatherres.StatusCode)
	if weatherres.StatusCode != 200 {
		fmt.Println("Oops, that request didn't work.  Let's try again.")
		// cityname = ""
		// statename = ""
		// latitude = ""
		// longitude = ""
		// parsed_weatherurl.Reset()
		// stringifiedParsedWeatherUrl = ""
		// tmpl2 = {{""}}
		// err2 = nil
		// weatherreq = nil
		// weatherres = nil
		retryWeatherCall()
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
		if minutecast != nil {
			fmt.Println("\nMinutecast: ", minutecast)
		} else {
			fmt.Println("Darn!  I couldn't get that data.")
		}
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
		MakeWeatherApiCall()
	case "exit":
		fmt.Println("\nOk, bye! 🤙")
		os.Exit(0)
	default:
		fmt.Println("\nSorry, didn't quite get that.  Can you try again?")
		checkfinished.IsUserDone()
	}

}

func retryWeatherCall() {
	// cityname = ""
	// statename = ""
	// GeoURLResponseBody = nil
	// latitude = ""
	// longitude = ""
	// userRequest = ""
	// Greeting()

	//*** Adding GetCity() & GetState() function calls below prevent this from an infinite loop
	// when nil latitude and longitude data are returned ***//
	GetCity()
	GetState()

	MakeWeatherApiCall()
}

func main() {
	// Greeting()
	// GetCity()
	// GetState()
	// MakeGeolocationCall(preparsedGeoUrl)
	MakeWeatherApiCall()
}
