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

// Session with Laura:

// package main

// import (
// 	"fmt"
// 	"os"
// 	"path/filepath"
// 	"testing"

// 	"github.com/aws/aws-sdk-go/service/s3"
// )

// // relevant type definitions:

// // type UploadedImageObject struct {
// // 	ImageUrl string
// // 	ID       int
// // }

// // look at the awsS3Client

// type mockS3Res struct {
// 	err error
// 	res s3.PutObjectOutput
// }

// type mockS3In struct {
// 	bucket string
// 	key    string
// }

// type MockS3Client struct {
// 	// some stuff
// 	vals map[mockS3In]mockS3Res
// 	seen map[mockS3In]bool
// }

// func NewMockS3Client() *MockS3Client {
// 	res := MockS3Client{
// 		vals: make(map[mockS3In]mockS3Res),
// 		seen: make(map[mockS3In]bool),
// 	}
// 	return &res
// }

// func (m MockS3Client) expect(key string, bucket string, out s3.PutObjectOutput, err error) {
// 	mk := mockS3In{key: key, bucket: bucket}
// 	m.vals[mk] = mockS3Res{err: err, res: out}
// }

// func (m MockS3Client) PutObject(put *s3.PutObjectInput) (s3.PutObjectOutput, error) {
// 	in := mockS3In{bucket: *put.Bucket, key: *put.Key}
// 	m.seen[in] = true
// 	res := m.vals[in]
// 	return res.res, res.err
// }

// // return true if all expected results were seen
// func (m MockS3Client) allSeen() bool {
// 	for k, _ := range m.vals {
// 		if !m.seen[k] {
// 			return false
// 		}
// 	}
// 	return true
// }

// func TestUploadImagesToS3(t *testing.T) {
// 	testCases := []struct {
// 		name         string
// 		bucket       string
// 		expectErr    error
// 		imagePaths   []string
// 		imageContent [][]byte
// 	}{
// 		{
// 			name:         "test 1",
// 			imagePaths:   []string{"image001.jpg", "image002.jpg", "image003.jpg"},
// 			imageContent: [][]byte{{}, {}, {}},
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			// todo create mock client
// 			mocks3c := NewMockS3Client()
// 			// todo set up expected calls to mock client - key,
// 			mocks3c.expect(tc.name, tc.bucket, nil, tc.expectErr)
// 			// create s3 uploader
// 			s3u := S3Uploader{s3c: mocks3c, bucketName: tc.bucket}

// 			// create a temporary directory to upload the images from
// 			temporaryDirectory, err := os.MkdirTemp(".", "test-download-images")
// 			if err != nil {
// 				t.Fatalf("failed to create a temporary directory: %v", err)
// 			}
// 			defer os.RemoveAll(temporaryDirectory)

// 			// create temporary images
// 			for index, imagePath := range testCase.imagePaths {
// 				// establish the filepath
// 				filePath := filepath.Join(temporaryDirectory, imagePath)

// 				// create the image file
// 				imageFile, err := os.Create(filePath)
// 				if err != nil {
// 					t.Fatalf("failed to create a temporary image: %v", err)
// 				}
// 				defer imageFile.Close()

// 				// get the image file content from the test case
// 				imageFileContent := testCase.imageContent[index]

// 				// write the image file content to the image file
// 				_, err = imageFile.Write(imageFileContent)
// 				if err != nil {
// 					t.Fatalf("failed to write content to the temporary image: %v", err)
// 				}
// 			}

// 			// call the UploadImagesToS3 function that we want to test against
// 			uploadedImageObjects, err := UploadImagesToS3(temporaryDirectory)
// 			if err != nil {
// 				t.Fatalf("error: UploadImagesToS3 failed: %v", err)
// 			}

// 			// temporary print to allow compilation
// 			fmt.Println(uploadedImageObjects)

// 			if !mocks3c.allSeen() {
// 				// test fail
// 			}

// 			// TODO: test against downloadedImageObjects
// 		})
// 	}
// }
