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

// define a type of struct
// (!) ` ` are struct tags, it provides metadata which can be used by Go packages like encoding/json, otherwise the unmarshal will use the field names
type Player struct {
	Name      string `json:"name"`
	HighScore int32  `json:"high_score"`
}

// [1] JSON
func parseJSON(file []byte, dataSlice *[]Player) error {
	// try to "unmarshal" the data from the json file into the jsonDataSlice variable
	err := json.Unmarshal(file, &dataSlice)

	// if we fail to "unmarshal" the json file then error
	if err != nil {
		return errors.New("[3] could not parse any data in JSON format")
	}

	// if the data slice is empty
	if len(*dataSlice) == 0 {
		return errors.New("[3] could not parse any data in JSON format")
	}

	return nil
}

// [2] Repeated JSON
func parseRepeatedJSON(file []byte, dataSlice *[]Player) error {
	// convert the file to a string and then split it on a newline character
	lines := strings.Split(string(file), "\n")

	// if no lines were parsed then error
	if len(lines) < 1 {
		return errors.New("[3] could not parse any data in repeated JSON format")
	}

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
		*dataSlice = append(*dataSlice, player)
	}

	// if the data slice is empty
	if len(*dataSlice) == 0 {
		return errors.New("[3] could not parse any data in repeated JSON format")
	}

	return nil
}

// [3] CSV
func parseCSV(file []byte, dataSlice *[]Player) error {
	// because of the way we have tried to use file []byte as a parameter..
	// it means we have to use bytes.NewReader on the file... this is JANKY
	reader := csv.NewReader(bytes.NewReader(file))

	// try to read the lines from the csv file
	lines, err := reader.ReadAll()

	// if we can't read the lines from the csv file then error
	if err != nil {
		return errors.New("[3] could not parse any data in CSV format")
	}

	if len(lines) == 0 {
		return errors.New("[3] could not parse any data in CSV format")
	}

	// (!!!) if we try to read one of the binary files it gives us this: [[��] [AyaPrisha����CharlieMargot]]
	// which is NOT CSV but passes beyond the error check and then causes a panic

	// loop over each line
	for _, line := range lines {
		// get the player name
		name := line[0]

		// the player has no name so skip
		if len(name) < 1 {
			continue
		}

		//  parse the string to get the score ((!) this is of type int not int32)
		highScore, err := strconv.Atoi(line[1])

		// if the player has no high score then skip
		if err != nil {
			continue
		}

		// add the player to the dataSlice
		*dataSlice = append(*dataSlice, Player{Name: name, HighScore: int32(highScore)})
	}

	// if the data slice is empty
	if len(*dataSlice) == 0 {
		return errors.New("[3] could not parse any data in CSV format")
	}

	return nil
}

// [4] BINARY
func parseBinary(file []byte, dataSlice *[]Player) error {
	// each record contains exactly four bytes
	// representing the score as a signed 32-bit integer (in the above endian format)
	// then the name of the player stored in UTF-8
	// which may not contain a null character, followed by a null terminating character

	// --- DEBUG PRINT TABLE
	// for index := 0; index < len(file); index++ {
	// 	// print out the bytes in the slice in various formats
	// 	fmt.Printf("slice index: %2d | Binary: %08b | Bytes (as an Integer): %3v | Hexadecimal: %2x | Decimal: %3d | ASCII : %c\n", index, file[index], file[index], file[index], file[index], file[index])
	// }

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

	// initialise two flags (explicitly two, yes could be achieved with one)
	var bigEndian bool
	var littleEndian bool

	// check the first two bytes of the file (deliberately using byte() to be explicit)
	// if they are "fe ff" it is big endian
	// if they are "ff fe" is is little endian
	if file[0] == byte(0xfe) && file[1] == byte(0xff) {
		bigEndian = true
	} else if file[0] == byte(0xff) && file[1] == byte(0xfe) {
		littleEndian = true
	} else {
		return errors.New("[3] could not establish Endianness of the binary file")
	}

	// set the start pointer to 2 to skip the first two bytes (which contain the Byte Order Mark BOM)
	// then loop over the byte slice
	for startIndex := 2; startIndex < len(file); {
		// initialise a variable to store the high score
		var highScore int32

		// parse the high score (signed 32-bit integer) based on the Endianness
		// (!) but why are we using Uint32 method (when we know its "signed") and then converting it to int32 (presumably "signed").. where is the native xEndian.Int32 method?
		if bigEndian {
			highScore = int32(binary.BigEndian.Uint32(file[startIndex : startIndex+4]))
		} else if littleEndian {
			highScore = int32(binary.LittleEndian.Uint32(file[startIndex : startIndex+4]))
		}

		// increment the start pointer by 4
		startIndex += 4

		// set the end pointer to the start pointer
		endIndex := startIndex

		// while the value is not a null terminator (deliberately using byte() to be explicit)
		// walk the end pointer along, to establish the end of the string
		for file[endIndex] != byte(0x00) {
			endIndex++
		}

		// parse the player name (UTF-8)
		name := string(file[startIndex:endIndex])

		// add the player to the dataSlice
		*dataSlice = append(*dataSlice, Player{Name: name, HighScore: highScore})

		// adjust the start pointer to skip past the null terminator
		startIndex = endIndex + 1
	}

	return nil
}

