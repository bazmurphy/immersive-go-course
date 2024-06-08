package main

import (
	"bytes"
	"errors"
	"io"
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
			name:                 "valid sunny response body",
			expectedResponseBody: "Today it will be sunny!"},
		{
			name:                 "rainy response body",
			expectedResponseBody: "I'd bring an umbrella, just in case...",
		},
		{
			name:                 "empty response body",
			expectedResponseBody: "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// construct an http response
			response := &http.Response{
				Body: io.NopCloser(bytes.NewBuffer([]byte(testCase.expectedResponseBody))),
			}

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
			actualRetryDuration, err := parseRetryAfterHeader(testCase.retryAfterHeader, testCase.currentTime)

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

			if actualRetryDuration != testCase.expectedDuration {
				t.Errorf("retryDuration: actual %v | expected %v", actualRetryDuration, testCase.expectedDuration)
			}
		})
	}
}

func TestHandleTooManyRequestsResponse(t *testing.T) {
	testCases := []struct {
		name                  string
		retryAfterHeader      string
		expectedRetryDuration time.Duration
		expectedError         error
	}{
		{
			name:                  "retry-after header is an integer as string (4 seconds)",
			retryAfterHeader:      "4",
			expectedRetryDuration: 4 * time.Second,
			expectedError:         nil,
		},
		{
			name:                  "retry-after header is an integer as string (10 seconds)",
			retryAfterHeader:      "10",
			expectedRetryDuration: 0,
			expectedError:         errors.New("[2] retry-after duration is more than 5 seconds"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// construct an http response
			expectedResponse := &http.Response{
				Header: http.Header{
					"Retry-After": []string{testCase.retryAfterHeader},
				},
			}

			// passing that response to the function we want to test
			actualRetryDuration, err := handleTooManyRequestsResponse(expectedResponse)

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

			if actualRetryDuration != testCase.expectedRetryDuration {
				t.Errorf("retryDuration: actual %v | expected %v", actualRetryDuration, testCase.expectedRetryDuration)
			}
		})
	}
}

func TestHandleStatusCode(t *testing.T) {
	testCases := []struct {
		name                  string
		statusCode            int
		retryAfterHeader      string
		expectedResponseBody  string
		expectedRetryDuration time.Duration
		expectedError         error
	}{
		{
			name:                  "status 200 ok",
			statusCode:            http.StatusOK,
			expectedResponseBody:  "response body 1",
			expectedRetryDuration: 0,
			expectedError:         nil,
		},
		{
			name:                  "status 429 too many requests - with retry-after - 3 second duration",
			statusCode:            http.StatusTooManyRequests,
			retryAfterHeader:      "3",
			expectedResponseBody:  "",
			expectedRetryDuration: 3 * time.Second,
			expectedError:         nil,
		},
		{
			name:                  "status 429 too many requests - with retry-after - 10 second duration",
			statusCode:            http.StatusTooManyRequests,
			retryAfterHeader:      "10",
			expectedResponseBody:  "",
			expectedRetryDuration: 0,
			expectedError:         errors.New("[1] handleTooManyRequestsResponse failed: [2] retry-after duration is more than 5 seconds"),
		},
		{
			name:                  "status 500 internal server error",
			statusCode:            http.StatusInternalServerError,
			expectedResponseBody:  "",
			expectedRetryDuration: 0,
			expectedError:         errors.New("[1] unhandled response status code"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// construct an http response
			expectedResponse := &http.Response{
				StatusCode: testCase.statusCode,
				Header: http.Header{
					"Retry-After": []string{testCase.retryAfterHeader},
				},
				Body: io.NopCloser(bytes.NewBufferString(testCase.expectedResponseBody)),
			}

			// passing that response to the function we want to test
			actualResponseBody, actualRetryDuration, err := handleStatusCode(expectedResponse)

			if err != nil {
				if testCase.expectedError != nil {
					// if we were expecting an error
					// we need to use .Error() because they are separate instances (reference types and so non-comparable)
					// and in this case because we have error propagation we need to check it contains (rather than constructing it all)
					if err.Error() != testCase.expectedError.Error() {
						t.Errorf("error: received %v | expected %v", err, testCase.expectedError)
					}
				} else {
					// if we were not expecting an error
					t.Errorf("received an error %v | expected no error", err)
				}
			}

			if testCase.expectedResponseBody != "" {
				if actualResponseBody != testCase.expectedResponseBody {
					t.Errorf("responseBody: actual %v | expected %v", actualResponseBody, testCase.expectedResponseBody)
				}
			}

			if testCase.expectedRetryDuration != 0 {
				if actualRetryDuration != testCase.expectedRetryDuration {
					t.Errorf("retryDuration: actual %v | expected %v", actualRetryDuration, testCase.expectedRetryDuration)
				}
			}
		})
	}
}

