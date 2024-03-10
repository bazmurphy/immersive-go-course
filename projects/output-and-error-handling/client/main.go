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

func handleOKResponse(response *http.Response) (string, error) {
	// try to read the response body
	responseBody, err := io.ReadAll(response.Body)

	// handle any error immediately
	if err != nil {
		return "", fmt.Errorf("failed to read response body : %w", err)
	}

	// use the defer keyword early to close the body regardless
	defer response.Body.Close()

	return string(responseBody), nil
}

// returned integer is the time in seconds
func parseRetryAfterHeader(retryAfterHeaderString string) (int, error) {

	// first condition "a while"
	if retryAfterHeaderString == "a while" {
		return 0, fmt.Errorf("retry-after header is 'a while' : %s", retryAfterHeaderString)
	}

	// second condition an integer in string format
	retryAfterInteger, err := strconv.Atoi(retryAfterHeaderString)

	// if err != nil {
	// 	return 0, nil
	// }
	// but i need to be able to continue from here... to check if it's a timestamp...?

	// if we successfully parse an integer return it
	if err == nil {
		return retryAfterInteger, nil
	}

	// third condition a timestamp in string format
	retryAfterTimeStamp, err := time.Parse(time.RFC1123, retryAfterHeaderString)
	// if we cannot parse the timestamp
	if err != nil {
		return 0, fmt.Errorf("retry-after header is invalid : %w", err)
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
	retryAfter, err := parseRetryAfterHeader(response.Header.Get("Retry-After"))

	// if we cannot parse the Retry-After Header string
	if err != nil {
		return 0, fmt.Errorf("failed to get a valid delay from retry-after header")
	}

	// if we reach here, we have an retry delay integer
	switch {
	case retryAfter > 5:
		return 0, fmt.Errorf("retry-after is more than 5 seconds : %d", retryAfter)
	case retryAfter > 1:
		return retryAfter, nil
	case retryAfter == 1:
		return 1, nil
	default:
		return 0, fmt.Errorf("retry-after condition not found")
	}
}

func handleResponse(response *http.Response) (string, error) {
	// handle the various response status codes (extensible for more status codes)
	switch response.StatusCode {
	case http.StatusOK:
		response, err := handleOKResponse(response)
		if err != nil {
			return "", fmt.Errorf("handleOKResponse failed : %w", err)
		}
		return response, nil
	case http.StatusTooManyRequests:
		// this retryDuration is an int (not a string)
		retryDuration, err := handleTooManyRequestsResponse(response)
		if err != nil {
			return "", fmt.Errorf("handleTooManyRequestsResponse failed : %w", err)
		}
		// this is not what i want... I need to send the int through to trigger a retry
		return "", fmt.Errorf("retry after %d seconds", retryDuration)
	default:
		return "", fmt.Errorf("unhandled response status code : %d", response.StatusCode)
	}
}

func makeGetRequest(url string) (string, error) {
	// try to make the get request
	response, err := http.Get(url)

	// if the request fails
	if err != nil {
		return "", fmt.Errorf("request failed : %w", err)
	}

	// try to get handle the response
	responseBody, err := handleResponse(response)
	// ^responseBody is the wrong naming, and it's doing two things...(bad)

	// if the response fails
	if err != nil {
		return "", fmt.Errorf("handleResponse failed : %w", err)
	}

	return responseBody, nil
}

func main() {
	response, err := makeGetRequest("http://localhost:8080")
	if err != nil {
		// if there is an error... (of which there should be a few)
		// then write the logic to handle it here... (but ideally not in main())
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	fmt.Fprintln(os.Stdout, response)
}
