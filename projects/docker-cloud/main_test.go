package main

import (
	"io"
	"log"
	"net/http"
	"testing"

	"github.com/ory/dockertest/v3"
)

func TestRoutes(t *testing.T) {
	type testCase struct {
		name                   string
		route                  string
		expectedResponseStatus int
		expectedResponseBody   string
	}

	testCases := []testCase{
		{
			name:                   "TestRootRoute",
			route:                  "/",
			expectedResponseStatus: http.StatusOK,
			expectedResponseBody:   "Hello, world.\n",
		},
		{
			name:                   "TestPingRoute",
			route:                  "/ping",
			expectedResponseStatus: http.StatusOK,
			expectedResponseBody:   "pong\n",
		},
	}

	// `NewPool` creates a new pool
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("fatal error: could not construct pool: %s", err)
	}

	// `Ping` pings the docker server
	err = pool.Client.Ping()
	if err != nil {
		log.Fatalf("fatal error: could not connect to docker: %s", err)
	}

	// `Run` starts a docker container
	resource, err := pool.Run("bazmurphy/docker-cloud", "latest", nil)
	if err != nil {
		t.Fatalf("error: could not start resource: %s", err)
	}

	defer pool.Purge(resource)

	var response *http.Response

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// `Retry` is an exponential backoff retry helper
			err = pool.Retry(func() error {
				response, err = http.Get("http://localhost:" + resource.GetPort("80/tcp") + testCase.route)
				if err != nil {
					t.Log("log: waiting for container to be ready")
					return err
				}
				return nil
			})

			if err != nil {
				t.Fatalf("error: could not connect to container: %s", err)
			}

			defer response.Body.Close()

			if response.StatusCode != testCase.expectedResponseStatus {
				t.Errorf("error: response status code | expected %d | received %d", testCase.expectedResponseStatus, response.StatusCode)
			}

			responseBody, err := io.ReadAll(response.Body)
			if err != nil {
				t.Fatalf("fatal error: could not read the response body: %s", err)
			}

			responseBodyAsString := string(responseBody)

			if responseBodyAsString != testCase.expectedResponseBody {
				t.Errorf("error: response body | expected %s | received %s", testCase.expectedResponseBody, responseBodyAsString)
			}
		})
	}
}
