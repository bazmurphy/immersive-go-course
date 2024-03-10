package main

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Comments ----------
// I write a lot of comments, it has helped me the entire course to learn, understand and remember
// They can all be removed at the end to clean it up

func handleOKResponse(response *http.Response) {
	responseBody, err := io.ReadAll(response.Body)

	// use the defer keyword early to close the body regardless
	defer response.Body.Close()

	if err != nil {
		// fmt.Fprintln(os.Stderr, "Parsing Response Body Failed. Error:", err)
		// deliberately without explicit error message
		fmt.Fprintln(os.Stderr, "Parsing Response Body Failed.")
		return
	}

	// deliberately using os.Stdout (not fmt.Println) even though unnecessary (for the sake of the exercise)
	fmt.Fprintln(os.Stdout, string(responseBody))
}

// note: returns an integer: 0 for fail || > 0 for success
func parseRetryAfterHeader(retryAfterHeaderString string) int {
	// first condition "a while"
	if retryAfterHeaderString == "a while" {
		return 0
	}

	// second condition an integer in string format
	retryAfterInteger, err := strconv.Atoi(retryAfterHeaderString)
	// if we successfully parse an integer return it
	if err == nil {
		return retryAfterInteger
	}

	// third condition a timestamp in string format
	retryAfterTimeStamp, err := time.Parse(time.RFC1123, retryAfterHeaderString)
	// if we cannot parse the timestamp then exit
	if err != nil {
		fmt.Fprintln(os.Stderr, "Retry-After Header is neither 'a while', an integer or a timestamp")
		return 0
	}

	// otherwise calculate the difference and return it in seconds (rounded up)
	now := time.Now()
	timeDifference := retryAfterTimeStamp.Sub(now)
	timeDifferenceInSeconds := int(math.Ceil(timeDifference.Seconds()))
	return timeDifferenceInSeconds
}

func handleTooManyRequestsResponse(response *http.Response, url string) {
	// get the Retry-After Header string
	retryAfter := parseRetryAfterHeader(response.Header.Get("Retry-After"))

	switch {
	case retryAfter > 5:
		fmt.Fprintln(os.Stderr, "Giving up. We can't get you the Weather.")
	case retryAfter > 1:
		fmt.Fprintf(os.Stderr, "Things may be a bit slow because we're doing a retry after %d seconds...\n", retryAfter)
		time.Sleep(time.Duration(retryAfter))
		makeGetRequest(url)
	case retryAfter <= 1 && retryAfter > 0:
		fmt.Fprintln(os.Stderr, "Things may be a bit slow because we're doing a retry after 1 second...")
		time.Sleep(1 * time.Second)
		makeGetRequest(url)
	case retryAfter == 0:
		fmt.Fprintln(os.Stderr, "Giving up. We can't get you the Weather.")
	}
	// do we actually need a default here...?
}

func handleResponse(response *http.Response, url string) {
	// handle the various response status codes (extensible for more status codes)
	switch response.StatusCode {
	case http.StatusOK:
		handleOKResponse(response)
	case http.StatusTooManyRequests:
		handleTooManyRequestsResponse(response, url)
	}
}

func makeGetRequest(url string) {
	response, err := http.Get(url)

	if err != nil {
		// fmt.Fprintln(os.Stderr, "Request Failed. Error:", err)
		// deliberately without explicit error message
		fmt.Fprintln(os.Stderr, "Request Failed.")
		return
	}
	// pass through the url for the second handler
	handleResponse(response, url)
}

func main() {
	// just for testing purposes, make an infinite loop
	for {
		// make the GET Request
		makeGetRequest("http://localhost:8080")
		// sleep for 1 second
		time.Sleep(1 * time.Second)
	}
}
