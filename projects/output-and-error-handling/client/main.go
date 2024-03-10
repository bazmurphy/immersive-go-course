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

// Current Issues ----------
// [1] There is too much going on in a single function, it needs breaking up
// [2] The nested conditional logic i am not a fan of...
// it's brittle...
// I would like to know the Idiomatic Go way of writing this...
// with correct conditional branching that can all be handled elegantly (with error checking)
// [3] Not all the messages are "perfectly" user friendly

func makeGetRequest(url string) {
	// make a basic GET request to the server
	response, err := http.Get(url)

	// if there is an error making the GET request
	// print it (improve the message later)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Request Failed. Error:", err)
		return
	}

	// handle the various response status codes (switch/case extensible for more status codes)
	switch response.StatusCode {
	case http.StatusOK:
		// fmt.Println("200 OK")

		// use the defer keyword early to close the body at the end of the scoped block regardless
		defer response.Body.Close()

		// read the response body
		responseBody, err := io.ReadAll(response.Body)

		// if there is an error reading the response body
		// print it (improve the message later)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Parsing Response Body Failed. Error:", err)
			return
		}

		// print out the response body
		// deliberately/explicitly using os.Stdout not fmt.Println even though unnecessary (for the sake of the exercise)
		fmt.Fprintln(os.Stdout, string(responseBody))

	case http.StatusTooManyRequests:
		// fmt.Println("429 Too Many Requests")

		// have a look at the header to try read the "Retry-After"
		// fmt.Println(response.Header)

		// get the retry after header string
		retryAfterHeaderString := response.Header.Get("Retry-After")
		// fmt.Fprintln(os.Stderr, "Retry After:", retryAfterHeaderString)

		// (I want a more elegant way to handle the attempted conversion of each variant of the Retry-After Header String)

		// i think this could/should handle more than "a while"
		if retryAfterHeaderString == "a while" {
			// we should give up and tell the user we can't get them the weather
			fmt.Fprintln(os.Stderr, "Giving up. We can't get you the Weather.")
			return
		}

		// try to get an integer from the retryAfter string
		retryAfterInteger, err := strconv.Atoi(retryAfterHeaderString)

		if err != nil {
			// fmt.Println("Retry-After is not an integer")

			// (the following is bad, because we should not do any logic beyond error handling in this scoped block)

			// try to get a timestamp from the retryAfter string
			// temporarily not use the err
			retryAfterTimeStamp, err := time.Parse(time.RFC1123, retryAfterHeaderString)
			// fmt.Println("retryAfterTimeStamp", retryAfterTimeStamp)

			if err != nil {
				// this cannot actually trigger from current server code
				// message not "user friendly" can change after
				fmt.Fprintln(os.Stderr, "Retry-After Header is neither 'a while', an integer or a timestamp")
				return
			}

			// get the time now
			now := time.Now()
			// fmt.Println("now", now)

			// get the time difference
			timeDifference := retryAfterTimeStamp.Sub(now)
			// fmt.Println("timeDifference", timeDifference)

			// round it up to the nearest second, and convert to an int
			timeDifferenceInSeconds := int(math.Ceil(timeDifference.Seconds()))
			// fmt.Println("timeDifferenceInSeconds:", timeDifferenceInSeconds)

			// assign it to the previously created variable
			retryAfterInteger = timeDifferenceInSeconds
		}

		// handle the various retry after durations
		switch {
		case retryAfterInteger > 5:
			// we should give up and tell the user we can't get them the weather
			fmt.Fprintln(os.Stderr, "Giving up. We can't get you the Weather.")
		case retryAfterInteger > 1:
			// we should notify the user that things may be a bit slow because we're doing a retry
			fmt.Fprintf(os.Stderr, "Things may be a bit slow because we're doing a retry after %d seconds...\n", retryAfterInteger)

			// sleep for X seconds
			time.Sleep(time.Duration(retryAfterInteger))

			// make the request again
			makeGetRequest(url)
		default:
			// retry after 1 second
			fmt.Fprintln(os.Stderr, "Retrying after 1 second...")

			// sleep for 1 second
			time.Sleep(time.Duration(retryAfterInteger))

			// make the request again
			makeGetRequest(url)
		}
	}
}

func main() {
	// just for testing purposes, make an infinite loop
	for {
		// make the GET Request
		makeGetRequest("http://localhost:8080")
		// sleep for 1 second
		time.Sleep(1 * time.Second)
	}
}
