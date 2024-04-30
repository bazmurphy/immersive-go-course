package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"gopkg.in/gographics/imagick.v2/imagick"
)

type ConvertImageCommand func(args []string) (*imagick.ImageCommandResult, error)

type Converter struct {
	cmd ConvertImageCommand
}

func (c *Converter) Grayscale(inputFilepath string, outputFilepath string) error {
	// Convert the image to grayscale using imagemagick
	// We are directly calling the convert command
	_, err := c.cmd([]string{
		"convert", inputFilepath, "-set", "colorspace", "Gray", outputFilepath,
	})
	return err
}

// TODO: return an error (second return value) here
func readInputCSV(inputCSVFilepath string) [][]string {
	inputCSVFile, err := os.Open(inputCSVFilepath)
	if err != nil {
		// if we don't have a csv file we can't continue and have to exit
		log.Fatalf("🔴 error: failed to open the input csv file: %v", err)
	}
	defer inputCSVFile.Close()

	reader := csv.NewReader(inputCSVFile)

	inputCSVRows, err := reader.ReadAll()
	// TODO: should we be using `Read`` and a loop here, or is `ReadAll`` ok?
	if err != nil {
		// TODO: is this really fatal or could we possibly continue?
		log.Fatalf("🔴 error: failed to read all the input csv rows: %v", err)
	}

	return inputCSVRows
}

func parseImageUrls(inputCSVRows [][]string, outputCSVRows [][]string) []string {
	var imageUrls []string

	for rowNumber, row := range inputCSVRows {
		// check row 0 for the correct column heading
		if rowNumber == 0 && len(row) != 1 && row[0] != "url" {
			log.Fatalf("🔴 error: the input csv has more than a single 'url' column heading\n")
		}

		// then start parsing from row 1 onwards
		if rowNumber > 0 {
			// TODO: is this the right way to check if the row is "empty"?
			// TODO: what about if it not empty but is a string of rubbish?
			if row[0] == "" {
				log.Printf("🟠 warn: no url found on row %d of the input csv\n", rowNumber)
				// TODO: should we really continue if there is no image url
				continue
			}

			// TODO: this will panic if no element is found at index position 0
			imageUrl := row[0]

			_, err := url.Parse(imageUrl)
			if err != nil {
				log.Printf("🟠 warn: invalid url %s on row %d of the input csv\n", imageUrl, rowNumber)
				continue
			}

			imageUrls = append(imageUrls, imageUrl)

			// [STEP 2] CSV APPENDING LOGIC
			outputCSVRows[rowNumber] = append(outputCSVRows[rowNumber], imageUrl)
		}
	}

	return imageUrls
}

