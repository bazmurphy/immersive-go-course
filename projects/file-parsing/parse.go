package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
)

// (!) ` ` are struct tags
// they provide metadata which can be used by Go packages like encoding/json, otherwise the unmarshal will use the field names
type Record struct {
	Name      string `json:"name"`
	HighScore int32  `json:"high_score"`
}

func parseFile(filename string) ([]Record, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("[1] error reading the file: %w", err)
	}

	fileExtension := path.Ext(filename)
	fmt.Println("DEBUG | fileExtension", fileExtension)

	var records []Record

	switch fileExtension {
	case "json":
		records, err = parseJSON(file)
		if err != nil {
			return nil, fmt.Errorf("[2] could not parse the json file: %w", err)
		}
	case "txt":
		records, err = parseRepeatedJSON(file)
		if err != nil {
			return nil, fmt.Errorf("[2] could not parse the repeated json file: %w", err)
		}
	case "csv":
		records, err = parseCSV(file)
		if err != nil {
			return nil, fmt.Errorf("[2] could not parse the csv file: %w", err)
		}
	case "bin":
		records, err = parseBinary(file)
		if err != nil {
			return nil, fmt.Errorf("[2] could not parse the binary file: %w", err)
		}
	}

	return records, nil
}

func parseFilesFromDirectory(directory string) error {
	files, err := os.ReadDir(directory)
	if err != nil {
		// if we couldn't read the directory
		return fmt.Errorf("[0] cannot find that directory: %w", err)
	}

	// loop over the files and try to parse each
	for _, file := range files {
		// optional message
		fmt.Fprintf(os.Stderr, "--> attempting to parse %s\n", file.Name())

		// build the file path
		filePath := filepath.Join(directory, file.Name())

		// try to parse the file
		dataSlice, err := parseFile(filePath)
		if err != nil {
			// if we can't parse the file then error
			return fmt.Errorf("[0] error parsing data from the file: %w", err)
		}

		// try to get the highest/lowest scoring players
		playerWithHighestScore, playerWithLowestScore, err := getHighestLowestScoringPlayers(dataSlice)
		if err != nil {
			// if we can't get the highest/lowest scoring players then error
			return fmt.Errorf("[0] error getting the highest/lowest scoring players: %w", err)
		}

		// print the highest/lowest scoring players
		fmt.Fprintf(os.Stdout, "%s had the highest score! (%d)\n%s had the lowest score! (%d)\n", playerWithHighestScore.Name, playerWithHighestScore.HighScore, playerWithLowestScore.Name, playerWithLowestScore.HighScore)
	}

	return nil
}
