package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/url"
	"os"
)

func ReadInputCSV(inputCSVFilepath string) ([][]string, error) {
	log.Println("🔵 attempting: to read rows from the input csv...")

	inputCSVFile, err := os.Open(inputCSVFilepath)
	if err != nil {
		return nil, fmt.Errorf("🔴 error: failed to open the input csv file: %v", err)
	}
	defer inputCSVFile.Close()

	reader := csv.NewReader(inputCSVFile)

	inputCSVRows, err := reader.ReadAll()
	// TODO: should we be using `Read` and a loop here... or is `ReadAll` ok?
	if err != nil {
		// TODO: is this really fatal or is there another way to possibly continue? (see the `Read` point above)
		return nil, fmt.Errorf("🔴 error: failed to read all the input csv rows: %v", err)
	}

	if len(inputCSVRows) == 0 {
		return nil, fmt.Errorf("🔴 error: the input csv file is empty")
	}

	log.Printf("🟢 success: read %d rows from the input csv", len(inputCSVRows))

	return inputCSVRows, nil
}

type ParsedImageUrlObject struct {
	ImageUrl string
	ID       int
}

func ParseImageUrls(inputCSVRows [][]string) ([]ParsedImageUrlObject, error) {
	log.Println("🔵 attempting: to parse image urls from the input csv...")

	var parsedImageUrlObjects []ParsedImageUrlObject

	for rowNumber, row := range inputCSVRows {
		// check row 0 for the correct single 'url' column heading
		if rowNumber == 0 && len(row) != 1 && row[0] != "url" {
			return nil, fmt.Errorf("🔴 error: the input csv has more than a single 'url' column heading")
		}

		// start parsing image urls from row 1 onwards
		if rowNumber > 0 {
			// TODO: is this the right way to check if the row is "empty"?
			// TODO: what about if it not empty but is a string of rubbish?
			if row[0] == "" {
				log.Printf("🟠 warn: no url found on row %d of the input csv\n", rowNumber)
				// TODO: should we really continue if there is no image url on this row
				continue
			}

			// TODO: this will panic if no element is found at index position 0
			imageUrl := row[0]

			_, err := url.Parse(imageUrl)
			if err != nil {
				log.Printf("🟠 warn: invalid url %s on row %d of the input csv\n", imageUrl, rowNumber)
				continue
			}

			imageUrlObject := ParsedImageUrlObject{
				ImageUrl: imageUrl,
				ID:       rowNumber,
			}

			parsedImageUrlObjects = append(parsedImageUrlObjects, imageUrlObject)
		}
	}

	log.Printf("🟢 success: parsed %d image urls from the input csv", len(parsedImageUrlObjects))

	return parsedImageUrlObjects, nil
}

func WriteOutputCSV(outputCSVFilepath string, data map[int][]string) error {
	log.Println("🔵 attempting: to create and write the output csv...")

	outputCSVFile, err := os.Create(outputCSVFilepath)
	if err != nil {
		return fmt.Errorf("🔴 error: failed to create the output csv file: %v", err)
	}
	defer outputCSVFile.Close()

	writer := csv.NewWriter(outputCSVFile)
	defer writer.Flush()

	outputCSVColumnHeadings := []string{"url", "input", "output", "s3url"}
	err = writer.Write(outputCSVColumnHeadings)
	if err != nil {
		return fmt.Errorf("🔴 error: failed to write column headings to the output csv file: %v", err)
	}

	for dataID, dataValues := range data {
		err := writer.Write(dataValues)
		if err != nil {
			return fmt.Errorf("🔴 error: failed to write row ID%d the output csv file: %v", dataID, err)
		}
	}

	log.Printf("🟢 success: the output csv file was successfully created at: %s\n", outputCSVFilepath)

	return nil
}
