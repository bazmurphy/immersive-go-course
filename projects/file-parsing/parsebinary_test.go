package main

import (
	"errors"
	"reflect"
	"testing"
)

func TestParseBinary(t *testing.T) {
	testCases := []struct {
		name                string
		binary              []byte
		expectedRecords     []Record
		expectedRecordCount int
		expectedError       error
	}{
		{
			name: "valid big endian binary",
			binary: []byte{
				0xfe, 0xff, // big endian byte order mark
				0x00, 0x00, 0x00, 0x64, // high_score: 100
				0x70, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x31, 0x00, // name: "player1"
				0x00, 0x00, 0x00, 0x96, // high_score: 150
				0x70, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x32, 0x00, // name: "player2"
				0x00, 0x00, 0x00, 0x32, // high_score: 50
				0x70, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x33, 0x00, // name: "player3"
			},
			expectedError: nil,
			expectedRecords: []Record{
				{Name: "player1", HighScore: 100},
				{Name: "player2", HighScore: 150},
				{Name: "player3", HighScore: 50},
			},
			expectedRecordCount: 3,
		},
		{
			name: "valid little endian binary",
			binary: []byte{
				0xff, 0xfe, // little endian byte order mark
				0x64, 0x00, 0x00, 0x00, // high_score: 100
				0x70, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x31, 0x00, // name: "player1"
				0x96, 0x00, 0x00, 0x00, // high_score: 150
				0x70, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x32, 0x00, // name: "player2"
				0x32, 0x00, 0x00, 0x00, // high_score: 50
				0x70, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x33, 0x00, // name: "player3"
			},
			expectedError: nil,
			expectedRecords: []Record{
				{Name: "player1", HighScore: 100},
				{Name: "player2", HighScore: 150},
				{Name: "player3", HighScore: 50},
			},
			expectedRecordCount: 3,
		},
		{
			name:                "invalid byte order mark",
			binary:              []byte{0x01, 0x02, 0x03, 0x04},
			expectedError:       errors.New("[3] expected file to start with either feff (big endian) or fffe (little endian)"),
			expectedRecords:     nil,
			expectedRecordCount: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var binaryParser BinaryParser

			records, err := binaryParser.Parse(tc.binary)
			if err != nil {
				if tc.expectedError != nil {
					if err.Error() != tc.expectedError.Error() {
						t.Errorf("got error %v | expected error %v", err, tc.expectedError)
					}
				} else {
					t.Errorf("got error %v | expected error %v", err, tc.expectedError)
				}
			}

			if len(records) != tc.expectedRecordCount {
				t.Errorf("got %d playerCount | expected %d playerCount", len(records), tc.expectedRecordCount)
			}

			if !reflect.DeepEqual(records, tc.expectedRecords) {
				t.Errorf("got dataSlice %v | expected %v", records, tc.expectedRecords)
			}
		})
	}
}
