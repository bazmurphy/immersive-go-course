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

func fetchImages(connection *pgx.Conn) ([]Image, error) {
	// make a select query to the database
	rows, err := connection.Query(context.Background(), "SELECT title, alt_text, url FROM images")
	if err != nil {
		return nil, fmt.Errorf("error making the query to the database: %w", err)
	}
	defer rows.Close()

	// make an empty slice to contain the images
	var images []Image

	// iterate through each row
	for rows.Next() {
		// which approach is better? [1] or [2]

		// [1] make nil value strings
		// var title, altText, url string
		// [1] get the values from the row and set them on the nil value strings
		// err = rows.Scan(&title, &altText, &url)

		// [2] make an empty Image struct
		var image Image
		// [2] get the values from the row and set them on the Image struct
		err = rows.Scan(&image.Title, &image.AltText, &image.URL)
		if err != nil {
			// Q: in this situation should we keep going regardless or exit with an error?
			// as in is it better to return some rows rather than none?

			// [1] just keep going (does a Scan error cause a panic?)
			// continue
			// [2] error:
			return nil, fmt.Errorf("error scanning the database row: %w", err)
		}

		// [1] append an Image struct with the 3 string values
		// images = append(images, Image{Title: title, URL: url, AltText: altText})
		// [2] append the Image struct to the images slice
		images = append(images, image)
	}

	return images, nil
}

func addNewImage(connection *pgx.Conn, newImage Image) error {
	// make an insert query to the database
	// (!) use .Exec because an INSERT doesn't return any rows
	rows, err := connection.Exec(context.Background(), "INSERT INTO images(title, alt_text, url) VALUES ($1, $2, $3);", newImage.Title, newImage.AltText, newImage.URL)
	fmt.Println("rows", rows, "err", err)
	if err != nil {
		return fmt.Errorf("error inserting new image into the database: %w", err)
	}
	return nil
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
		fmt.Fprintln(os.Stderr, "error no value on the environment variable DATABASE_URL")
		os.Exit(1)
	}

	// create a connection to the database
	// (!) investigate context
	connection, err := pgx.Connect(context.Background(), connectionString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error connecting to the database: %v\n", err)
		os.Exit(1)
	}
	defer connection.Close(context.Background())

	http.HandleFunc("/images.json", func(w http.ResponseWriter, r *http.Request) {
		indentQueryParameter := r.URL.Query().Get("indent")

		var identInteger int

		if indentQueryParameter != "" {
			identInteger, err = strconv.Atoi(indentQueryParameter)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if identInteger < 0 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		switch r.Method {
		case "GET":
			images, err := fetchImages(connection)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error fetching images from the database: %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var jsonByteSlice []byte

			if identInteger > 0 {
				indentString := strings.Repeat(" ", identInteger)
				jsonByteSlice, err = json.MarshalIndent(images, "", indentString)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			} else {
				jsonByteSlice, err = json.Marshal(images)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}

			// In the project readme curl shows "Content-Type: text/json" but I usually see "Content-Type: application/json"
			// So I initially tried:
			// w.Header().Add("Content-Type", "text/json")
			// Q: But what is the difference(?)
			// A: text/json is not a standard MIME type recognized by most clients, including web browsers. The standard MIME type for JSON data is application/json.
			// More: https://www.ietf.org/rfc/rfc4627.txt The MIME media type for JSON text is application/json.
			w.Header().Add("Content-Type", "application/json")

			// Which is better?
			// [1] convert the byte slice to a string before writing it to the writer
			// fmt.Fprint(w, string(imageByteSlice))
			// [2] directly write the byte slice to the writer (i would assume better performance?)
			_, err = w.Write(jsonByteSlice)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
		case "POST":
			// create a new image nil value struct
			var newImage Image

			// decode the body as json and store it in the new image struct
			// (!) more error handling here necessary here
			err := json.NewDecoder(r.Body).Decode(&newImage)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error decoding the json: %v\n", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// TODO: before we try to add it to the database have to make sure it is valid/sanitised/etc.

			// add the new image to the database
			err = addNewImage(connection, newImage)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error adding new image to the database: %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.Header().Add("Content-Type", "application/json")

			var jsonByteSlice []byte

			if identInteger > 0 {
				indentString := strings.Repeat(" ", identInteger)
				jsonByteSlice, err = json.MarshalIndent(newImage, "", indentString)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			} else {
				// marshal the new image struct into a json byte slice
				jsonByteSlice, err = json.Marshal(newImage)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}

			// write the json byte slice to the response body
			_, err = w.Write(jsonByteSlice)
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

	http.ListenAndServe(":8080", nil)
}
