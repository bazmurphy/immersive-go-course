package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"
)

// This is the beginning of the test implementation
// It is not comprehensive by any means

func TestParseResponseBody(t *testing.T) {
	t.Run("response with valid body", func(t *testing.T) {
		// expect back a string from the body
		want := "Today it will be sunny!"

		// io.NopCloser is a non-closing reader
		response := &http.Response{Body: io.NopCloser(bytes.NewBufferString(want))}

		responseBody, err := parseResponseBody(response)

		if err != nil {
			t.Errorf("could not parse body")
		}

		if responseBody != want {
			t.Errorf("got responseBody %q | want %q", responseBody, want)
		}
	})

	t.Run("response with no body", func(t *testing.T) {
		// expect back an error
		wantErr := errors.New("[2] there is no response body")

		// this is how to make an http response with no body (without panic nil pointer deference)
		response := &http.Response{Body: nil}

		responseBody, err := parseResponseBody(response)

		// there should be an error
		if err == nil {
			t.Errorf("expected an error")
		}

		// the errors should match
		if err.Error() != wantErr.Error() {
			t.Errorf("got error %v | want %v", err, wantErr)
		}

		// the responseBody should be ""
		if responseBody != "" {
			t.Errorf("got responseBody %v | want %v", responseBody, "")
		}
	})
}

func TestParseRetryAfterHeader(t *testing.T) {
	t.Run("a while", func(t *testing.T) {
		// expect back an error
		wantErr := errors.New("[3] retry-after header is 'a while'")

		// pass it a retry header of "a while"
		retryDuration, err := parseRetryAfterHeader("a while")

		// there should be an error
		if err == nil {
			t.Errorf("expected an error")
		}

		// the error should match
		if err.Error() != wantErr.Error() {
			t.Errorf("got error %q | want %q", err, wantErr)
		}

		// the retry duration should be 0
		if retryDuration != 0 {
			t.Errorf("got retryDuration %v | want %v", retryDuration, 0)
		}
	})

	t.Run("string duration '3'", func(t *testing.T) {
		// expect back 3 seconds (time.Duration)
		want := 3 * time.Second

		// pass it a retry header of "3"
		retryDuration, err := parseRetryAfterHeader("3")

		// there should not be an error
		if err != nil {
			t.Errorf("got error %v | want %v", err, nil)
		}

		// the retry duration should match
		if retryDuration != want {
			t.Errorf("got retryDuration %v | want %v", retryDuration, want)
		}
	})

	t.Run("string timestamp 4 seconds in the future'", func(t *testing.T) {
		// expect back 5 seconds (time.Duration)
		want := 5 * time.Second

		// get the time
		now := time.Now()
		fmt.Println("--- DEBUG now", now)

		// add 5 seconds to it
		futureTime := now.Add(5 * time.Second)
		fmt.Println("--- DEBUG futureTime", futureTime)

		// create a timestamp from it
		timestamp := futureTime.UTC().Format(http.TimeFormat)
		fmt.Println("--- DEBUG timestamp", timestamp)

		retryDuration, err := parseRetryAfterHeader(timestamp)
		fmt.Println("--- DEBUG retryDuration", retryDuration)

		// there should not be an error
		if err != nil {
			t.Errorf("got error %v | want %v", err, nil)
		}

		// the retry duration should match
		if retryDuration != want {
			t.Errorf("got retryDuration %v | want %v", retryDuration, want)
			// FAIL
			// got retryDuration 4.622255471s | want 5s
			// presumably because by the time the conversion and comparison happens time is lost (?)
		}
	})
}

func TestHandleTooManyRequestsResponse(t *testing.T) {
	t.Run("retry-after header with string '4'", func(t *testing.T) {
		// create an http response with a header with retry-after of "4"
		response := &http.Response{Header: http.Header{"Retry-After": []string{"4"}}}

		retryDuration, err := handleTooManyRequestsResponse(response)

		// expect back 4 seconds (time.Duration)
		want := 4 * time.Second

		// there should not be an error
		if err != nil {
			t.Errorf("got err %v | want %v", err, nil)
		}

		// the retry duration should match
		if retryDuration != want {
			t.Errorf("got retryDuration %v | want %v", retryDuration, want)
		}
	})

	t.Run("retry-after header with string '10'", func(t *testing.T) {
		// create an http response with a header with retry-after of "10"
		response := &http.Response{Header: http.Header{"Retry-After": []string{"10"}}}

		retryDuration, err := handleTooManyRequestsResponse(response)

		// expect back an error
		want := errors.New("[2] retry-after is more than 5 seconds")

		// the retry duration should be 0
		if retryDuration != 0 {
			t.Errorf("got retryDuration %v and want %v ", retryDuration, 0)
		}

		// check the errors match
		if err.Error() != want.Error() {
			t.Errorf("got %v | want %v", err, want)
		}
	})
}
