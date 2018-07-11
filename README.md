# Poncho # 

## Overview ##
Poncho is a command line app that asks users for their location data and then allows them to select various kinds of weather data for the app to provide them.

## Requirements ##
Poncho is a command line app written in Go, and in order to use it you'll need to have access to an application that can run a command line interface 
to your computer's operating system and you'll need to have Go installed on your computer. [Here's](https://www.codecademy.com/articles/command-line-interface)
and article from Codecademy with more information about command line interfaces.  If you need help installing go on your computer, 
[this resource](https://golang.org/doc/install) can help out.

### Dependencies ###
Poncho relies on a few dependencies that you'll also need to have installed on your computer to run properly. This project uses the `dep` library to help
with managing those.  To install those dependencies on your machine, [install dep](https://golang.github.io/dep/docs/installation.html) 
and then run the `dep ensure` command inside the directory where the app lives.

## Use ##
Once you've cloned Poncho and run `go get` for its dependencies, you can use the command `go run poncho.go` to start it up.  The app will then tell you what to do.  
Enjoy!
