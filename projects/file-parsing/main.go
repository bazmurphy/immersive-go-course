package main

import (
	"bytes"
	"encoding/binary"
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

// [3] CSV
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

	if len(lines) > 0 {
		return nil, errors.New("[2] no lines found in the file")
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

// BIG ENDIAN (starts with fe ff) (00 at the end is the null terminator)
// data is displayed in groups of 2 bytes
// each group representing a hexadecimal value
// byte order where most significant byte is stored first

// 0000000 [fe ff] [00 00 00 0a] 41 79 61 00 00 00 00 1e 50 72
// 0000020 69 73 68 61 00 ff ff ff ff 43 68 61 72 6c 69 65
// 0000040 00 00 00 00 19 4d 61 72 67 6f 74 [00]
// 0000054

// LITTLE ENDIAN (starts with ff fe) (00 at the end is the null terminator)
// data is displayed in groups of 2 bytes
// each group representing a hexadecimal value
// byte order where least significant byte is stored first

// 0000000 [ff fe] [0a 00 00 00] 41 79 61 00 1e 00 00 00 50 72
// 0000020 69 73 68 61 00 ff ff ff ff 43 68 61 72 6c 69 65
// 0000040 00 19 00 00 00 4d 61 72 67 6f 74 [00]
// 0000054

// [4] BINARY
func parseBinary(file []byte) ([]Player, error) {
	var bigEndian bool
	var littleEndian bool
	fmt.Println(bigEndian, littleEndian) // temporary

	// check the first two bytes of the file
	// if they are "fe ff" it is big endian
	// if they are "ff fe" is is little endian
	if file[0] == 0xfe && file[1] == 0xff {
		fmt.Println("Big Endian")
		bigEndian = true
	} else if file[0] == 0xff && file[1] == 0xfe {
		fmt.Println("Little Endian")
		littleEndian = true
	} else {
		fmt.Println("could not establish Endianness")
	}

	// each record contains exactly four bytes
	// representing the score as a signed 32-bit integer (in the above endian format)
	// then the name of the player stored in UTF-8
	// which may not contain a null character, followed by a null terminating character

	var startIndex int
	var EndIndex int
	fmt.Println(startIndex, EndIndex) // temporary

	// start at index 2 (skipping the first two bytes)
	for index := 2; index < len(file); index++ {
		// print out the bytes in the slice in various formats
		fmt.Printf("slice index: %2d | Binary: %08b | Bytes (as an Integer): %3v | Hexadecimal: %2x | Decimal: %3d | ASCII : %c\n", index, file[index], file[index], file[index], file[index], file[index])
	}

	// test to extract the first score
	var firstScore int32

	if bigEndian {
		firstScore = int32(binary.BigEndian.Uint32(file[2:6]))
		fmt.Println("firstScore", firstScore)
	} else if littleEndian {
		firstScore = int32(binary.LittleEndian.Uint32(file[2:6]))
		fmt.Println("firstScore", firstScore)
	}

	return nil, nil
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

	dataSlice, err = parseBinary(file)
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
	// [4] Binary (Big Endian)
	file, err := os.ReadFile("examples/custom-binary-be.bin")
	// [4] Binary (Little Endian)
	// file, err := os.ReadFile("examples/custom-binary-le.bin")

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

	// if the dataSlice is empty then error
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
