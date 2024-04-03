package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

// the errors have [x] numbers to test the error propagation visually
// admittedly they are not 'user friendly' (at this stage)

func readResponseBody(response *http.Response) (string, error) {
	responseBody, err := io.ReadAll(response.Body)

	if err != nil {
		return "", fmt.Errorf("[2] failed to read response body : %w", err)
	}

	return string(responseBody), nil
}

// add a second argument of time to allow testing
func parseRetryAfterHeader(retryAfterHeaderString string, currentTime time.Time) (time.Duration, error) {
	// Q1: what are the pros/cons of not directly err != nil (?) on these conversion/parsers
	// Q2: How can we be certain that some standard library methods do/don't "generally" panic(?)

	retryAfterInteger, err := strconv.Atoi(retryAfterHeaderString)

	if err == nil {
		return time.Duration(retryAfterInteger) * time.Second, nil
	}

	retryAfterTime, err := http.ParseTime(retryAfterHeaderString)
	if err == nil {
		return retryAfterTime.Sub(currentTime), nil
	}

	return 0, fmt.Errorf("[3] retry-after header is invalid: %s", retryAfterHeaderString)
}

func handleTooManyRequestsResponse(response *http.Response) (time.Duration, error) {
	retryDuration, err := parseRetryAfterHeader(response.Header.Get("Retry-After"), time.Now())

	if err != nil {
		return 0, fmt.Errorf("[2] failed to get a valid delay from retry-after header: %w", err)
	}

	switch {
	case retryDuration > 5*time.Second:
		return 0, fmt.Errorf("[2] retry-after duration is more than 5 seconds")
	case retryDuration > 0:
		return retryDuration, nil
	default:
		return 0, fmt.Errorf("[2] retry-after duration not found")
	}
}

func handleStatusCode(response *http.Response) (string, time.Duration, error) {
	switch response.StatusCode {
	case http.StatusOK:
		responseBody, err := readResponseBody(response)

		if err != nil {
			return "", 0, fmt.Errorf("[1] readResponseBody failed: %w", err)
		}

		return responseBody, 0, nil
	case http.StatusTooManyRequests:
		retryDuration, err := handleTooManyRequestsResponse(response)

		if err != nil {
			return "", 0, fmt.Errorf("[1] handleTooManyRequestsResponse failed: %w", err)
		}

		return "", retryDuration, nil
	default:
		return "", 0, fmt.Errorf("[1] unhandled response status code: %d %v", response.StatusCode, response.Body)
	}
}

func makeGetRequest(url string) (string, time.Duration, error) {
	response, err := http.Get(url)

	if err != nil {
		return "", 0, fmt.Errorf("[0] request failed: %w", err)
	}

	defer response.Body.Close()

	responseBody, retryDuration, err := handleStatusCode(response)

	if err != nil {
		return "", 0, fmt.Errorf("[0] handleStatusCode failed: %w", err)
	}

	if retryDuration > 0 {
		return "", retryDuration, nil
	}

	return responseBody, 0, nil

	// an alternative to the above 4 clauses (but didn't know this was possible):
	// return handleStatusCode(response)
}

const maxRetries = 3

func main() {
	retryCounter := 0

	for {
		response, retryDuration, err := makeGetRequest("http://localhost:8080")

		if err != nil {
			// option1: show all the errors propagated
			fmt.Fprintln(os.Stderr, err)
			// option2: ignore error propagation and just provide a basic message
			// fmt.Fprintln(os.Stderr, "Giving up, we can't get you the weather.")
			os.Exit(2)
		}

		if retryDuration > 0 {
			if retryCounter >= maxRetries {
				fmt.Fprintf(os.Stderr, "We made a maximum of %d retries, Giving up, we can't get you the weather.", maxRetries)
				os.Exit(1)
			}

			retryCounter++

			fmt.Fprintf(os.Stderr, "Retry Attempt %d/%d | Will retry after %v...\n", retryCounter, maxRetries, retryDuration)
			time.Sleep(time.Duration(retryDuration))
			continue
		}

		fmt.Fprintln(os.Stdout, response)
		break
	}
}
