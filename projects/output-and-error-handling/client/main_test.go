package main

import (
	"bytes"
	"errors"
	"fmt"
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
			// create an http server
			// [I don't like this^ because this is not the behaviour of our "real server"]
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

			fmt.Println(actualResponseBody, actualRetryDuration, err)

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

// Notes on Constructing an HTTP Response Body (this is really not intuitive)

// field Body is a io.ReadCloser:

// type ReadCloser interface {
// 	Reader
// 	Closer
// }

// func NopCloser(r Reader) ReadCloser {
// 	if _, ok := r.(WriterTo); ok {
// 		return nopCloserWriterTo{r}
// 	}
// 	return nopCloser{r}
// }

// NopCloser returns a ReadCloser with a no-op Close method wrapping the provided [Reader] r
