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

func main() {
	inputCSVFilepath := flag.String("input", "", "A path to a CSV file to be processed")
	outputCSVFilepath := flag.String("output", "", "A path to a directory to output the resulting CSV")

	flag.Parse()

	if *inputCSVFilepath == "" {
		fmt.Printf("🔴 error: failed to provide the 'input' flag\n")
		flag.Usage()
		os.Exit(1)
	}

	csvFile, err := os.Open(*inputCSVFilepath)
	if err != nil {
		fmt.Printf("🔴 error: failed to open csv file: %v", err)
		// if we don't have a csv file we can't continue and have to exit
		os.Exit(1)
	}

	reader := csv.NewReader(csvFile)

	rows, err := reader.ReadAll()
	// TODO: should we be using Read and a loop here, or is ReadAll ok
	if err != nil {
		fmt.Printf("🔴 error: failed to read all the csv rows: %v", err)
		// TODO: is this really fatal or can we continue?
		os.Exit(1)
	}

	// (!) this is the important variable to store the output csv information as we go through
	var csvOutputRows [][]string

	csvOutputColumnHeadings := []string{"url", "input", "output", "s3url"}

	csvOutputRows = append(csvOutputRows, csvOutputColumnHeadings)

	// fmt.Println("DEBUG | [1] csvOutputRows", csvOutputRows)

	var imageUrls []string

	for rowNumber, row := range rows {
		// row 0 is the column headings, so skip it
		if rowNumber > 0 {
			if row[0] == "" {
				// TODO: is this the right way to check if the row is "empty"?
				// TODO: warn but continue?
				fmt.Printf("🟠 warn: no url found on row %d of the csv\n", rowNumber)
				continue
			}

			// TODO: this will panic if no element is found at index position 0
			imageUrl := row[0]

			// fmt.Println("DEBUG | imageUrl", imageUrl)

			_, err := url.Parse(imageUrl)
			if err != nil {
				fmt.Printf("🟠 warn: invalid url %s on row %d of the csv\n", imageUrl, rowNumber)
			}

			imageUrls = append(imageUrls, imageUrl)

			// (!!!) CSV APPENDING LOGIC
			csvOutputRows = append(csvOutputRows, []string{imageUrl})
		}
	}

	// fmt.Println("DEBUG | [2] csvOutputRows", csvOutputRows)

	temporaryDownloadsDirectory, err := os.MkdirTemp("", "downloads-*")
	if err != nil {
		fmt.Printf("🔴 error: failed to create a temporary downloads directory: %v", err)
		// if we can't make the temporary directory we have to exit
		os.Exit(1)
	}
	// fmt.Println("DEBUG | temporaryDirectory", temporaryDirectory)

	defer os.RemoveAll(temporaryDownloadsDirectory)

	for index, imageUrl := range imageUrls {
		// TODO: use context with timeout here (otherwise it can hang infinitely)
		// TODO: use some retry logic here (try 3 times and then give up)
		response, err := http.Get(imageUrl)
		if err != nil {
			fmt.Printf("🟠 warn: failed to get image url response %d/%d from url %s\n", index+1, len(imageUrls), imageUrl)
			break
		}
		// fmt.Println("DEBUG | response", response)

		defer response.Body.Close()

		contentType := response.Header.Get("Content-Type")
		// fmt.Println("DEBUG | contentType", contentType)

		mediaType, _, err := mime.ParseMediaType(contentType)
		if err != nil {
			fmt.Printf("🟠 warn: failed to parse media type from content type: %v", err)
			break
		}
		// fmt.Println("DEBUG | mediaType", mediaType)

		// fileExtensions, err := mime.ExtensionsByType(mediaType)
		// if err != nil {
		// 	fmt.Printf("warn: failed to get extensions from the mime type: %v", err)
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
			fmt.Println("🟠 warn: the image url was not a valid media type")
		}

		if fileExtension == "" {
			// if we reach here it is not safe to proceed,
			// because we will be copying the response body data into a file on the OS, which is dangerous if its malicious
			break
		}

		// fmt.Println("DEBUG | fileExtension", fileExtension)

		parsedUrl, err := url.Parse(imageUrl)
		if err != nil {
			fmt.Printf("🟠 warn: cannot parse the image name: %v", err)
			break
		}
		// fmt.Println("DEBUG | parsedUrl", parsedUrl)

		path := parsedUrl.Path
		// fmt.Println("DEBUG | path", path)

		compositeParts := strings.Split(path, "/")
		// fmt.Println("DEBUG | compositeParts", compositeParts)

		fileName := compositeParts[len(compositeParts)-1]
		// fmt.Println("DEBUG | fileName", fileName)

		outputFilepath := filepath.Join(temporaryDownloadsDirectory, fileName+fileExtension)
		// fmt.Println("DEBUG | outputFilepath", outputFilepath)

		// (!!!) CSV APPENDING LOGIC
		csvOutputRows[index+1] = append(csvOutputRows[index+1], outputFilepath)

		temporaryFile, err := os.Create(outputFilepath)
		if err != nil {
			fmt.Printf("🟠 warn: failed to create a temporary image file: %v", err)
			break
		}
		// fmt.Println("DEBUG | temporaryFile", temporaryFile)

		defer temporaryFile.Close()

		_, err = io.Copy(temporaryFile, response.Body)
		if err != nil {
			fmt.Printf("🟠 warn: failed to save image %d\n with url %s\n", index+1, imageUrl)
			break
		}
		// fmt.Println("DEBUG | bytesCopied", bytesCopied)
	}

	// fmt.Println("DEBUG | [3] csvOutputRows", csvOutputRows)

	temporaryGrayscaleDirectory, err := os.MkdirTemp("", "grayscale-*")
	if err != nil {
		fmt.Printf("🔴 error: failed to create a temporary grayscale directory: %v", err)
		os.Exit(1)
	}
	// fmt.Println("DEBUG | temporaryGrayscaleDirectory", temporaryGrayscaleDirectory)

	defer os.RemoveAll(temporaryGrayscaleDirectory)

	// Set up imagemagick
	imagick.Initialize()

	defer imagick.Terminate()

	// Build a Converter struct that will use imagick
	c := &Converter{
		cmd: imagick.ConvertImageCommand,
	}

	temporaryDownloadsFiles, err := os.ReadDir(temporaryDownloadsDirectory)
	if err != nil {
		fmt.Printf("🔴 error: failed to read files from the temporary downloads directory: %v", err)
		os.Exit(1)
	}
	// fmt.Println("DEBUG | temporaryDownloadsFiles", temporaryDownloadsFiles)

	for index, file := range temporaryDownloadsFiles {
		// fmt.Println("DEBUG | file", file)

		if file.IsDir() {
			// ignore directories
			log.Printf("🟠 warn: ignoring a directory...\n")
			continue
		}

		inputFilepath := filepath.Join(temporaryDownloadsDirectory, file.Name())
		// fmt.Println("DEBUG | inputFilepath", inputFilepath)

		fileExtension := filepath.Ext(file.Name())
		// fmt.Println("DEBUG | fileExtension", fileExtension)

		fileName := file.Name()[:len(file.Name())-len(fileExtension)]
		// fmt.Println("DEBUG | fileName", fileName)

		outputFilename := fmt.Sprintf("%s-grayscale%s", fileName, fileExtension)
		// fmt.Println("DEBUG | outputFilename", outputFilename)

		outputFilepath := filepath.Join(temporaryGrayscaleDirectory, outputFilename)
		// fmt.Println("DEBUG | outputFilepath", outputFilepath)

		// (!!!) CSV APPENDING LOGIC
		csvOutputRows[index+1] = append(csvOutputRows[index+1], outputFilepath)

		// Log what we're going to do
		log.Printf("🔵 processing: %q to %q\n", inputFilepath, outputFilepath)

		// Do the conversion!
		err := c.Grayscale(inputFilepath, outputFilepath)
		if err != nil {
			log.Printf("🟠 warn: failed to convert the image to grayscale: %v\n", err)
			continue
		}

		// Log what we did
		log.Printf("🟢 processed: %q to %q\n", inputFilepath, outputFilepath)
	}

	// fmt.Println("DEBUG | [4] csvOutputRows", csvOutputRows)

	sess := session.Must(session.NewSessionWithOptions(session.Options{SharedConfigState: session.SharedConfigEnable}))

	s3Client := s3.New(sess)

	temporaryGrayscaleFiles, err := os.ReadDir(temporaryGrayscaleDirectory)
	if err != nil {
		fmt.Printf("🔴 error: failed to read files from the temporary grayscale directory: %v", err)
		os.Exit(1)
	}
	// fmt.Println("DEBUG | temporaryGrayscaleFiles", temporaryGrayscaleFiles)

	for index, temporaryGrayscaleFile := range temporaryGrayscaleFiles {
		// fmt.Println("DEBUG | temporaryGrayscaleFile", temporaryGrayscaleFile)

		if temporaryGrayscaleFile.IsDir() {
			// ignore directories
			continue
		}

		filepath := filepath.Join(temporaryGrayscaleDirectory, temporaryGrayscaleFile.Name())

		openedFile, err := os.Open(filepath)
		if err != nil {
			fmt.Printf("🔴 error: failed to open the file %s from the temporary grayscale directory: %v", temporaryGrayscaleFile, err)
		}
		// fmt.Println("DEBUG | openedFile", openedFile)

		defer openedFile.Close()

		// s3BucketName := os.Getenv("AWS_S3_BUCKET")
		// TODO: Deal with dynamically loading this via Docker environment file

		s3BucketName := "bazmurphy-batch-processing"
		// fmt.Println("DEBUG | s3BucketName", s3BucketName)

		s3Key := temporaryGrayscaleFile.Name()

		_, err = s3Client.PutObject(&s3.PutObjectInput{
			Bucket: aws.String(s3BucketName),
			Key:    aws.String(s3Key),
			Body:   openedFile,
		})

		if err != nil {
			fmt.Printf("🔴 error: failed to upload the file %s to the aws s3 bucket: %v\n", temporaryGrayscaleFile, err)
			break
		}

		awsRegion := *sess.Config.Region
		// fmt.Println("DEBUG | awsRegion", awsRegion)

		objectURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s3BucketName, awsRegion, s3Key)
		// fmt.Println("DEBUG | objectURL", objectURL)

		log.Printf("🟢 uploaded file to aws s3: %s\n", objectURL)

		// (!!!) CSV APPENDING LOGIC
		csvOutputRows[index+1] = append(csvOutputRows[index+1], objectURL)
	}

	// fmt.Println("DEBUG | [5] csvOutputRows", csvOutputRows)

	outputCSVFile, err := os.Create(*outputCSVFilepath)
	if err != nil {
		fmt.Printf("🔴 error: failed to create the output CSV file: %v\n", err)
		os.Exit(1)
	}

	defer outputCSVFile.Close()

	writer := csv.NewWriter(outputCSVFile)

	err = writer.WriteAll(csvOutputRows)
	if err != nil {
		fmt.Printf("🔴 error: failed to write all the rows to the output CSV file: %v\n", err)
		os.Exit(1)
	}

	log.Printf("🟢 an output csv file was successfully created at: %s\n", *outputCSVFilepath)
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