// TODO: return an error (return value) here
func downloadImages(imageUrls []string, temporaryDownloadsDirectory string, outputCSVRows [][]string) {
	for index, imageUrl := range imageUrls {
		// TODO: use context with timeout here (otherwise it can hang infinitely)
		// TODO: use some retry logic here (try 3 times and then give up)
		response, err := http.Get(imageUrl)
		if err != nil {
			log.Printf("🟠 warn: failed to get image url response from url %s\n", imageUrl)
			// TODO: think about if i want to continue or break here
			continue
		}
		defer response.Body.Close()

		// TODO: check the response status code and handle things appropriately
		if response.StatusCode != http.StatusOK {
			log.Printf("🟠 warn: response had status code %d", response.StatusCode)
		}

		contentType := response.Header.Get("Content-Type")

		mediaType, _, err := mime.ParseMediaType(contentType)
		if err != nil {
			log.Printf("🟠 warn: failed to parse media type from content type: %v", err)
			// TODO: think about if i want to continue or break here
			continue
		}

		// fileExtensions, err := mime.ExtensionsByType(mediaType)
		// if err != nil {
		// 	log.Printf("warn: failed to get extensions from the mime type: %v", err)
		// 	continue
		// }
		// fmt.Println("DEBUG | fileExtensions", fileExtensions)

		// TODO: the above^ gives an array of possible extensions
		// but in this case fileExtensions[0] is ".jpe" which is weird
		// so I can't rely on the first index value being the file extension I want/expect

		var fileExtension string

		switch mediaType {
		case "image/jpeg":
			fileExtension = ".jpg"
			// TODO: extend this with other cases for other image file types
			// (although ideally it would be better to rely on the method above)
		default:
			log.Printf("🟠 warn: the image url was not a valid media type")
		}

		if fileExtension == "" {
			// if we reach here it is not safe to proceed,
			// because we will be copying the response body data into a file on the OSwhich is dangerous if it is malicious
			// (also i can't break inside the switch/case)
			continue
		}

		parsedUrl, err := url.Parse(imageUrl)
		if err != nil {
			log.Printf("🟠 warn: cannot parse the image name: %v", err)
			// TODO: if we cannot parse the image we should skip... right?
			continue
		}

		path := parsedUrl.Path

		compositeParts := strings.Split(path, "/")

		fileName := compositeParts[len(compositeParts)-1]

		outputFilepath := filepath.Join(temporaryDownloadsDirectory, fileName+fileExtension)

		// TODO: how to move this out of here (single responsibility principle)
		// [STEP 3] CSV APPENDING LOGIC
		outputCSVRows[index+1] = append(outputCSVRows[index+1], outputFilepath)

		temporaryFile, err := os.Create(outputFilepath)
		if err != nil {
			log.Printf("🟠 warn: failed to create a temporary image file: %v", err)
			// TODO: think about if i want to continue or break here
			continue
		}
		defer temporaryFile.Close()

		_, err = io.Copy(temporaryFile, response.Body)
		if err != nil {
			log.Printf("🟠 warn: failed to save image %d\n with url %s\n", index+1, imageUrl)
			// TODO: think about if i want to continue or break here
			continue
		}
	}
}

// TODO: return an error (return value) here
func convertImagesToGrayscale(temporaryDownloadsDirectory, temporaryGrayscaleDirectory string, outputCSVRows [][]string) {
	imagick.Initialize()
	defer imagick.Terminate()

	c := &Converter{
		cmd: imagick.ConvertImageCommand,
	}

	temporaryDownloadsFiles, err := os.ReadDir(temporaryDownloadsDirectory)
	if err != nil {
		log.Fatalf("🔴 error: failed to read files from the temporary downloads directory: %v", err)
	}

	for index, file := range temporaryDownloadsFiles {

		if file.IsDir() {
			// ignore directories
			log.Printf("🟠 warn: ignoring a directory...\n")
			continue
		}

		inputFilepath := filepath.Join(temporaryDownloadsDirectory, file.Name())

		fileExtension := filepath.Ext(file.Name())

		fileName := file.Name()[:len(file.Name())-len(fileExtension)]

		outputFilename := fmt.Sprintf("%s-grayscale%s", fileName, fileExtension)

		outputFilepath := filepath.Join(temporaryGrayscaleDirectory, outputFilename)

		// TODO: how to move this out of here (single responsibility principle)
		// [STEP 4] CSV APPENDING LOGIC
		outputCSVRows[index+1] = append(outputCSVRows[index+1], outputFilepath)

		log.Printf("🔵 processing: %q to %q\n", inputFilepath, outputFilepath)

		err := c.Grayscale(inputFilepath, outputFilepath)
		if err != nil {
			log.Printf("🟠 warn: failed to convert the image to grayscale: %v\n", err)
			continue
		}

		// Log what we did
		log.Printf("🟢 processed: %q to %q\n", inputFilepath, outputFilepath)
	}
}

