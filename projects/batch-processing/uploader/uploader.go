package uploader

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func UploadImagesToS3(temporaryGrayscaleDirectory string, outputCSVRows [][]string) error {
	temporaryGrayscaleFiles, err := os.ReadDir(temporaryGrayscaleDirectory)
	if err != nil {
		return fmt.Errorf("🔴 error: failed to read files from the temporary grayscale directory: %v", err)
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{SharedConfigState: session.SharedConfigEnable}))

	s3Client := s3.New(sess)

	for index, file := range temporaryGrayscaleFiles {
		// ignore directories
		if file.IsDir() {
			continue
		}

		filepath := filepath.Join(temporaryGrayscaleDirectory, file.Name())

		openedFile, err := os.Open(filepath)
		if err != nil {
			// TODO: do I actually want this to be fatal... or just skip this file and continue to the next file...
			return fmt.Errorf("🔴 error: failed to open the file %s from the temporary grayscale directory: %v", file, err)
		}

		defer openedFile.Close()

		// s3BucketName := os.Getenv("AWS_S3_BUCKET")
		// TODO: Deal with dynamically loading this via Docker using an environment file
		s3BucketName := "bazmurphy-batch-processing"
		s3Key := file.Name()

		// TODO: should I implement retry logic here in case the upload attempt fails for some network reason
		// USE: s3Client.PutObjectWithContext()
		_, err = s3Client.PutObject(&s3.PutObjectInput{
			Bucket: aws.String(s3BucketName),
			Key:    aws.String(s3Key),
			Body:   openedFile,
		})

		if err != nil {
			// TODO: need to implement some sort of retry logic here (or is it better to use CONTEXT per above^)
			// TODO: should this continue or be fatal?... maybe just give up on uploading this file and try the next...?
			// TODO: this will currently proceed even if we failed to upload the file
			log.Printf("🔴 failed to upload the file %s to the aws s3 bucket: %v\n", file.Name(), err)
			continue
		}

		awsRegion := *sess.Config.Region

		objectURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s3BucketName, awsRegion, s3Key)

		log.Printf("🟢 uploaded file to aws s3: %s\n", objectURL)

		// TODO: how to move this out of here (single responsibility principle)
		// [STEP 5] CSV APPENDING LOGIC
		outputCSVRows[index+1] = append(outputCSVRows[index+1], objectURL)
	}

	return nil
}
