package main

import (
	"bytes"
	"encoding/binary"
	"errors"
)

func parseBinary(file []byte) ([]Record, error) {
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
		return nil, errors.New("[3] expected file to start with either feff (big endian) or fffe (little endian)")
	}

	var records []Record

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
		records = append(records, Record{Name: name, HighScore: highScore})

		// adjust the start pointer to skip past the null terminator
		startIndex = endIndex + 1
	}

	return records, nil
}
