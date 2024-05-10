package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// (!) ` ` are struct tags
// they provide metadata which can be used by Go packages like encoding/json, otherwise the unmarshal will use the field names
type Record struct {
	Name      string `json:"name"`
	HighScore int32  `json:"high_score"`
}

// define an interface for a FileParser
type FileParser interface {
	// to implement this interface the type must have the Parse method (with the exact function signature)
	Parse(file []byte) ([]Record, error)
}

func parseFile(filename string) ([]Record, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("[1] error reading the file: %w", err)
	}

	// get the file extension
	fileExtension := strings.ToLower(filepath.Ext(filename))
	// fileExtension := path.Ext(filename)
	fmt.Println("DEBUG | fileExtension", fileExtension)

	// instantiate the FileParser interface
	var fileParser FileParser

	switch fileExtension {
	// but in examples/ :
	// the json file is .txt
	// the repeated json file is .txt
	case ".json":
		fileParser = &JSONParser{}
	case ".csv":
		fileParser = &CSVParser{}
	case ".bin":
		fileParser = &BinaryParser{}
	default:
		return nil, fmt.Errorf("[1] error cannot handle this file extension")
	}

	var records []Record

	// try to parse the file
	records, err = fileParser.Parse(file)
	if err != nil {
		// if we can't parse the file then error
		return nil, fmt.Errorf("[1] error parsing data from the file: %w", err)
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
		records, err := parseFile(filePath)
		if err != nil {
			// if we can't parse the file then error
			return fmt.Errorf("[0] error parsing records from the file: %w", err)
		}

		// try to get the highest/lowest scoring players
		playerWithHighestScore, playerWithLowestScore, err := getHighestLowestScoringPlayers(records)
		if err != nil {
			// if we can't get the highest/lowest scoring players then error
			return fmt.Errorf("[0] error getting the highest/lowest scoring players: %w", err)
		}

		// print the highest/lowest scoring players
		fmt.Fprintf(os.Stdout, "%s had the highest score! (%d)\n%s had the lowest score! (%d)\n", playerWithHighestScore.Name, playerWithHighestScore.HighScore, playerWithLowestScore.Name, playerWithLowestScore.HighScore)
	}

	return nil
}
