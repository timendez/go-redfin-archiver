package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var debugModeEnabled = false

func main() {
	configureLogging()
	redfinUrl := os.Args[1]
	htmlText := extractHTML(redfinUrl)
	address := extractAddress(htmlText)
	imageUrlPrefix, imageUrlSuffix := extractImageURLPrefixAndSuffix(htmlText)
	outputDir := createDir(address)
	downloadImages(imageUrlPrefix, imageUrlSuffix, outputDir)
}

func configureLogging() {
	// Third arg is debug mode
	debugModeEnabled = len(os.Args) > 2

	// Open a log file for writing. Create the file if it doesn't exist.
	file, err := os.OpenFile("go-redfin-archiver.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}

	// Create a MultiWriter that writes logs to both the file and the standard output.
	multiWriter := io.MultiWriter(file, os.Stdout)

	// Set the log output to the MultiWriter
	log.SetOutput(multiWriter)
}

func extractAddress(htmlText string) string {
	addressPatternWithTitle := `<title>([0-9A-Za-z\s#.]+),`
	regex := regexp.MustCompile(addressPatternWithTitle)
	matches := regex.FindStringSubmatch(htmlText)
	if len(matches) <= 0 {
		log.Fatal("Can't find address in HTML, please open a github issue https://github.com/timendez/go-redfin-archiver/issues/new and provide the Redfin URL.")
	}

	if debugModeEnabled {
		log.Println("matches[0] = " + matches[0])
	}

	// Strip <title> tag from address match
	addressWithoutTitleTag := strings.Replace(matches[0], "<title>", "", 1)
	if debugModeEnabled {
		log.Println("addressWithoutTitleTag = " + addressWithoutTitleTag)
	}

	// Drop trailing comma
	return addressWithoutTitleTag[:len(addressWithoutTitleTag)-1]
}

func extractImageURLPrefixAndSuffix(htmlText string) (string, string) {
	bigPhotoCDNPrefixPattern := `https://ssl\.cdn-redfin\.com/photo/\d+/bigphoto/\w+/[0-9A-Za-z]+`
	bigPhotoCDNSuffixPattern := `_\d.jpg` // Can't guarantee only 1 number, so best to be safe and use a regex rather than indexing on a fully matched CDN URL.
	prefixRegex := regexp.MustCompile(bigPhotoCDNPrefixPattern)
	suffixRegex := regexp.MustCompile(bigPhotoCDNSuffixPattern)
	prefixMatches := prefixRegex.FindStringSubmatch(htmlText) // The first match in the HTML here is our main image's prefix, minus `_0.jpg` or `_1.jpg` on the end of the string
	suffixMatches := suffixRegex.FindStringSubmatch(htmlText) // The first match in the HTML here is our main image's suffix, like `_0.jpg` or `_1.jpg` from the end of the string
	if len(prefixMatches) <= 0 {
		log.Fatal("Can't find big photo prefix CDN pattern in HTML, please open a github issue https://github.com/timendez/go-redfin-archiver/issues/new and provide the Redfin URL.")
	}
	if len(suffixMatches) <= 0 {
		log.Fatal("Can't find big photo suffix CDN pattern in HTML, please open a github issue https://github.com/timendez/go-redfin-archiver/issues/new and provide the Redfin URL.")
	}

	return prefixMatches[0], suffixMatches[0]
}

func createDir(redfinUrl string) string {
	outputDir := fmt.Sprintf("archives/%s", redfinUrl)
	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		log.Fatal(err)
	}
	if debugModeEnabled {
		log.Println("Directory created for " + outputDir)
	}
	return outputDir
}

func extractHTML(redfinUrl string) string {
	if debugModeEnabled {
		println("hello")
		log.Printf("Redfin URL = %s\n", redfinUrl)
		println("RedfinUrl = " + redfinUrl)
	}

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

	htmlString := string(htmlText)
	if debugModeEnabled {
		log.Println(htmlString)
	}

	return htmlString
}

/**
 * Redfin likes to sometimes go from https://ssl.cdn-redfin.com/photo/69/bigphoto/499/OC22362195_2_1.jpg to
 * https://ssl.cdn-redfin.com/photo/69/bigphoto/499/OC22362195_3_2.jpg
 * Note that the 2nd to last digit keeps incrementing as well.
 * I don't know why.
 * @param imageUrlSuffix looks like `_1.jpg`
 */
func incrementImageSuffix(imageUrlSuffix string) string {
	regex := regexp.MustCompile(`\d+`) // Just yoink the number
	matches := regex.FindStringSubmatch(imageUrlSuffix)
	if len(matches) <= 0 {
		log.Println("Can't increment imageUrlSuffix of " + imageUrlSuffix)
	}

	imageSuffixNumber, err := strconv.Atoi(matches[0])
	if err != nil {
		log.Println(err)
	}

	// Increment and restringify to `_2.jpg`
	return fmt.Sprintf("_%d.jpg", imageSuffixNumber+1)
}

func downloadImages(imageUrlPrefix string, imageUrlSuffix string, outputDir string) {
	idx := 0
	attemptedToIncrementImageSuffix := false
	attemptedToIncrementMiddleNumber := false

	for {
		indexedString := func() string {
			if idx == 0 {
				return ""
			}
			return fmt.Sprintf("_%d", idx)
		}()
		url := fmt.Sprintf("%s%s%s", imageUrlPrefix, indexedString, imageUrlSuffix)
		if debugModeEnabled {
			log.Println("Downloading image of URL " + url)
		}

		fileName := fmt.Sprintf("%s/image%d.jpg", outputDir, idx)
		err := downloadFile(url, fileName)
		if err != nil {
			if !attemptedToIncrementImageSuffix {
				imageUrlSuffix = incrementImageSuffix(imageUrlSuffix) // Overwrite imageUrlSuffix
				attemptedToIncrementImageSuffix = true
			} else if !attemptedToIncrementMiddleNumber {
				idx++
				attemptedToIncrementMiddleNumber = true
			} else {
				// We've counted up in both directions and still failed to find images, probably exhausted them all.
				break
			}
		} else {
			// Reset the dang attempts, we were successful
			attemptedToIncrementImageSuffix = false
			attemptedToIncrementMiddleNumber = false
			idx++
		}
	}

	if debugModeEnabled {
		log.Println("Successfully downloaded all images 🥳")
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