// ----------------

// Notes on constructing HTTP Response Body (for tests?)

// In the `http.Response` struct, the `Body` field is of type `i`o.ReadCloser`.
// This means that it should implement both the `io.Reader` and `io.Closer` interfaces.
// The `io.Reader` interface allows reading from the response body,
// while the `io.Closer` interface provides a way to close the response body when it's no longer needed.

// To create a response body manually, we need to provide an implementation of `io.ReadCloser`.
// This is where `io.NopCloser` comes into play.

// `io.NopCloser` is a function provided by the io package.
// It takes an `io.Reader` as input and returns an `io.ReadCloser`.
// The returned `io.ReadCloser` has a no-op (no operation) `Close` method, meaning it doesn't do anything when called.
// The purpose of `io.NopCloser` is to transform an `io.Reader` into an `io.ReadCloser` by adding a dummy `Close` method.

// func NopCloser(r io.Reader) io.ReadCloser

// response := &http.Response{
// 	Body: io.NopCloser(bytes.NewBufferString(responseBody)),
// }

// bytes.NewBufferString(testCase.responseBody) creates a *bytes.Buffer from the responseBody string.
// This buffer implements the io.Reader interface, allowing us to read from it.

// We pass the *bytes.Buffer to io.NopCloser, which returns an io.ReadCloser.
// The returned io.ReadCloser wraps the *bytes.Buffer and adds a no-op Close method.

// Finally, we assign the resulting io.ReadCloser to the Body field of the http.Response struct.

// By using io.NopCloser(bytes.NewBufferString(...))
// we create a response body that can be read from and closed (even though the Close method doesn't do anything in this case).

// ----------------

func TestMakeGetRequest(t *testing.T) {
	testCases := []struct {
		name                  string
		statusCode            int
		retryAfterHeader      string
		responseBody          string
		expectedResponseBody  string
		expectedRetryDuration time.Duration
		expectedError         error
	}{
		{
			name:                  "status 200 ok with response body",
			statusCode:            http.StatusOK,
			retryAfterHeader:      "",
			responseBody:          "Today it will be sunny!",
			expectedResponseBody:  "Today it will be sunny!",
			expectedRetryDuration: 0,
			expectedError:         nil,
		},
		{
			name:                  "status 429 too many requests with 3 second duration",
			statusCode:            http.StatusTooManyRequests,
			retryAfterHeader:      "3",
			responseBody:          "",
			expectedResponseBody:  "",
			expectedRetryDuration: 3 * time.Second,
			expectedError:         nil,
		},
		{
			name:                  "status 500 internal server error",
			statusCode:            http.StatusInternalServerError,
			retryAfterHeader:      "",
			responseBody:          "",
			expectedResponseBody:  "",
			expectedRetryDuration: 0,
			expectedError:         errors.New("[0] handleStatusCode failed: [1] unhandled response status code"),
		},
	}

	for _, testCase := range testCases {
		// construct an http server
		server := httptest.NewServer(
			http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					if testCase.retryAfterHeader != "" {
						w.Header().Set("Retry-AFter", testCase.retryAfterHeader)
					}
					w.WriteHeader(testCase.statusCode)
					w.Write([]byte(testCase.responseBody))
				},
			),
		)
		defer server.Close()

		actualResponseBody, actualRetryDuration, err := makeGetRequest(server.URL)

		if err != nil {
			// if we expect there to be an error
			if testCase.expectedError != nil {
				t.Logf("actual error: %v", err)
				t.Logf("expected error: %v", testCase.expectedError)

				if err.Error() != testCase.expectedError.Error() {
					t.Errorf("error: received %v | expected %v", err, testCase.expectedError)
				}
			} else {
				t.Errorf("received an error %v | expected no error", err)
			}
		}

		// if we expect there to be a response body
		if testCase.expectedResponseBody != "" {
			if actualResponseBody != testCase.expectedResponseBody {
				t.Errorf("responseBody: actual %v | expected %v", actualResponseBody, testCase.expectedResponseBody)
			}
		}

		// if we expect there to be a retry duration
		if testCase.expectedRetryDuration != 0 {
			if actualRetryDuration != testCase.expectedRetryDuration {
				t.Errorf("retryDuration: actual %v | expected %v", actualRetryDuration, testCase.expectedRetryDuration)
			}
		}
	}
}
