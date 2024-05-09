package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// relevant type definitions:

// type UploadedImageObject struct {
// 	ImageUrl string
// 	ID       int
// }

func TestUploadImagesToS3(t *testing.T) {
	testCases := []struct {
		name         string
		imagePaths   []string
		imageContent [][]byte
	}{
		{
			name:         "test 1",
			imagePaths:   []string{"image001.jpg", "image002.jpg", "image003.jpg"},
			imageContent: [][]byte{{}, {}, {}},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// create a temporary directory to upload the images from
			temporaryDirectory, err := os.MkdirTemp(".", "test-download-images")
			if err != nil {
				t.Fatalf("failed to create a temporary directory: %v", err)
			}
			defer os.RemoveAll(temporaryDirectory)

			// create temporary images
			for index, imagePath := range testCase.imagePaths {
				// establish the filepath
				filePath := filepath.Join(temporaryDirectory, imagePath)

				// create the image file
				imageFile, err := os.Create(filePath)
				if err != nil {
					t.Fatalf("failed to create a temporary image: %v", err)
				}
				defer imageFile.Close()

				// get the image file content from the test case
				imageFileContent := testCase.imageContent[index]

				// write the image file content to the image file
				_, err = imageFile.Write(imageFileContent)
				if err != nil {
					t.Fatalf("failed to write content to the temporary image: %v", err)
				}
			}

			// call the UploadImagesToS3 function that we want to test against
			uploadedImageObjects, err := UploadImagesToS3(temporaryDirectory)
			if err != nil {
				t.Fatalf("error: UploadImagesToS3 failed: %v", err)
			}

			// temporary print to allow compilation
			fmt.Println(uploadedImageObjects)

			// TODO: test against downloadedImageObjects
		})
	}
}
