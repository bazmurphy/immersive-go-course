package main

import (
	"bytes"
	"encoding/binary"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
	// try to "unmarshal" the data from the json file into the data slice variable
	err := json.Unmarshal(file, &dataSlice)

	// if we fail to "unmarshal" the json file then error
	if err != nil {
		return errors.New("[3] could not parse any data in JSON format")
	}

	return nil
}

// [2] Repeated JSON
func parseRepeatedJSON(file []byte, dataSlice *[]Player) error {
	// convert the file to a string and then split it on a newline character
	lines := strings.Split(string(file), "\n")

	// if no lines were parsed then error
	if len(lines) == 0 {
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

		// the player has no name so skip
		if len(name) == 0 {
			// warn the user of the program in some way
			fmt.Fprintf(os.Stderr, "could not parse a player name from csv file line %d", index+1)
			// should i continue or break and error
			continue
		}

		//  parse the string to get the score ((!) this is of type int not int32)
		highScore, err := strconv.Atoi(line[1])

		// if the player has no high score then skip
		if err != nil {
			// warn the user of the program in some way
			fmt.Fprintf(os.Stderr, "could not parse a player high score from csv file line %d", index+1)
			// should i continue or break and error
			continue
		}

		// add the player to the dataSlice
		*dataSlice = append(*dataSlice, Player{Name: name, HighScore: int32(highScore)})
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

	var byteOrder binary.ByteOrder

	// check the first two bytes of the file
	// if they are "fe ff" it is big endian
	// if they are "ff fe" is is little endian
	if bytes.Equal(file[:2], []byte{0xfe, 0xff}) {
		byteOrder = binary.BigEndian
	} else if bytes.Equal(file[:2], []byte{0xff, 0xfe}) {
		byteOrder = binary.LittleEndian
	} else {
		return errors.New("[3] expected file to start with either feff (big endian) or fffe (little endian)")
	}

	// set the start pointer to 2 to skip the first two bytes (which contain the Byte Order Mark BOM)
	// then loop over the byte slice
	for startIndex := 2; startIndex < len(file); {
		// parse the high score (signed 32-bit integer) based on the Endianness
		// (!) but why are we using Uint32 method (when we know its "signed") and then converting it to int32 (presumably "signed").. where is the native xEndian.Int32 method?
		highScore := int32(byteOrder.Uint32(file[startIndex : startIndex+4]))

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

	// doing this sequentially isn't great...
	// how to establish what type of file it is before attempting to parse it?

	err := parseJSON(file, &dataSlice)
	// if err != nil {
	// 	fmt.Println(err) // temporary
	// }
	if err == nil {
		return dataSlice, nil
	}

	err = parseRepeatedJSON(file, &dataSlice)
	// if err != nil {
	// 	fmt.Println(err) // temporary
	// }
	if err == nil {
		return dataSlice, nil
	}

	err = parseCSV(file, &dataSlice)
	// if err != nil {
	// 	fmt.Println(err) // temporary
	// }
	if err == nil {
		return dataSlice, nil
	}

	err = parseBinary(file, &dataSlice)
	// if err != nil {
	// 	fmt.Println(err) // temporary
	// }
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
		return nil, fmt.Errorf("[1] error reading the file: %w", err)
	}

	dataSlice, err := attemptToParse(file)

	// if we can't parse the file then error
	if err != nil {
		return nil, fmt.Errorf("[1] error parsing data from the file: %w", err)
	}

	return dataSlice, nil
}

func getHighestLowestScorePlayers(dataSlice []Player) (Player, Player, error) {
	// check if the slice of structs is empty
	if len(dataSlice) == 0 {
		return Player{}, Player{}, errors.New("[1] the parsed data contains no players")
	}

	// create two variables to hold the highest/lowest scoring players, and set them as the first player in the slice
	playerWithHighestScore := dataSlice[0]
	playerWithLowestScore := dataSlice[0]

	// loop over the slice of structs and establish the highest/lowest scoring players
	for _, player := range dataSlice {
		if player.HighScore > playerWithHighestScore.HighScore {
			playerWithHighestScore = player
		} else if player.HighScore < playerWithLowestScore.HighScore {
			playerWithLowestScore = player
		}
	}

	return playerWithHighestScore, playerWithLowestScore, nil
}

func parseFilesFromDirectory(directory string) error {
	// get a list of files from the directory
	files, err := os.ReadDir(directory)

	// if we couldn't read the directory
	if err != nil {
		return fmt.Errorf("[0] cannot find that directory: %w", err)
	}

	// loop over the files and try to parse each
	for _, file := range files {
		// optional message
		fmt.Fprintf(os.Stderr, "--> attempting to parse %s\n", file.Name())

		// dynamically build the file path
		filePath := filepath.Join(directory, file.Name())

		// try to parse the file
		dataSlice, err := parseFile(filePath)

		// if we can't parse the file then error
		if err != nil {
			return fmt.Errorf("[0] error parsing data from the file: %w", err)
		}

		// try to get the highest/lowest scoring players
		playerWithHighestScore, playerWithLowestScore, err := getHighestLowestScorePlayers(dataSlice)

		// if we can't get the highest/lowest scoring players then error
		if err != nil {
			return fmt.Errorf("[0] error getting the highest/lowest scoring players: %w", err)
		}

		// print the highest/lowest scoring players
		fmt.Fprintf(os.Stdout, "%s had the highest score! (%d)\n%s had the lowest score! (%d)\n", playerWithHighestScore.Name, playerWithHighestScore.HighScore, playerWithLowestScore.Name, playerWithLowestScore.HighScore)
	}

	return nil
}

func main() {
	err := parseFilesFromDirectory("examples")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(2)
	}
}
