package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// the relevant type definitions:

// type ParsedImageUrlObject struct {
// 	ImageUrl string
// 	ID       int
// }

//	type DownloadedImageObject struct {
//		ImageFilepath string
//		ID            int
//	}

func TestDownloadImages(t *testing.T) {
	testCases := []struct {
		name       string
		imagePaths []string
	}{
		{
			name:       "test 1",
			imagePaths: []string{"image001.jpg", "image002.jpg", "image003.jpg"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// create a temporary directory to download the images to
			temporaryDirectory, err := os.MkdirTemp(".", "test-download-images")
			if err != nil {
				t.Fatalf("failed to create a temporary directory: %v", err)
			}
			defer os.RemoveAll(temporaryDirectory)

			// create a test server to serve the images
			testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// TODO: implement the image serving using the imagePaths in the testCase
			}))
			defer testServer.Close()

			// create a slice of ParsedImageURLObjects
			parsedImageUrlObjects := []ParsedImageUrlObject{}
			for index, imagePath := range testCase.imagePaths {
				newParsedImageUrlObject := ParsedImageUrlObject{
					ImageUrl: testServer.URL + "/" + imagePath,
					ID:       index + 1,
				}
				parsedImageUrlObjects = append(parsedImageUrlObjects, newParsedImageUrlObject)
			}

			// call the DownImages function that we want to test against
			downloadedImageObjects, err := DownloadImages(parsedImageUrlObjects, temporaryDirectory)
			if err != nil {
				t.Fatalf("error: DownloadImages failed: %v", err)
			}

			// temporary print to allow compilation
			fmt.Println(downloadedImageObjects)

			// TODO: test against downloadedImageObjects
		})
	}
}
