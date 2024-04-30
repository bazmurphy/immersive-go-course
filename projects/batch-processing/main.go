package main

import (
	"flag"
	"log"
	"os"

	"github.com/CodeYourFuture/immersive-go-course/batch-processing/converter"
	"github.com/CodeYourFuture/immersive-go-course/batch-processing/csv"
	"github.com/CodeYourFuture/immersive-go-course/batch-processing/downloader"
	"github.com/CodeYourFuture/immersive-go-course/batch-processing/uploader"
)

func main() {
	inputCSVFilepath := flag.String("input", "", "A path to a CSV file to be processed")
	outputCSVFilepath := flag.String("output", "", "A path to a directory to output the resulting CSV")

	flag.Parse()

	if *inputCSVFilepath == "" || *outputCSVFilepath == "" {
		flag.Usage()
		log.Fatalf("🔴 error: failed to provide both an 'input' and 'output' flag\n")
	}

	inputCSVRows, err := csv.ReadInputCSV(*inputCSVFilepath)
	if err != nil {
		// TODO: format this error
		log.Fatalf(err.Error())
	}

	outputCSVRows := csv.CreateOutputCSV(inputCSVRows)
	// fmt.Println("DEBUG | [STEP 1] outputCSVRows", outputCSVRows)

	imageUrls, err := csv.ParseImageUrls(inputCSVRows, outputCSVRows)
	if err != nil {
		// TODO: format this error
		log.Fatalf(err.Error())
	}
	// fmt.Println("DEBUG | [STEP 2] outputCSVRows", outputCSVRows)

	temporaryDownloadsDirectory, err := os.MkdirTemp("", "downloads-*")
	if err != nil {
		// if we can't make the temporary downloads directory we have to exit
		log.Fatalf("🔴 error: failed to create a temporary downloads directory: %v", err)
	}
	defer os.RemoveAll(temporaryDownloadsDirectory)

	err = downloader.DownloadImages(imageUrls, temporaryDownloadsDirectory, outputCSVRows)
	if err != nil {
		// TODO: format this error
		log.Fatalf(err.Error())
	}
	// fmt.Println("DEBUG | [STEP 3] outputCSVRows", outputCSVRows)

	temporaryGrayscaleDirectory, err := os.MkdirTemp("", "grayscale-*")
	if err != nil {
		// if we can't make the temporary grayscale directory we have to exit
		log.Fatalf("🔴 error: failed to create a temporary grayscale directory: %v", err)
	}
	defer os.RemoveAll(temporaryGrayscaleDirectory)

	err = converter.ConvertImagesToGrayscale(temporaryDownloadsDirectory, temporaryGrayscaleDirectory, outputCSVRows)
	if err != nil {
		// TODO: format this error
		log.Fatalf(err.Error())
	}
	// fmt.Println("DEBUG | [STEP 4] outputCSVRows", outputCSVRows)

	// TODO: handle an error (return value) here
	err = uploader.UploadImagesToS3(temporaryGrayscaleDirectory, outputCSVRows)
	if err != nil {
		// TODO: format this error
		log.Fatalf(err.Error())
	}
	// fmt.Println("DEBUG | [STEP 5] outputCSVRows", outputCSVRows)

	err = csv.WriteOutputCSV(*outputCSVFilepath, outputCSVRows)
	if err != nil {
		// TODO: format this error
		log.Fatalf(err.Error())
	}
}

// baz@baz-pc:/media/baz/external/coding/immersive-go-course/projects/batch-processing$ go run . --input=inputs/unsplash.csv --output=outputs/unsplash.csv
// 2024/04/30 14:45:20 🔵 processing: "/tmp/downloads-1171486988/photo-1506815444479-bfdb1e96c566.jpg" to "/tmp/grayscale-979035497/photo-1506815444479-bfdb1e96c566-grayscale.jpg"
// 2024/04/30 14:45:20 🟢 processed: "/tmp/downloads-1171486988/photo-1506815444479-bfdb1e96c566.jpg" to "/tmp/grayscale-979035497/photo-1506815444479-bfdb1e96c566-grayscale.jpg"
// 2024/04/30 14:45:20 🔵 processing: "/tmp/downloads-1171486988/photo-1533738363-b7f9aef128ce.jpg" to "/tmp/grayscale-979035497/photo-1533738363-b7f9aef128ce-grayscale.jpg"
// 2024/04/30 14:45:20 🟢 processed: "/tmp/downloads-1171486988/photo-1533738363-b7f9aef128ce.jpg" to "/tmp/grayscale-979035497/photo-1533738363-b7f9aef128ce-grayscale.jpg"
// 2024/04/30 14:45:20 🔵 processing: "/tmp/downloads-1171486988/photo-1540979388789-6cee28a1cdc9.jpg" to "/tmp/grayscale-979035497/photo-1540979388789-6cee28a1cdc9-grayscale.jpg"
// 2024/04/30 14:45:20 🟢 processed: "/tmp/downloads-1171486988/photo-1540979388789-6cee28a1cdc9.jpg" to "/tmp/grayscale-979035497/photo-1540979388789-6cee28a1cdc9-grayscale.jpg"
// 2024/04/30 14:45:21 🟢 uploaded file to aws s3: https://bazmurphy-batch-processing.s3.eu-west-2.amazonaws.com/photo-1506815444479-bfdb1e96c566-grayscale.jpg
// 2024/04/30 14:45:21 🟢 uploaded file to aws s3: https://bazmurphy-batch-processing.s3.eu-west-2.amazonaws.com/photo-1533738363-b7f9aef128ce-grayscale.jpg
// 2024/04/30 14:45:21 🟢 uploaded file to aws s3: https://bazmurphy-batch-processing.s3.eu-west-2.amazonaws.com/photo-1540979388789-6cee28a1cdc9-grayscale.jpg
// 2024/04/30 14:45:21 🟢 an output csv file was successfully created at: outputs/unsplash.csv
// baz@baz-pc:/media/baz/external/coding/immersive-go-course/projects/batch-processing$