func attemptToParse(file []byte) ([]Player, error) {
	// initialise a slice of Player structs
	var dataSlice []Player

	// doing this sequentially isn't great... how to establish what type of file it is before attempting to parse it(?)

	err := parseJSON(file, &dataSlice)
	if err != nil {
		fmt.Println(err) // temporary
	}
	if err == nil {
		return dataSlice, nil
	}

	err = parseRepeatedJSON(file, &dataSlice)
	if err != nil {
		fmt.Println(err) // temporary
	}
	if err == nil {
		return dataSlice, nil
	}

	err = parseCSV(file, &dataSlice)
	if err != nil {
		fmt.Println(err) // temporary
	}
	if err == nil {
		return dataSlice, nil
	}

	err = parseBinary(file, &dataSlice)
	if err != nil {
		fmt.Println(err) // temporary
	}
	if err == nil {
		return dataSlice, nil
	}

	return nil, fmt.Errorf("[2] could not parse the file: %w", err)
}

func parseFile(filename string) ([]Player, error) {
	// try to read the file
	file, err := os.ReadFile(filename)

	// if we can't read the file then error
	if err != nil {
		return nil, errors.New("[1] error reading the file")
	}

	dataSlice, err := attemptToParse(file)

	// if we can't parse the file then error
	if err != nil {
		return nil, errors.New("[1] error parsing data from the file")
	}

	// if the dataSlice is empty then error
	if len(dataSlice) == 0 {
		return nil, errors.New("[1] the file likely contains no data")
	}

	return dataSlice, nil
}

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
	examplesDirectory := "examples"

	// get a list of files from the examples directory
	exampleFiles, err := os.ReadDir(examplesDirectory)

	if err != nil {
		fmt.Fprintf(os.Stderr, "[0] cannot find any example files: %v\n", err)
		os.Exit(2)
	}

	// loop over the example files and run parseFile()
	for _, exampleFile := range exampleFiles {
		// dynamically build the file path
		filePath := fmt.Sprintf("%s/%s", examplesDirectory, exampleFile.Name())

		// try to parse the file
		dataSlice, err := parseFile(filePath)

		// if we can't parse the parse then error
		if err != nil {
			fmt.Fprintf(os.Stderr, "[0] error parsing data from the file: %v\n", err)
			os.Exit(2)
		}

		// try to get the highest/lowest scoring players
		playerWithHighestScore, playerWithLowestScore, err := getHighestLowestScorePlayers(dataSlice)

		// if we can't get the highest/lowest scoring players then error
		if err != nil {
			fmt.Fprintf(os.Stderr, "[0] error getting the highest/lowest scoring players")
			os.Exit(2)
		}

		// print the highest/lowest scoring players
		fmt.Fprintf(os.Stdout, "%s had the highest score! (%d)\n%s had the lowest score! (%d)\n", playerWithHighestScore.Name, playerWithHighestScore.HighScore, playerWithLowestScore.Name, playerWithLowestScore.HighScore)
	}
}
