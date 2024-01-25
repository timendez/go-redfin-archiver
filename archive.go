package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

func main() {
	redfinUrl := os.Args[1]
	htmlText := extractHTML(redfinUrl)
	address := extractAddress(htmlText)
	imageUrlPrefix := extractImageURLPrefix(htmlText)
	outputDir := createDir(address)
	downloadImages(imageUrlPrefix, outputDir)
}

func extractAddress(htmlText string) string {
	addressPatternWithTitle := `<title>([0-9A-Za-z\s#.]+),`
	regex := regexp.MustCompile(addressPatternWithTitle)
	matches := regex.FindStringSubmatch(htmlText)
	if len(matches) <= 0 {
		log.Fatal("Can't find address in HTML, please open a github issue https://github.com/timendez/go-redfin-archiver/issues/new and provide the Redfin URL.")
	}

	// Strip <title> tag from address match
	addressWithoutTitleTag := strings.Replace(matches[0], "<title>", "", 1)

	// Drop trailing comma
	return addressWithoutTitleTag[:len(addressWithoutTitleTag)-1]
}

func extractImageURLPrefix(htmlText string) string {
	bigPhotoCDNPattern := `https://ssl\.cdn-redfin\.com/photo/\d+/bigphoto/\w+/[0-9A-Za-z]+`
	regex := regexp.MustCompile(bigPhotoCDNPattern)
	matches := regex.FindStringSubmatch(htmlText)
	if len(matches) <= 0 {
		log.Fatal("Can't find big photo CDN pattern in HTML, please open a github issue https://github.com/timendez/go-redfin-archiver/issues/new and provide the Redfin URL.")
	}

	// The first match in the HTML here is our main image, minus `_0.jpg` on the end of the string
	return matches[0]
}

func createDir(redfinUrl string) string {
	outputDir := fmt.Sprintf("archives/%s", redfinUrl)
	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		log.Fatal(err)
	}
	return outputDir
}

func extractHTML(redfinUrl string) string {
	// Set up request
	client := &http.Client{}
	req, err := http.NewRequest("GET", redfinUrl, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("TE", "trailers")

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	// Wrap error handling in a closure for closing hte body
	defer func(bodyReadCloser io.ReadCloser) {
		err := bodyReadCloser.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(resp.Body)

	// Read HTML body into a variable
	htmlText, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	return string(htmlText)
}

func downloadImages(imageUrlPrefix string, outputDir string) {
	idx := 0
	for {
		indexedString := func() string {
			if idx == 0 {
				return ""
			}
			return fmt.Sprintf("_%d", idx)
		}()
		// TODO Tim handle older listings like https://www.redfin.com/CA/Morgan-Hill/10868-Dougherty-Ave-95037/home/807367
		url := fmt.Sprintf("%s%s_0.jpg", imageUrlPrefix, indexedString)
		fileName := fmt.Sprintf("%s/image%d.jpg", outputDir, idx)
		err := downloadFile(url, fileName)
		if err != nil {
			break
		}
		idx++
	}
}

// Modified from https://golangbyexample.com/download-image-file-url-golang/
func downloadFile(url string, fileName string) error {
	// Get the resp bytes from the url
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(resp.Body)

	if resp.StatusCode != 200 {
		return errors.New("image not successfully got")
	}

	// Create an empty file
	file, err := os.Create(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(file)

	// Write the bytes to the file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