func uploadToS3(temporaryGrayscaleDirectory string, outputCSVRows [][]string) {
	temporaryGrayscaleFiles, err := os.ReadDir(temporaryGrayscaleDirectory)
	if err != nil {
		log.Fatalf("🔴 error: failed to read files from the temporary grayscale directory: %v", err)
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{SharedConfigState: session.SharedConfigEnable}))

	s3Client := s3.New(sess)

	for index, file := range temporaryGrayscaleFiles {

		if file.IsDir() {
			// ignore directories
			continue
		}

		filepath := filepath.Join(temporaryGrayscaleDirectory, file.Name())

		openedFile, err := os.Open(filepath)
		if err != nil {
			// TODO: do I actually want this to be fatal...? or rather skip this file and continue to the next..
			log.Fatalf("🔴 error: failed to open the file %s from the temporary grayscale directory: %v", file, err)
		}

		defer openedFile.Close()

		// s3BucketName := os.Getenv("AWS_S3_BUCKET")
		// TODO: Deal with dynamically loading this via Docker environment file
		s3BucketName := "bazmurphy-batch-processing"
		s3Key := file.Name()

		// TODO: should I implement retry logic here in case the upload attempt fails for some network reason
		_, err = s3Client.PutObject(&s3.PutObjectInput{
			Bucket: aws.String(s3BucketName),
			Key:    aws.String(s3Key),
			Body:   openedFile,
		})

		if err != nil {
			// TODO: should this continue or be fatal?... maybe just give up on this file and try the next...(?)
			log.Printf("🔴 error: failed to upload the file %s to the aws s3 bucket: %v\n", file, err)
			continue
		}

		awsRegion := *sess.Config.Region

		objectURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s3BucketName, awsRegion, s3Key)

		log.Printf("🟢 uploaded file to aws s3: %s\n", objectURL)

		// TODO: how to move this out of here (single responsibility principle)
		// [STEP 5] CSV APPENDING LOGIC
		outputCSVRows[index+1] = append(outputCSVRows[index+1], objectURL)
	}
}

func writeOutputCSV(outputCSVFilepath string, outputCSVRows [][]string) {
	outputCSVFile, err := os.Create(outputCSVFilepath)
	if err != nil {
		// if we can't create the output csv the program is useless
		log.Fatalf("🔴 error: failed to create the output csv file: %v\n", err)
	}
	defer outputCSVFile.Close()

	writer := csv.NewWriter(outputCSVFile)

	err = writer.WriteAll(outputCSVRows)
	if err != nil {
		log.Fatalf("🔴 error: failed to write all the rows to the output csv file: %v\n", err)
		os.Exit(1)
	}

	log.Printf("🟢 an output csv file was successfully created at: %s\n", outputCSVFilepath)
}

func main() {
	inputCSVFilepath := flag.String("input", "", "A path to a CSV file to be processed")
	outputCSVFilepath := flag.String("output", "", "A path to a directory to output the resulting CSV")

	flag.Parse()

	if *inputCSVFilepath == "" || *outputCSVFilepath == "" {
		flag.Usage()
		log.Fatalf("🔴 error: failed to provide both an 'input' and 'output' flag\n")
	}

	// TODO: handle an error (second return value) here
	inputCSVRows := readInputCSV(*inputCSVFilepath)

	// (!) this is the important data structure that will receive the various output csv information as we proceed
	outputCSVRows := make([][]string, len(inputCSVRows))
	// fmt.Println("DEBUG | [STEP 0] outputCSVRows", outputCSVRows)

	outputCSVColumnHeadings := []string{"url", "input", "output", "s3url"}
	outputCSVRows[0] = outputCSVColumnHeadings
	// fmt.Println("DEBUG | [STEP 1] outputCSVRows", outputCSVRows)

	imageUrls := parseImageUrls(inputCSVRows, outputCSVRows)
	// fmt.Println("DEBUG | [STEP 2] outputCSVRows", outputCSVRows)

	temporaryDownloadsDirectory, err := os.MkdirTemp("", "downloads-*")
	if err != nil {
		// if we can't make the temporary directory we have to exit
		log.Fatalf("🔴 error: failed to create a temporary downloads directory: %v", err)
	}
	defer os.RemoveAll(temporaryDownloadsDirectory)

	// TODO: handle an error (return value) here
	downloadImages(imageUrls, temporaryDownloadsDirectory, outputCSVRows)
	// fmt.Println("DEBUG | [STEP 3] outputCSVRows", outputCSVRows)

	temporaryGrayscaleDirectory, err := os.MkdirTemp("", "grayscale-*")
	if err != nil {
		log.Fatalf("🔴 error: failed to create a temporary grayscale directory: %v", err)
	}
	defer os.RemoveAll(temporaryGrayscaleDirectory)

	// TODO: handle an error (return value) here
	convertImagesToGrayscale(temporaryDownloadsDirectory, temporaryGrayscaleDirectory, outputCSVRows)
	// fmt.Println("DEBUG | [STEP 4] outputCSVRows", outputCSVRows)

	// TODO: handle an error (return value) here
	uploadToS3(temporaryGrayscaleDirectory, outputCSVRows)
	// fmt.Println("DEBUG | [STEP 5] outputCSVRows", outputCSVRows)

	writeOutputCSV(*outputCSVFilepath, outputCSVRows)
}

