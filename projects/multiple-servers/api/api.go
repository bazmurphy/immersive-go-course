package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"
)

func Run(databaseURL string, port string) {
	database, err := pgx.Connect(context.Background(), databaseURL)
	if err != nil {
		fmt.Printf("error: failed to connect to database: %v", err)
		os.Exit(1)
	}

	defer database.Close(context.Background())

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		html := `
        <!DOCTYPE html>
        <html lang="en">
        <head>
            <meta charset="UTF-8" />
            <meta name="viewport" content="width=device-width, initial-scale=1.0" />
            <title>API Server</title>
        </head>
        <body>
            <h1>API Server Root /</h1>
        </body>
        </html>
    `
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, html)
	})

	http.HandleFunc("/images.json", func(w http.ResponseWriter, r *http.Request) {
		// When switching to the live images from the database, we get a CORS error:
		// Cross-Origin Request Blocked: The Same Origin Policy disallows reading the remote resource at http://localhost:8081/images.json. (Reason: CORS header ‘Access-Control-Allow-Origin’ missing). Status code: 200.
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case http.MethodGet:
			images, err := GetAllImages(database)
			if err != nil {
				fmt.Printf("error: getting all images: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			imagesAsJSON, err := json.Marshal(images)
			if err != nil {
				fmt.Printf("error: failed marshalling images slice to json: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write(imagesAsJSON)

		case http.MethodPost:
			var newImage Image

			err := json.NewDecoder(r.Body).Decode(&newImage)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: decoding the json: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			insertImage, err := AddImage(database, newImage)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: adding new image to the database: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			insertImageBytes, err := json.Marshal(insertImage)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: adding new image to the database: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.Header().Add("Content-Type", "application/json")

			_, err = w.Write(insertImageBytes)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	})

	fmt.Printf("api server running on on port %s\n", port)

	http.ListenAndServe(":"+port, nil)
}
