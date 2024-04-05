package main

import (
	"encoding/base64"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"golang.org/x/time/rate"
)

// This has comments and prints everywhere as I learn

func main() {
	// try to load the .env file
	err := godotenv.Load()

	if err != nil {
		log.Fatal("error loading the .env file")
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// $ curl -i 'http://localhost:8080/'
		// HTTP/1.1 200 OK
		// Date: Fri, 29 Mar 2024 08:09:41 GMT
		// Content-Length: 12
		// Content-Type: text/plain; charset=utf-8

		// w.Write([]byte("Hello World!"))

		// > curl -i 'http://localhost:8080/'
		// HTTP/1.1 200 OK
		// Content-Type: text/html
		// Date: Sun, 24 Jul 2022 09:42:30 GMT
		// Content-Length: 42

		// <!DOCTYPE html><html><em>Hello, world</em>

		// w.Header().Add("Content-Type", "text/html")
		// w.Header().Add("Content-Type", "text/plain")
		// html := "<!DOCTYPE html><html><em>Hello, world</em>"
		// w.Write([]byte(html))

		switch r.Method {
		case http.MethodGet:
			w.Header().Add("Content-Type", "text/html")
			// html := "<!DOCTYPE html><html><em>Hello, world</em>"
			// w.Write([]byte(html))

			fmt.Fprintf(w, "<!DOCTYPE html>\n<html>\n<em>Hello, world</em>\n")

			queryParameters := r.URL.Query()
			// fmt.Println("queryParameters", queryParameters)
			// queryParameters map[test:[123]]
			// a map of key: string and value of []string

			// fmt.Println("len(queryParameters)", len(queryParameters))

			if len(queryParameters) > 0 {
				// fmt.Println("queryParameters", queryParameters)
				fmt.Fprintf(w, "<p>Query parameters:</p>\n<ul>\n")

				for key, values := range queryParameters {
					// fmt.Println("key", key, "value", value)

					fmt.Fprintf(w, "<li>%s: [", html.EscapeString(key))

					for _, value := range values {
						// fmt.Println("index", index, "element", element)
						fmt.Fprint(w, html.EscapeString(value))
					}

					fmt.Fprintf(w, "]</li>\n")
				}

				fmt.Fprintf(w, "</ul>\n</html>")
			}
		case http.MethodPost:
			// fmt.Println("request", r)

			// without a body

			// &{POST / HTTP/1.1 1 1 map[Accept:[*/*] Accept-Encoding:[gzip, deflate, br] Cache-Control:[no-cache] Connection:[keep-alive] Content-Length:[0] Postman-Token:[7c51a315-20a3-4543-a9c4-5f0f45590c12] User-Agent:[PostmanRuntime/7.32.1]] {} <nil> 0 [] false localhost:8080 map[] map[] <nil> map[] [::1]:57008 / <nil> <nil> <nil> 0xc00009e050 0xc00012a180 [] map[]}

			// with a body

			// &{POST / HTTP/1.1 1 1 map[Accept:[*/*] Accept-Encoding:[gzip, deflate, br] Cache-Control:[no-cache] Connection:[keep-alive] Content-Length:[32] Content-Type:[text/plain] Postman-Token:[2f070da0-2e81-4c31-ba67-94d3f2d6fe4d] User-Agent:[PostmanRuntime/7.32.1]] 0xc000092040 <nil> 32 [] false localhost:8080 map[] map[] <nil> map[] [::1]:56178 / <nil> <nil> <nil> 0xc00009e050 0xc00012a180 [] map[]}

			requestBody, err := io.ReadAll(r.Body)

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("error reading the request body"))
				return
			}

			if len(requestBody) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("no body was received with the request"))
				return
			}

			w.Write([]byte(requestBody))
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/200", func(w http.ResponseWriter, r *http.Request) {
		// > curl -i 'http://localhost:8080/200'
		// HTTP/1.1 200 OK
		// Date: Sat, 25 Jun 2022 11:16:17 GMT
		// Content-Length: 3
		// Content-Type: text/plain; charset=utf-8

		// 200

		w.WriteHeader(http.StatusOK)
		// we use Fprint to write it to the writer
		fmt.Fprintf(w, "200")
	})

	http.HandleFunc("/404", func(w http.ResponseWriter, r *http.Request) {
		// > curl -i 'http://localhost:8080/404'
		// HTTP/1.1 404 Not Found
		// Content-Type: text/plain; charset=utf-8
		// X-Content-Type-Options: nosniff
		// Date: Sat, 25 Jun 2022 11:17:29 GMT
		// Content-Length: 19

		// 404 page not found

		// w.WriteHeader(http.StatusNotFound)
		// fmt.Fprintf(w, "404 page not found")

		// instead we can use http.NotFound method
		http.NotFound(w, r)
	})

	http.HandleFunc("/500", func(w http.ResponseWriter, r *http.Request) {
		// > curl -i 'http://localhost:8080/500'
		// HTTP/1.1 500 Internal Server Error
		// Date: Sat, 25 Jun 2022 11:16:30 GMT
		// Content-Length: 21
		// Content-Type: text/plain; charset=utf-8

		// Internal server error

		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Internal server error")
	})

	http.HandleFunc("/authenticated", func(w http.ResponseWriter, r *http.Request) {
		// get the authorization header
		authorizationHeader := r.Header.Get("Authorization")
		// fmt.Println("authorizationHeader", authorizationHeader)

		// there was no authentication header
		if len(authorizationHeader) == 0 {
			w.Header().Set("WWW-Authenticate", `Basic realm="User Authentication"`)
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, "401 Unauthorized [no authorization header]")
			return
		}

		// authorizationHeader header should be in the format "Basic <base64 encoded payload>"
		authorizationString := strings.Split(authorizationHeader, " ")
		// fmt.Println("authorizationString", authorizationString)

		// must be 2 composite parts and the first must be "Basic"
		if len(authorizationString) != 2 || authorizationString[0] != "Basic" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "400 Bad Request [malformed authorization header]")
			return
		}

		// decode the base64 encoded payload
		payload, err := base64.StdEncoding.DecodeString(authorizationString[1])
		// fmt.Println("payload", payload, "err", err)

		// the payload couldn't be decoded
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "400 Bad Request [couldn't decode the authorization header payload]")
			return
		}

		// split the payload into it's composite parts
		credentials := strings.Split(string(payload), ":")
		// fmt.Println("credentials", credentials)

		if len(credentials) != 2 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "400 Bad Request [invalid credentials format]")
		}

		username := credentials[0]
		password := credentials[1]
		// fmt.Println("username", username, "password", password)

		// username/password must not be empty
		if len(username) == 0 || len(password) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "400 Bad Request [either username or password were missing]")
			return
		}

		// get the username/password from the .env file
		usernameFromEnv := os.Getenv("AUTH_USERNAME")
		passwordFromEnv := os.Getenv("AUTH_PASSWORD")

		// see if the username and password match from the .env
		if username != usernameFromEnv || password != passwordFromEnv {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, "401 Unauthorized [either username or password were not correct]")
			return
		}

		// if we reach here the authorization was successful
		w.Header().Add("Content-Type", "text/html")
		fmt.Fprintf(w, "<!DOCTYPE html>\n<html>\n<p>Hello %s!</p>\n", html.EscapeString(username))
	})

	// create a limiter
	limiter := rate.NewLimiter(100, 30)

	http.HandleFunc("/limited", func(w http.ResponseWriter, r *http.Request) {
		// use the limiter
		if limiter.Allow() {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "200")
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, "503 Service Unavailable [rate limited]")
		}
	})

	http.ListenAndServe(":8080", nil)
}

