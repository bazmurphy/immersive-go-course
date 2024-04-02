package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

type Image struct {
	Title   string `json:"title"`
	AltText string `json:"alt_text"`
	URL     string `json:"url"`
}

func main() {
	// load the .env file
	err := godotenv.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading the .env file: %v\n", err)
		os.Exit(1)
	}

	// get the database connection string
	connectionString := os.Getenv("DATABASE_URL")
	if connectionString == "" {
		fmt.Fprintln(os.Stderr, "Error no value on the environment variable DATABASE_URL")
		os.Exit(1)
	}

	// create a connection to the database
	// (!) investigate context
	connection, err := pgx.Connect(context.Background(), connectionString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to the database: %v\n", err)
		os.Exit(1)
	}
	defer connection.Close(context.Background())

	// make a select query to the database
	rows, err := connection.Query(context.Background(), "SELECT title, alt_text, url FROM images")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error making the query to the database: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	// make an empty slice to contain the images
	var images []Image

	// iterate through each row
	for rows.Next() {
		// make an empty Image struct
		var image Image
		// get the values from the row and set them on the Image struct
		err = rows.Scan(&image.Title, &image.AltText, &image.URL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error scanning the database row: %v\n", err)
			continue
		}

		// append the Image struct to the images slice
		images = append(images, image)
	}

	// Original Hardcoded images
	// images := []Image{
	// 	{
	// 		Title:   "Sunset",
	// 		AltText: "Clouds at sunset",
	// 		URL:     "https://images.unsplash.com/photo-1506815444479-bfdb1e96c566?ixlib=rb-1.2.1&ixid=MnwxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8&auto=format&fit=crop&w=1000&q=80"},
	// 	{
	// 		Title:   "Mountain",
	// 		AltText: "A mountain at sunset",
	// 		URL:     "https://images.unsplash.com/photo-1540979388789-6cee28a1cdc9?ixlib=rb-1.2.1&ixid=MnwxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8&auto=format&fit=crop&w=1000&q=80"},
	// }

	http.HandleFunc("/images.json", func(w http.ResponseWriter, r *http.Request) {
		var imageByteSlice []byte
		var err error

		indentQueryParameter := r.URL.Query().Get("indent")

		// there is error handling repetition here (panic aversion)
		// despite me using the var err error above to try to prevent that... suggestions(?)
		if indentQueryParameter != "" {
			identInteger, err := strconv.Atoi(indentQueryParameter)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if identInteger < 0 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			indentString := strings.Repeat(" ", identInteger)
			imageByteSlice, err = json.MarshalIndent(images, "", indentString)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			imageByteSlice, err = json.Marshal(images)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		// In the project readme curl shows "Content-Type: text/json" but I usually see "Content-Type: application/json"
		// I tried:
		// w.Header().Add("Content-Type", "text/json")
		// Q: But what is the difference(?)
		// A: text/json is not a standard MIME type recognized by most clients, including web browsers. The standard MIME type for JSON data is application/json.
		// More: https://www.ietf.org/rfc/rfc4627.txt The MIME media type for JSON text is application/json.
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Which is better?
		// [1] convert the byte slice to a string before writing it to the writer
		// fmt.Fprint(w, string(imageByteSlice))
		// [2] directly write the byte slice to the writer (i would assume better performance?)
		_, err = w.Write(imageByteSlice)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	http.ListenAndServe(":8080", nil)
}
