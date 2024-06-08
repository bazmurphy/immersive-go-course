package main

import (
	"encoding/json"
	"errors"
	"strings"
)

type RepeatedJSONParser struct{}

func (p *RepeatedJSONParser) Parse(file []byte) ([]Record, error) {
	var records []Record

	// convert the file to a string and then split it on a newline character
	lines := strings.Split(string(file), "\n")

	// if no lines were parsed then error
	if len(lines) == 0 {
		return nil, errors.New("[3] could not parse any data in repeated JSON format")
	}

	// initialise a record struct to iteratively store each valid player
	var record Record

	// loop over the slice of strings
	for _, line := range lines {

		// if the line is empty keep going
		if len(line) == 0 {
			// TODO: maybe add a log here
			continue
		}

		// if the line starts with a # it is a comment so ignore it ((!)use single quotes for characters)
		if line[0] == '#' {
			// TODO: maybe add a log here
			continue
		}

		// try to parse the player struct from the line and store it in the record struct
		err := json.Unmarshal([]byte(line), &record)
		// if we can't parse a player struct from the line then error but keep going
		if err != nil {
			// TODO: silently erroring is bad...
			continue
		}

		// add the player struct to the slice of players
		records = append(records, record)
	}

	return records, nil
}
