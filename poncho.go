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
	geourl := "https://maps.googleapis.com/maps/api/geocode/json?address={{.City}},+{{.State}}&key=AIzaSyA2tf3-8Sxj_gu73TBHE46Qwkn9EXZHtaw"

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

	parsedJson, err := gabs.ParseJSON([]byte(output))

	latitude, didlatwork := parsedJson.Path("results.geometry.location.lat").Data().(interface{})
	longitude, didlongwork := parsedJson.Path("results.geometry.location.lng").Data().(interface{})
	fmt.Println("Latitude: ", latitude, didlatwork)
	fmt.Println("Longitude: ", longitude, didlongwork)

}
