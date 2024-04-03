package main

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestParseResponseBody(t *testing.T) {
	type testCase struct {
		name                 string
		responseBody         string
		expectedResponseBody string
	}

	testCases := []testCase{
		{
			name:                 "valid sunny response",
			responseBody:         "Today it will be sunny!",
			expectedResponseBody: "Today it will be sunny!",
		}, {
			name:                 "rainy response",
			responseBody:         "I'd bring an umbrella, just in case...",
			expectedResponseBody: "I'd bring an umbrella, just in case...",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			response := &http.Response{Body: io.NopCloser(bytes.NewBufferString(tc.responseBody))}

			responseBody, err := readResponseBody(response)

			if err != nil {
				t.Errorf("could not parse body")
			}

			if responseBody != tc.expectedResponseBody {
				t.Errorf("got responseBody %q | want %q", responseBody, tc.expectedResponseBody)
			}
		})
	}
}

func TestParseRetryAfterHeader(t *testing.T) {
	type testCase struct {
		name             string
		retryAfterHeader string
		currentTime      time.Time
		expectedDuration time.Duration
		expectedError    error
	}

	testCases := []testCase{
		{
			name:             "retry-after header is a string ('a while')",
			retryAfterHeader: "a while",
			currentTime:      time.Now(),
			expectedDuration: 0,
			expectedError:    errors.New("[3] retry-after head is invalid: a while"),
		},
		{
			name:             "retry-after header is an integer as string (3 seconds)",
			retryAfterHeader: "3",
			currentTime:      time.Now(),
			expectedDuration: 3 * time.Second,
			expectedError:    nil,
		},
		{
			name:             "retry-after header is a timestamp as string (5 seconds)",
			retryAfterHeader: time.Date(2023, 4, 1, 12, 0, 5, 0, time.UTC).UTC().Format(http.TimeFormat),
			currentTime:      time.Date(2023, 4, 1, 12, 0, 0, 0, time.UTC),
			expectedDuration: 5 * time.Second,
			expectedError:    nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			retryDuration, err := parseRetryAfterHeader(tc.retryAfterHeader, tc.currentTime)

			if err != nil {
				// how to handle these properly in table driven tests?
			}

			if retryDuration != tc.expectedDuration {
				t.Errorf("retry duration got %v | want %v", retryDuration, tc.expectedDuration)
			}
		})
	}
}

func TestHandleTooManyRequestsResponse(t *testing.T) {
	type testCase struct {
		name             string
		retryAfterHeader string
		expectedDuration time.Duration
		expectedError    error
	}

	testCases := []testCase{
		{
			name:             "retry-after header is an integer as string (4 seconds)",
			retryAfterHeader: "4",
			expectedDuration: 4 * time.Second,
			expectedError:    nil,
		}, {
			name:             "retry-after header is an integer as string (10 seconds)",
			retryAfterHeader: "10",
			expectedDuration: 0,
			expectedError:    errors.New("[2] retry-after duration is more than 5 seconds"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			response := &http.Response{Header: http.Header{"Retry-After": []string{tc.retryAfterHeader}}}
			retryDuration, err := handleTooManyRequestsResponse(response)

			if err != nil {
				// how to handle these properly in table driven tests?
			}

			if retryDuration != tc.expectedDuration {
				t.Errorf("retry duration got %v | want %v", retryDuration, tc.expectedDuration)
			}
		})
	}
}