// Notes ----------

// Basic
// See RFC 7617, base64-encoded credentials.
// https://datatracker.ietf.org/doc/html/rfc7617

//  The "Basic" authentication scheme offers very poor security, but is widely supported and easy to set up. It is introduced in more detail below.

// Basic authentication scheme

// The "Basic" HTTP authentication scheme is defined in RFC 7617, which transmits credentials as user ID/password pairs, encoded using base64.

// Security of basic authentication

// As the user ID and password are passed over the network as clear text (it is base64 encoded, but base64 is a reversible encoding), the basic authentication scheme is not secure. HTTPS/TLS should be used with basic authentication. Without these additional security enhancements, basic authentication should not be used to protect sensitive or valuable information.

// WWW-Authenticate

// The HTTP WWW-Authenticate response header defines the HTTP authentication methods ("challenges") that might be used to gain access to a specific resource.

// A server using HTTP authentication will respond with a 401 Unauthorized response to a request for a protected resource. This response must include at least one WWW-Authenticate header and at least one challenge, to indicate what authentication schemes can be used to access the resource (and any additional data that each particular scheme needs).

// After receiving the WWW-Authenticate header, a client will typically prompt the user for credentials, and then re-request the resource. This new request uses the Authorization header to supply the credentials to the server, encoded appropriately for the selected "challenge" authentication method. The client is expected to select the most secure of the challenges it understands (note that in some cases the "most secure" method is debatable).

// // Challenges specified in single header
// WWW-Authenticate: challenge1, ..., challengeN

// Possible challenge formats (scheme dependent)
// WWW-Authenticate: <auth-scheme>
// WWW-Authenticate: <auth-scheme> realm=<realm>

// Basic authentication requires realm and allows for optional use of charset key, but does not support token68.

// WWW-Authenticate: Basic realm=<realm>
// WWW-Authenticate: Basic realm=<realm>, charset="UTF-8"

// ApacheBench ----------

// ab -n 10000 -c 100 'http://localhost:8080/limited'
// This is ApacheBench, Version 2.3 <$Revision: 1903618 $>
// Copyright 1996 Adam Twiss, Zeus Technology Ltd, http://www.zeustech.net/
// Licensed to The Apache Software Foundation, http://www.apache.org/

// Benchmarking localhost (be patient)
// Completed 1000 requests
// Completed 2000 requests
// Completed 3000 requests
// Completed 4000 requests
// Completed 5000 requests
// Completed 6000 requests
// Completed 7000 requests
// Completed 8000 requests
// Completed 9000 requests
// Completed 10000 requests
// Finished 10000 requests

// Server Software:
// Server Hostname:        localhost
// Server Port:            8080

// Document Path:          /limited
// Document Length:        3 bytes

// Concurrency Level:      100
// Time taken for tests:   0.596 seconds
// Complete requests:      10000
// Failed requests:        9911
//    (Connect: 0, Receive: 0, Length: 9911, Exceptions: 0)
// Non-2xx responses:      9911
// Total transferred:      1754927 bytes
// HTML transferred:       416529 bytes
// Requests per second:    16773.29 [#/sec] (mean)
// Time per request:       5.962 [ms] (mean)
// Time per request:       0.060 [ms] (mean, across all concurrent requests)
// Transfer rate:          2874.60 [Kbytes/sec] received

// Connection Times (ms)
//               min  mean[+/-sd] median   max
// Connect:        0    3   0.8      3       5
// Processing:     1    3   0.8      3       7
// Waiting:        0    2   0.8      2       5
// Total:          4    6   0.5      6       9

// Percentage of the requests served within a certain time (ms)
//   50%      6
//   66%      6
//   75%      6
//   80%      6
//   90%      7
//   95%      7
//   98%      7
//   99%      8
//  100%      9 (longest request)
