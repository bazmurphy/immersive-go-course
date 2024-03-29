package main

import (
	"fmt"
	"net/http"
)

// NOTES ----------

// In Go, when handling HTTP requests,
// you typically set the HTTP response body
// by writing to the http.ResponseWriter object passed into your handler function.

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World!"))
		// $ curl -i "http://localhost:8080/"
		// HTTP/1.1 200 OK
		// Date: Fri, 29 Mar 2024 08:09:41 GMT
		// Content-Length: 12
		// Content-Type: text/plain; charset=utf-8
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

	http.ListenAndServe(":8080", nil)
}
