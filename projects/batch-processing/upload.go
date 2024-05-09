package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/joho/godotenv"
)

type UploadedImageObject struct {
	ImageUrl string
	ID       int
}

func UploadImagesToS3(temporaryGrayscaleDirectory string) ([]UploadedImageObject, error) {
	log.Println("🔵 attempting: to upload the images to AWS S3...")

	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("🔴 error: cannot load the .env file")
	}

	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion == "" {
		return nil, fmt.Errorf("🔴 error: cannot get the AWS_REGION environment variable")
	}

	awsAccessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	if awsAccessKeyID == "" {
		return nil, fmt.Errorf("🔴 error: cannot get the AWS_ACCESS_KEY_ID environment variable")
	}

	awsSecretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if awsSecretAccessKey == "" {
		return nil, fmt.Errorf("🔴 error: cannot get the AWS_SECRET_ACCESS_KEY environment variable")
	}

	awsS3BucketName := os.Getenv("AWS_S3_BUCKET")
	if awsS3BucketName == "" {
		return nil, fmt.Errorf("🔴 error: cannot get the AWS_S3_BUCKET environment variable")
	}

	awsSession := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region:      aws.String(awsRegion),
			Credentials: credentials.NewStaticCredentials(awsAccessKeyID, awsSecretAccessKey, ""),
		},
	}))

	awsS3Client := s3.New(awsSession)

	temporaryGrayscaleFiles, err := os.ReadDir(temporaryGrayscaleDirectory)
	if err != nil {
		return nil, fmt.Errorf("🔴 error: failed to read files from the temporary grayscale directory: %v", err)
	}

	var uploadedImageObjects []UploadedImageObject

	for index, file := range temporaryGrayscaleFiles {
		// ignore directories
		if file.IsDir() {
			continue
		}

		filepath := filepath.Join(temporaryGrayscaleDirectory, file.Name())

		openedFile, err := os.Open(filepath)
		if err != nil {
			// TODO: do I actually want this to be fatal... or just skip this file and continue to the next file...
			return nil, fmt.Errorf("🔴 error: failed to open the file %s from the temporary grayscale directory: %v", file, err)
		}

		defer openedFile.Close()

		awsS3Key := file.Name()

		// TODO: should I implement retry logic here in case the upload attempt fails for some network reason
		// USE: awsS3Client.PutObjectWithContext()
		_, err = awsS3Client.PutObject(&s3.PutObjectInput{
			Bucket: aws.String(awsS3BucketName),
			Key:    aws.String(awsS3Key),
			Body:   openedFile,
		})

		if err != nil {
			// TODO: need to implement some sort of retry logic here (or is it better to use CONTEXT per above^)
			// TODO: should this continue or be fatal?... maybe just give up on uploading this file and try the next...?
			// TODO: this will currently proceed even if we failed to upload the file
			log.Printf("🔴 failed to upload the file %s to the aws s3 bucket: %v\n", file.Name(), err)
			continue
		}

		objectURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", awsS3BucketName, awsRegion, awsS3Key)

		// TODO: how to move this out of here (single responsibility principle)
		// [STEP 5] CSV APPENDING LOGIC
		// outputCSVRows[index+1] = append(outputCSVRows[index+1], objectURL)

		uploadedImageObject := UploadedImageObject{
			ImageUrl: objectURL,
			ID:       index + 1,
		}

		uploadedImageObjects = append(uploadedImageObjects, uploadedImageObject)

		log.Printf("🟢 success: uploaded image to AWS S3: %s\n", objectURL)
	}

	log.Printf("🟢 success: uploaded %d images to AWS S3", len(uploadedImageObjects))

	return uploadedImageObjects, nil
}
