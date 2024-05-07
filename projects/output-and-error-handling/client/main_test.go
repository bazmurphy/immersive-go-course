package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestReadResponseBody(t *testing.T) {
	testCases := []struct {
		name                 string
		expectedResponseBody string
	}{
		{
			name:                 "valid sunny response",
			expectedResponseBody: "Today it will be sunny!"},
		{
			name:                 "rainy response",
			expectedResponseBody: "I'd bring an umbrella, just in case...",
		},
		{
			name:                 "another response",
			expectedResponseBody: "another response body",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// create an http server
			server := httptest.NewServer(
				// create a single http handler
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte(testCase.expectedResponseBody))
				}),
			)
			defer server.Close()

			// get a response from the http server
			response, err := http.Get(server.URL)
			if err != nil {
				t.Fatalf("error: failed to make the get request: %v", err)
			}
			defer response.Body.Close()

			actualResponseBody, err := readResponseBody(response)

			if err != nil {
				t.Errorf("received error %v | expected NO error", err)
			}

			if actualResponseBody != testCase.expectedResponseBody {
				t.Errorf("actual response body: %v | expected response body: %v", testCase.expectedResponseBody, actualResponseBody)
			}
		})
	}
}

func TestParseRetryAfterHeader(t *testing.T) {
	testCases := []struct {
		name             string
		retryAfterHeader string
		currentTime      time.Time
		expectedDuration time.Duration
		expectedError    error
	}{
		{
			name:             "retry-after header is an integer as string (3 seconds)",
			retryAfterHeader: "3",
			currentTime:      time.Now(),
			expectedDuration: 3 * time.Second,
			expectedError:    nil,
		},
		{
			name:             "retry-after header is an integer as string (12 seconds)",
			retryAfterHeader: "12",
			currentTime:      time.Now(),
			expectedDuration: 12 * time.Second,
			expectedError:    nil,
		},
		{
			name:             "retry-after header is a timestamp as string (5 seconds)",
			retryAfterHeader: time.Date(2023, 4, 1, 12, 0, 5, 0, time.UTC).UTC().Format(http.TimeFormat),
			currentTime:      time.Date(2023, 4, 1, 12, 0, 0, 0, time.UTC),
			expectedDuration: 5 * time.Second,
			expectedError:    nil,
		},
		{
			name:             "retry-after header is a timestamp as string (10 seconds)",
			retryAfterHeader: time.Date(2023, 4, 1, 12, 0, 10, 0, time.UTC).UTC().Format(http.TimeFormat),
			currentTime:      time.Date(2023, 4, 1, 12, 0, 0, 0, time.UTC),
			expectedDuration: 10 * time.Second,
			expectedError:    nil,
		},
		{
			name:             "retry-after header is a string ('a while')",
			retryAfterHeader: "a while",
			currentTime:      time.Now(),
			expectedDuration: 0,
			expectedError:    errors.New("[3] retry-after header is invalid: a while"),
		},
		{
			name:             "retry-after header is a string ('something else')",
			retryAfterHeader: "something else",
			currentTime:      time.Now(),
			expectedDuration: 0,
			expectedError:    errors.New("[3] retry-after header is invalid: something else"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			retryDuration, err := parseRetryAfterHeader(testCase.retryAfterHeader, testCase.currentTime)

			if err != nil {
				if testCase.expectedError != nil {
					// if we were expecting an error
					// we need to use .Error() because they are separate instances (reference types and so non-comparable)
					if err.Error() != testCase.expectedError.Error() {
						t.Errorf("error: received %v | expected %v", err, testCase.expectedError)
					}
				} else {
					// if we were not expecting an error
					t.Errorf("error: received %v | expected NO error", err)
				}
			}

			if retryDuration != testCase.expectedDuration {
				t.Errorf("retryDuration: actual %v | expected %v", retryDuration, testCase.expectedDuration)
			}
		})
	}
}

func TestHandleTooManyRequestsResponse(t *testing.T) {
	testCases := []struct {
		name             string
		retryAfterHeader string
		expectedDuration time.Duration
		expectedError    error
	}{
		{
			name:             "retry-after header is an integer as string (4 seconds)",
			retryAfterHeader: "4",
			expectedDuration: 4 * time.Second,
			expectedError:    nil,
		},
		{
			name:             "retry-after header is an integer as string (10 seconds)",
			retryAfterHeader: "10",
			expectedDuration: 0,
			expectedError:    errors.New("[2] retry-after duration is more than 5 seconds"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			// create an http response
			expectedResponse := &http.Response{
				Header: http.Header{
					"Retry-After": []string{testCase.retryAfterHeader},
				},
			}

			// pass that response to the function we are testing here
			retryDuration, err := handleTooManyRequestsResponse(expectedResponse)

			if err != nil {
				if testCase.expectedError != nil {
					// if we were expecting an error
					// we need to use .Error() because they are separate instances (reference types and so non-comparable)
					if err.Error() != testCase.expectedError.Error() {
						t.Errorf("error: received %v | expected %v", err, testCase.expectedError)
					}
				} else {
					// if we were not expecting an error
					t.Errorf("received an error %v | expected no error", err)
				}
			}

			if retryDuration != testCase.expectedDuration {
				t.Errorf("retryDuration: actual %v | expected %v", retryDuration, testCase.expectedDuration)
			}
		})
	}
}
