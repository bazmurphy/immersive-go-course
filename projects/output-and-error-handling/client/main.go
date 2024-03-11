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

func parseResponseBody(response *http.Response) (string, error) {
	// try to read the response body
	responseBody, err := io.ReadAll(response.Body)

	// handle any error immediately
	if err != nil {
		return "", fmt.Errorf("[2] failed to read response body : %w", err)
	}

	// use the defer keyword early to close the body regardless
	defer response.Body.Close()

	return string(responseBody), nil
}

// returned integer is the retry duration in seconds
func parseRetryAfterHeader(retryAfterHeaderString string) (int, error) {

	// first condition "a while"
	if retryAfterHeaderString == "a while" {
		return 0, fmt.Errorf("[3] retry-after header is 'a while': %s", retryAfterHeaderString)
	}

	// second condition an integer in string format
	retryAfterInteger, err := strconv.Atoi(retryAfterHeaderString)

	// should immediately error check here...
	// but i need to be able to continue after...
	// in order to check if it's a timestamp...?

	// if err != nil {
	//  return 0, nil
	// }

	// if we successfully parse an integer return it
	if err == nil {
		return retryAfterInteger, nil
	}

	// third condition a timestamp in string format
	retryAfterTimeStamp, err := time.Parse(time.RFC1123, retryAfterHeaderString)

	// if we cannot parse the timestamp
	if err != nil {
		return 0, fmt.Errorf("[3] retry-after header is invalid: %w", err)
	}

	// otherwise calculate the difference and return it in seconds (rounded up)
	now := time.Now()
	difference := retryAfterTimeStamp.Sub(now)
	timestampDifferenceInSeconds := int(math.Ceil(difference.Seconds()))

	return timestampDifferenceInSeconds, nil
}

// the return int is the delay in seconds (using an int here is not great... need to think of something better)
func handleTooManyRequestsResponse(response *http.Response) (int, error) {
	// get the Retry-After Header string
	retryDuration, err := parseRetryAfterHeader(response.Header.Get("Retry-After"))

	// if we cannot parse the Retry-After Header string
	if err != nil {
		return 0, fmt.Errorf("[2] failed to get a valid delay from retry-after header")
	}

	// if we reach here, we have an retry delay integer
	switch {
	case retryDuration > 5:
		return 0, fmt.Errorf("[2] retry-after is more than 5 seconds: %d", retryDuration)
	case retryDuration > 1:
		return retryDuration, nil
	case retryDuration == 1:
		return 1, nil
	default:
		return 0, fmt.Errorf("[2] retry-after condition not found")
	}
}

// this is returning a responseBody, a retry duration in seconds, and an error
// this breaks single responsibility... need to rethink it
func handleStatusCode(response *http.Response) (string, int, error) {
	// handle the various response status codes (extensible for more status codes)
	switch response.StatusCode {
	// 200
	case http.StatusOK:
		// parse the response body
		responseBody, err := parseResponseBody(response)

		// if we cannot parse the body return an error
		if err != nil {
			return "", 0, fmt.Errorf("[1] parseResponseBody failed: %w", err)
		}

		// otherwise return the parsed body
		return responseBody, 0, nil

	// 429
	case http.StatusTooManyRequests:
		// this retryDuration is an int (not a string)
		retryDuration, err := handleTooManyRequestsResponse(response)

		// if we cannot get a valid retryDuration
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

func makeGetRequest(url string) (string, int, error) {
	// try to make the get request
	response, err := http.Get(url)

	// if the request fails
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

	// we have a parsed response body
	return responseBody, 0, nil
}

func main() {
	response, retryDuration, err := makeGetRequest("http://localhost:8080")
	// fmt.Printf("response %s\nretryDuration %d\nerr %v\n\n", response, retryDuration, err)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	if retryDuration > 0 {
		fmt.Printf("retry after %d seconds", retryDuration)
		// how to effectively make a retry here?
		// need to probably rewrite the main function
		// to loop until there is a successful response (or an exit)
	}

	fmt.Fprintln(os.Stdout, response)
}
