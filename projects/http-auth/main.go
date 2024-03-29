package main

import (
	"fmt"
	"html"
	"io"
	"net/http"
)

// This has comments and prints everywhere as I learn

func main() {
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

					keyString := fmt.Sprintf("<li>%s: [", html.EscapeString(key))
					// fmt.Println("keyString", keyString)

					fmt.Fprint(w, keyString)

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
				w.Write([]byte("error reading the request body"))
			}

			if len(requestBody) == 0 {
				w.Write([]byte("no body was received with the request"))
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

	http.ListenAndServe(":8080", nil)
}
