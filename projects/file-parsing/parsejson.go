package main

import (
	"encoding/json"
	"errors"
)

type JSONParser struct{}

func (p *JSONParser) Parse(file []byte) ([]Record, error) {
	var records []Record

	// try to "unmarshal" the data from the json file into the data slice variable
	err := json.Unmarshal(file, &records)

	// if we fail to "unmarshal" the json file then error
	if err != nil {
		return nil, errors.New("[3] could not parse any data in JSON format")
	}

	return records, nil
}
