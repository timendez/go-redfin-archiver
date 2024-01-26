# Go Redfin Archiver
Tool to download a Redfin listing. Currently, downloads listing images into a directory `./archives`.

## Prerequisites
1. [Go 1.21.4+](https://go.dev/doc/install)

## Usage
You can either follow the Quick Run section to compile and run in the same step, or Build & Run to build the executable first.

### Quick Run
From this directory:
1. `go run archive.go https://www.redfin.com/CA/San-Jose/206-Grayson-Ter-95126/home/2122534`

### Build & Run
1. `go build`
2. `./go-redfin-archiver.exe https://www.redfin.com/CA/San-Jose/206-Grayson-Ter-95126/home/2122534`

### With Debug Mode
Debug mode creates a `go-redfin-archiver.log` file, and spits out debug info to `stdout`.
1. `go run archive.go https://www.redfin.com/CA/San-Jose/206-Grayson-Ter-95126/home/2122534 debug`
2. `./go-redfin-archiver.exe https://www.redfin.com/CA/San-Jose/206-Grayson-Ter-95126/home/2122534 debug`

## Demo
![Demo download](./demo.gif)

## Troubleshooting
1. I get `Can't find address in HTML, please open a github issue https://github.com/timendez/go-redfin-archiver/issues/new and provide the Redfin URL.`
   1. Chances are you're getting rate limited due to too much usage. Try using a VPN.