// baz@baz-pc:/media/baz/external/coding/immersive-go-course/projects/batch-processing$ go run . --input=inputs/unsplash.csv --output=outputs/unsplash.csv
// 2024/04/29 20:20:43 🔵 processing: "/tmp/downloads-868093528/photo-1506815444479-bfdb1e96c566.jpg" to "/tmp/grayscale-1115925022/photo-1506815444479-bfdb1e96c566-grayscale.jpg"
// 2024/04/29 20:20:43 🟢 processed: "/tmp/downloads-868093528/photo-1506815444479-bfdb1e96c566.jpg" to "/tmp/grayscale-1115925022/photo-1506815444479-bfdb1e96c566-grayscale.jpg"
// 2024/04/29 20:20:43 🔵 processing: "/tmp/downloads-868093528/photo-1533738363-b7f9aef128ce.jpg" to "/tmp/grayscale-1115925022/photo-1533738363-b7f9aef128ce-grayscale.jpg"
// 2024/04/29 20:20:43 🟢 processed: "/tmp/downloads-868093528/photo-1533738363-b7f9aef128ce.jpg" to "/tmp/grayscale-1115925022/photo-1533738363-b7f9aef128ce-grayscale.jpg"
// 2024/04/29 20:20:43 🔵 processing: "/tmp/downloads-868093528/photo-1540979388789-6cee28a1cdc9.jpg" to "/tmp/grayscale-1115925022/photo-1540979388789-6cee28a1cdc9-grayscale.jpg"
// 2024/04/29 20:20:43 🟢 processed: "/tmp/downloads-868093528/photo-1540979388789-6cee28a1cdc9.jpg" to "/tmp/grayscale-1115925022/photo-1540979388789-6cee28a1cdc9-grayscale.jpg"
// 2024/04/29 20:20:43 🟢 uploaded file to aws s3: https://bazmurphy-batch-processing.s3.eu-west-2.amazonaws.com/photo-1506815444479-bfdb1e96c566-grayscale.jpg
// 2024/04/29 20:20:43 🟢 uploaded file to aws s3: https://bazmurphy-batch-processing.s3.eu-west-2.amazonaws.com/photo-1533738363-b7f9aef128ce-grayscale.jpg
// 2024/04/29 20:20:43 🟢 uploaded file to aws s3: https://bazmurphy-batch-processing.s3.eu-west-2.amazonaws.com/photo-1540979388789-6cee28a1cdc9-grayscale.jpg
// 2024/04/29 20:20:43 🟢 an output csv file was successfully created at: outputs/unsplash.csv
// baz@baz-pc:/media/baz/external/coding/immersive-go-course/projects/batch-processing$
