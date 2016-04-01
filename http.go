package main

import (
	"encoding/json"
	"github.com/cheggaaa/pb"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

var OFFERHOST string = "https://pricing.us-east-1.amazonaws.com"
var OFFERINDEX string = "/offers/v1.0/aws/index.json"

// GET a URL, return a Json type or an error
func http2json(url string) (j Json, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &j)
	if err != nil {
		return
	}

	return
}

// GET a URL, save it to a file
func http2file(url, filePath string, progressBar bool) (err error) {

	out, err := os.Create(filePath)
	if err != nil {
		return
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if progressBar {
		// Set up the progress bar
		bar := pb.New(int(resp.ContentLength)).SetUnits(pb.U_BYTES)
		bar.Start()

		// Create a progress-bar-compatible reader
		reader := bar.NewProxyReader(resp.Body)
		// Read the body to the file, via the progress bar
		_, err = io.Copy(out, reader)
	} else {
		// No progress bar
		_, err = io.Copy(out, resp.Body)
	}

	return
}
