package csv

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/url"
	"os"
)

func ReadInputCSV(inputCSVFilepath string) ([][]string, error) {
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

	return inputCSVRows, nil
}

func CreateOutputCSV(inputCSVRows [][]string) [][]string {
	// this is the important data structure that will receive the various output csv information as we proceed through
	outputCSVRows := make([][]string, len(inputCSVRows))
	outputCSVColumnHeadings := []string{"url", "input", "output", "s3url"}
	outputCSVRows[0] = outputCSVColumnHeadings
	return outputCSVRows
}

func ParseImageUrls(inputCSVRows [][]string, outputCSVRows [][]string) ([]string, error) {
	var imageUrls []string

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

			imageUrls = append(imageUrls, imageUrl)

			// TODO: how to move this out of here (single responsibility principle)
			// [STEP 2] CSV APPENDING LOGIC
			outputCSVRows[rowNumber] = append(outputCSVRows[rowNumber], imageUrl)
		}
	}

	return imageUrls, nil
}

func WriteOutputCSV(outputCSVFilepath string, outputCSVRows [][]string) error {
	outputCSVFile, err := os.Create(outputCSVFilepath)
	if err != nil {
		return fmt.Errorf("🔴 error: failed to create the output csv file: %v", err)
	}
	defer outputCSVFile.Close()

	writer := csv.NewWriter(outputCSVFile)

	err = writer.WriteAll(outputCSVRows)
	if err != nil {
		return fmt.Errorf("🔴 error: failed to write all the rows to the output csv file: %v", err)
	}

	log.Printf("🟢 an output csv file was successfully created at: %s\n", outputCSVFilepath)

	return nil
}
