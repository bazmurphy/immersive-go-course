package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Comments ----------
// I write a lot of comments, it has helped me the entire course to learn, understand and remember
// They can all be removed at the end to clean it up

// Error Messages ----------
// The Errors are deliberately numbered [x] so I could follow them up the call stack
// They can be replaced with "user friendly" messages later

func parseResponseBody(response *http.Response) (string, error) {
	// check if there is a response body
	if response.Body == nil {
		return "", fmt.Errorf("[2] there is no response body")
	}

	// try to read the response body
	responseBody, err := io.ReadAll(response.Body)

	// if we cannot read the response body then return an error
	if err != nil {
		return "", fmt.Errorf("[2] failed to read response body : %w", err)
	}

	// use the defer keyword early to close the body regardless
	defer response.Body.Close()

	return string(responseBody), nil
}

func parseRetryAfterHeader(retryAfterHeaderString string) (time.Duration, error) {
	// fmt.Println("--- DEBUG retryAfterHeaderString", retryAfterHeaderString)

	// first condition "a while"
	if retryAfterHeaderString == "a while" {
		return 0, fmt.Errorf("[3] retry-after header is 'a while'")
	}

	// second condition an integer in string format
	retryAfterInteger, err := strconv.Atoi(retryAfterHeaderString)

	// if we successfully parse an integer
	if err == nil {
		// convert the integer into seconds and return it
		retryDuration := time.Duration(retryAfterInteger) * time.Second
		return retryDuration, nil
	}

	// third condition a timestamp in string format
	// (!) http.ParseTime can parse various formats
	// specified in the RFCs for HTTP (RFC1123, RFC850, ANSI C etc)
	// so it is a better choice than time.Parse(RFC1123, _)
	retryAfterTimeStamp, err := http.ParseTime(retryAfterHeaderString)

	// if we cannot parse the timestamp then return an error
	if err != nil {
		return 0, fmt.Errorf("[3] retry-after header is invalid: %w", err)
	}

	// if we can parse the timestamp then calculate the difference and return it
	retryDuration := time.Until(retryAfterTimeStamp)

	return retryDuration, nil
}

func handleTooManyRequestsResponse(response *http.Response) (time.Duration, error) {
	// parse the Retry-After Header string
	retryDuration, err := parseRetryAfterHeader(response.Header.Get("Retry-After"))

	// if we cannot parse the Retry-After Header string then return an error
	if err != nil {
		return 0, fmt.Errorf("[2] failed to get a valid delay from retry-after header")
	}

	// if we reach here we have a retry duration (of type time.Duration)
	switch {
	case retryDuration > 5*time.Second:
		return 0, fmt.Errorf("[2] retry-after is more than 5 seconds")
	case retryDuration > 0:
		return retryDuration, nil
	default:
		return 0, fmt.Errorf("[2] retry-after condition not found")
	}
}

// this is returning a responseBody, a retry duration in seconds, and an error
// this breaks single responsibility... need to rethink it... suggestions?
func handleStatusCode(response *http.Response) (string, time.Duration, error) {
	// handle the various response status codes (extensible for more status codes)
	switch response.StatusCode {
	// 200
	case http.StatusOK:
		// parse the response body
		responseBody, err := parseResponseBody(response)

		// if we cannot parse the response body then return an error
		if err != nil {
			return "", 0, fmt.Errorf("[1] parseResponseBody failed: %w", err)
		}

		// otherwise return the parsed response body
		return responseBody, 0, nil

	// 429
	case http.StatusTooManyRequests:
		// get the retry duration
		retryDuration, err := handleTooManyRequestsResponse(response)

		// if we cannot get a valid retry duration then return an error
		if err != nil {
			return "", 0, fmt.Errorf("[1] handleTooManyRequestsResponse failed: %w", err)
		}

		// return a retry duration
		return "", retryDuration, nil

	// Other Status Codes
	default:
		return "", 0, fmt.Errorf("[1] unhandled response status code: %d", response.StatusCode)
	}
}

func makeGetRequest(url string) (string, time.Duration, error) {
	// try to make the get request
	response, err := http.Get(url)

	// if the request fails then return an error
	if err != nil {
		return "", 0, fmt.Errorf("[0] request failed: %w", err)
	}

	// send the response to the status code handler
	responseBody, retryDuration, err := handleStatusCode(response)

	// if there is an error
	if err != nil {
		return "", 0, fmt.Errorf("[0] handleStatusCode failed: %w", err)
	}

	// if there is a retry duration
	if retryDuration > 0 {
		return "", retryDuration, nil
	}

	// return the parsed response body
	return responseBody, 0, nil
}

func main() {
	// loop until a successful response or an unrecoverable error
	for {
		response, retryDuration, err := makeGetRequest("http://localhost:8080")

		// if we have an unrecoverable error
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}

		// if we have a retry duration
		if retryDuration > 0 {
			fmt.Fprintf(os.Stderr, "Retrying after %v...\n", retryDuration)
			time.Sleep(time.Duration(retryDuration))
			continue
		}

		// if we have a successful response
		fmt.Fprintln(os.Stdout, response)
		break
	}
}
