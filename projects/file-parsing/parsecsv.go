package main

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strconv"
)

func parseCSV(file []byte) ([]Record, error) {
	// because of the way we have tried to use file []byte as a parameter..
	// it means we have to use bytes.NewReader on the file... this is JANKY
	reader := csv.NewReader(bytes.NewReader(file))

	// try to read the lines from the csv file
	// TODO: do this in chunks (!) (if it is a 100000000 line CSV file)
	lines, err := reader.ReadAll()
	if err != nil {
		return nil, errors.New("[3] could not parse any data in CSV format")
	}

	var records []Record

	// loop over each line
	for index, line := range lines {
		// (!!!) if we try to read one of the binary files it gives us this: [[��] [AyaPrisha����CharlieMargot]]
		// so we need to add an additional check here:

		// if the line does not have 2 values (name, highscore) then skip
		// ((!) this is brittle if a line has valid data but random extra values etc)
		if len(line) != 2 {
			continue
		}

		// get the player name...
		// (!) but this is assuming the name is at index position 0 which it may not be(?)
		name := line[0]

		if len(name) == 0 {
			// if the player has no name skip
			// warn the user of the program in some way
			fmt.Fprintf(os.Stderr, "could not parse a player name from csv file line %d", index+1)
			// should i continue or break and error
			continue
		}

		//  parse the string to get the score ((!) this is of type int not int32)
		highScore, err := strconv.Atoi(line[1])
		if err != nil {
			// if the player has no high score then skip
			// warn the user of the program in some way
			fmt.Fprintf(os.Stderr, "could not parse a player high score from csv file line %d", index+1)
			// should i continue or break and error
			continue
		}

		records = append(records, Record{Name: name, HighScore: int32(highScore)})
	}

	return records, nil
}
