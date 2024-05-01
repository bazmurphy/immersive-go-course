package main

import (
	"flag"
	"log"
	"os"
)

func main() {
	log.Printf("⏳ batch-processing started...")

	inputCSVFilepath := flag.String("input", "", "A path to a CSV file to be processed")
	outputCSVFilepath := flag.String("output", "", "A path to a directory to output the resulting CSV")

	flag.Parse()

	if *inputCSVFilepath == "" || *outputCSVFilepath == "" {
		log.Fatalf("🔴 error: failed to provide both an '--input' and '--output' flag\n")
	}

	inputCSVRows, err := ReadInputCSV(*inputCSVFilepath)
	if err != nil {
		log.Fatalf(err.Error())
	}

	parsedImageUrlObjects, err := ParseImageUrls(inputCSVRows)
	if err != nil {
		log.Fatalf(err.Error())
	}

	temporaryDownloadsDirectory, err := os.MkdirTemp("", "downloads-*")
	if err != nil {
		log.Fatalf("🔴 error: failed to create a temporary downloads directory: %v", err)
	}
	defer os.RemoveAll(temporaryDownloadsDirectory)

	downloadedImageObjects, err := DownloadImages(parsedImageUrlObjects, temporaryDownloadsDirectory)
	if err != nil {
		log.Fatalf(err.Error())
	}

	temporaryGrayscaleDirectory, err := os.MkdirTemp("", "grayscale-*")
	if err != nil {
		log.Fatalf("🔴 error: failed to create a temporary grayscale directory: %v", err)
	}
	defer os.RemoveAll(temporaryGrayscaleDirectory)

	convertedImageObjects, err := ConvertImagesToGrayscale(temporaryDownloadsDirectory, temporaryGrayscaleDirectory)
	if err != nil {
		log.Fatalf(err.Error())
	}

	uploadedImageObjects, err := UploadImagesToS3(temporaryGrayscaleDirectory)
	if err != nil {
		log.Fatalf(err.Error())
	}

	dataMap := GenerateDataMap(parsedImageUrlObjects, downloadedImageObjects, convertedImageObjects, uploadedImageObjects)

	err = WriteOutputCSV(*outputCSVFilepath, dataMap)
	if err != nil {
		log.Fatalf(err.Error())
	}

	log.Printf("✅ batch-processing complete!\n")
}

// baz@baz-pc:/media/baz/external/coding/immersive-go-course/projects/batch-processing$ go run . --input=inputs/unsplash.csv --output=outputs/unsplash.csv
// 2024/05/01 13:41:39 ⏳ batch-processing started...
// 2024/05/01 13:41:39 🔵 attempting: to read rows from the input csv...
// 2024/05/01 13:41:39 🟢 success: read 4 rows from the input csv
// 2024/05/01 13:41:39 🔵 attempting: to parse image urls from the input csv...
// 2024/05/01 13:41:39 🟢 success: parsed 3 image urls from the input csv
// 2024/05/01 13:41:39 🔵 attempting: to download the images from the image urls...
// 2024/05/01 13:41:39 🟢 success: downloaded 3 images from the image urls
// 2024/05/01 13:41:39 🔵 attempting: to convert images to grayscale...
// 2024/05/01 13:41:39 🔵 attempting: to convert "/tmp/downloads-3055009226/photo-1506815444479-bfdb1e96c566.jpg" to "/tmp/grayscale-3768060023/photo-1506815444479-bfdb1e96c566-grayscale.jpg"
// 2024/05/01 13:41:39 🟢 success: converted "/tmp/downloads-3055009226/photo-1506815444479-bfdb1e96c566.jpg" to "/tmp/grayscale-3768060023/photo-1506815444479-bfdb1e96c566-grayscale.jpg"
// 2024/05/01 13:41:39 🔵 attempting: to convert "/tmp/downloads-3055009226/photo-1533738363-b7f9aef128ce.jpg" to "/tmp/grayscale-3768060023/photo-1533738363-b7f9aef128ce-grayscale.jpg"
// 2024/05/01 13:41:39 🟢 success: converted "/tmp/downloads-3055009226/photo-1533738363-b7f9aef128ce.jpg" to "/tmp/grayscale-3768060023/photo-1533738363-b7f9aef128ce-grayscale.jpg"
// 2024/05/01 13:41:39 🔵 attempting: to convert "/tmp/downloads-3055009226/photo-1540979388789-6cee28a1cdc9.jpg" to "/tmp/grayscale-3768060023/photo-1540979388789-6cee28a1cdc9-grayscale.jpg"
// 2024/05/01 13:41:39 🟢 success: converted "/tmp/downloads-3055009226/photo-1540979388789-6cee28a1cdc9.jpg" to "/tmp/grayscale-3768060023/photo-1540979388789-6cee28a1cdc9-grayscale.jpg"
// 2024/05/01 13:41:39 🟢 success: converted 3 images to grayscale
// 2024/05/01 13:41:39 🔵 attempting: to upload the images to AWS S3...
// 2024/05/01 13:41:39 🟢 success: uploaded image to AWS S3: https://bazmurphy-batch-processing.s3.eu-west-2.amazonaws.com/photo-1506815444479-bfdb1e96c566-grayscale.jpg
// 2024/05/01 13:41:39 🟢 success: uploaded image to AWS S3: https://bazmurphy-batch-processing.s3.eu-west-2.amazonaws.com/photo-1533738363-b7f9aef128ce-grayscale.jpg
// 2024/05/01 13:41:39 🟢 success: uploaded image to AWS S3: https://bazmurphy-batch-processing.s3.eu-west-2.amazonaws.com/photo-1540979388789-6cee28a1cdc9-grayscale.jpg
// 2024/05/01 13:41:39 🟢 success: uploaded 3 images to AWS S3
// 2024/05/01 13:41:39 🔵 attempting: to create and write the output csv...
// 2024/05/01 13:41:39 🟢 success: the output csv file was successfully created at: outputs/unsplash.csv
// 2024/05/01 13:41:39 ✅ batch-processing complete!
// baz@baz-pc:/media/baz/external/coding/immersive-go-course/projects/batch-processing$
