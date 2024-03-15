package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// NOTE:
// there is repetition with the "dataSlice"
// could move to using a single var and pointers (later)

// define a type of struct
// (!) ` ` are struct tags, it provides metadata which can be used by Go packages like encoding/json, otherwise the unmarshal will use the field names
type Player struct {
	Name      string `json:"name"`
	HighScore int    `json:"high_score"`
}

// [1] JSON
func parseJSON(file []byte) ([]Player, error) {
	// initialise a slice of structs
	var dataSlice []Player

	// try to "unmarshal" the data from the json file into the jsonDataSlice variable
	err := json.Unmarshal(file, &dataSlice)

	// if we fail to "unmarshal" the json file then error
	if err != nil {
		return nil, errors.New("[2] the file is likely not [1] JSON")
	}

	// if the data slice is empty
	if len(dataSlice) == 0 {
		return nil, errors.New("[2] the file is likely empty")
	}

	return dataSlice, nil
}

// [2] Repeated JSON
func parseRepeatedJSON(file []byte) ([]Player, error) {
	// initialise a slice of structs
	var dataSlice []Player

	// convert the file to a string and then split it on a newline character
	lines := strings.Split(string(file), "\n")

	// (!) we need to error here if the lines is empty

	// initialise a player struct to iteratively store each valid player
	var player Player

	// loop over the slice of strings
	for _, line := range lines {

		// if the line is empty keep going
		if len(line) == 0 {
			continue
		}

		// if the line starts with a # it is a comment so ignore it ((!)use single quotes for characters)
		if line[0] == '#' {
			continue
		}

		// try to parse the player struct from the line and store it in the player struct
		err := json.Unmarshal([]byte(line), &player)

		// if we can't parse a player struct from the line then error but keep going
		if err != nil {
			continue
		}

		// add the player struct to the slice of players
		dataSlice = append(dataSlice, player)
	}

	// if the data slice is empty
	if len(dataSlice) == 0 {
		return nil, errors.New("[2] the file is likely empty")
	}

	return dataSlice, nil
}

// [2] CSV
func parseCSV(file []byte) ([]Player, error) {
	// initialise a slice of structs
	var dataSlice []Player

	// because of the way we have tried to use file []byte as a parameter..
	// it means we have to use bytes.NewReader on the file... this is JANKY
	reader := csv.NewReader(bytes.NewReader(file))

	// try to read the lines from the csv file
	lines, err := reader.ReadAll()

	// if we can't read the lines from the csv file then error
	if err != nil {
		return nil, errors.New("[2] the file is likely not [3] CSV")
	}

	// loop over each line
	for _, line := range lines {
		// intialise a player
		var player Player

		// set the player name
		player.Name = line[0]

		// convert the string to an integer
		// set the player high score
		player.HighScore, err = strconv.Atoi(line[1])

		if err != nil {
			// player has no high score value then skip
			continue
		}

		// add the player to the dataSlice
		dataSlice = append(dataSlice, player)
	}

	// if the data slice is empty
	if len(dataSlice) == 0 {
		return nil, errors.New("[2] the file is likely empty")
	}

	return dataSlice, nil
}

func parseFile(file []byte) ([]Player, error) {
	dataSlice, err := parseJSON(file)
	// if err != nil {
	// 	// fmt.Println(err)
	// }
	if err == nil {
		return dataSlice, nil
	}

	dataSlice, err = parseRepeatedJSON(file)
	// if err != nil {
	// 	// fmt.Println(err)
	// }
	if err == nil {
		return dataSlice, nil
	}

	dataSlice, err = parseCSV(file)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	if err == nil {
		return dataSlice, nil
	}

	return nil, fmt.Errorf("[1] could not parse the file: %w", err)
}

// (!) implement an error here
func getHighestLowestScorePlayers(jsonDataSlice []Player) (Player, Player, error) {
	// create two variables to hold the highest/lowest scoring players
	var playerWithHighestScore Player
	var playerWithLowestScore Player

	// loop over the slice of structs and establish the highest/lowest scoring players
	for _, player := range jsonDataSlice {
		if player.HighScore > playerWithHighestScore.HighScore {
			playerWithHighestScore = player
		} else if player.HighScore < playerWithLowestScore.HighScore {
			playerWithLowestScore = player
		}
	}

	return playerWithHighestScore, playerWithLowestScore, nil
}

func main() {
	// try to read the file
	// [1] JSON
	// file, err := os.ReadFile("examples/json.txt")
	// [2] RepeatedJSON
	// file, err := os.ReadFile("examples/repeated-json.txt")
	// [3] CSV
	// file, err := os.ReadFile("examples/data.csv")

	file, err := os.ReadFile("examples/empty.txt")

	// if we can't read the file then error
	if err != nil {
		fmt.Fprintf(os.Stderr, "[0] error reading the file: %v\n", err)
		os.Exit(2)
	}

	dataSlice, err := parseFile(file)

	// if we can't parse the file then error
	if err != nil {
		fmt.Fprintf(os.Stderr, "[0] error parsing data from the file: %v\n", err)
		os.Exit(2)
	}

	// if the dataSlice is empty
	if len(dataSlice) == 0 {
		fmt.Fprintf(os.Stderr, "[0] the file likely contains no data\n")
		os.Exit(2)
	}

	playerWithHighestScore, playerWithLowestScore, err := getHighestLowestScorePlayers(dataSlice)

	// if we can't get the highest/lowest scoring players then error
	if err != nil {
		fmt.Fprintf(os.Stderr, "[0] error getting the highest/lowest scoring players: %v\n", err)
		os.Exit(2)
	}

	// print the highest/lowest scoring players
	fmt.Fprintf(os.Stdout, "%s had the highest score! (%d)\n%s had the lowest score! (%d)\n", playerWithHighestScore.Name, playerWithHighestScore.HighScore, playerWithLowestScore.Name, playerWithLowestScore.HighScore)
}
