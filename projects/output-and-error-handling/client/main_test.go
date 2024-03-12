package main

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"
)

// This is the beginning of the test implementation
// It is not comprehensive by any means

func TestParseResponseBody(t *testing.T) {
	t.Run("response with valid body", func(t *testing.T) {
		body := "Today it will be sunny!"

		// io.NopCloser is a non-closing reader
		response := &http.Response{Body: io.NopCloser(bytes.NewBufferString(body))}

		result, err := parseResponseBody(response)

		if err != nil {
			t.Errorf("could not parse body")
		}

		if result != body {
			t.Errorf("got %q | want %q", body, result)
		}
	})

	t.Run("response with invalid body", func(t *testing.T) {
		// this is how to make an http response with no body (without panicking nil deference)
		response := &http.Response{Body: http.NoBody}

		result, err := parseResponseBody(response)

		want := "failed to read response body : EOF"

		if result != "" {
			t.Errorf("got %v | want %v", result, "")
		}

		if err.Error() != want {
			t.Errorf("got %v | want %v", err, want)
		}
	})
}

func TestParseRetryAfterHeader(t *testing.T) {
	t.Run("a while", func(t *testing.T) {
		// pass it a retry header of "a while"
		retryDuration, err := parseRetryAfterHeader("a while")

		// expect back an error
		want := errors.New("[3] retry-after header is 'a while'")

		if retryDuration != 0 || err == nil {
			t.Errorf("expected an error")
		}

		if err.Error() != want.Error() {
			t.Errorf("got %q | want %q", err, want)
		}
	})

	t.Run("string duration '3'", func(t *testing.T) {
		// pass it a retry header of "3"
		retryDuration, err := parseRetryAfterHeader("3")

		// expect back integer 3
		want := 3

		if retryDuration != want || err != nil {
			t.Errorf("got %v and %v | want %v and nil", retryDuration, err, want)
		}
	})

	t.Run("string timestamp 4 seconds in the future'", func(t *testing.T) {
		// get the time
		now := time.Now()
		// add X seconds to it
		futureTime := now.Add(4 * time.Second)
		// create a timestamp from it
		timestamp := futureTime.Format(time.RFC1123)

		retryDuration, err := parseRetryAfterHeader(timestamp)

		// expect back integer 4
		want := 4

		if retryDuration != want || err != nil {
			t.Errorf("got %v and %v | want %v and nil", retryDuration, err, want)
		}
	})
}

func TestHandleTooManyRequestsResponse(t *testing.T) {
	t.Run("retry-after header with string '4'", func(t *testing.T) {
		// create an http response with a header with retry-after of "4"
		response := &http.Response{Header: http.Header{"Retry-After": []string{"4"}}}

		retryDuration, err := handleTooManyRequestsResponse(response)

		// expect back integer 4
		want := 4

		if retryDuration != want || err != nil {
			t.Errorf("got %v and %v | want %v and nil", retryDuration, err, want)
		}
	})

	t.Run("retry-after header with string '10'", func(t *testing.T) {
		// create an http response with a header with retry-after of "10"
		response := &http.Response{Header: http.Header{"Retry-After": []string{"10"}}}

		retryDuration, err := handleTooManyRequestsResponse(response)

		want := errors.New("[2] retry-after is more than 5 seconds")

		// check the retryDuration is 0
		if retryDuration != 0 {
			t.Errorf("got retryDuration %v and want %v ", retryDuration, 0)
		}

		// check the errors match
		if err.Error() != want.Error() {
			t.Errorf("got %v | want %v", err, want)
		}
	})
}
