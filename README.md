# Go Redfin Archiver
Tool to download a Redfin listing. Currently downloads images into a directory `./`
## Prerequisites
1. [Go 1.21.4+](https://go.dev/doc/install)

## Usage
You can either follow the Quick Run section to compile and run in the same step, or Build & Run to build the executable first. 
### Quick Run
From this directory:
1. `go run https://www.redfin.com/CA/Morgan-Hill/10868-Dougherty-Ave-95037/home/807367`

### Build & Run
1. `go build`
2. `./go-redfin-archiver.exe https://www.redfin.com/CA/Morgan-Hill/10868-Dougherty-Ave-95037/home/807367`

## Demo
![Demo download](./demo.gif)
