package main

import (
	"errors"
	"reflect"
	"testing"
)

func TestParseCSV(t *testing.T) {
	testCases := []struct {
		name                string
		csv                 []byte
		expectedRecords     []Record
		expectedRecordCount int
		expectedError       error
	}{
		{
			name: "valid CSV",
			csv:  []byte("playerOne,50\nplayerTwo,0\nplayerThree,25"),
			expectedRecords: []Record{
				{Name: "playerOne", HighScore: 50},
				{Name: "playerTwo", HighScore: 0},
				{Name: "playerThree", HighScore: 25},
			},
			expectedRecordCount: 3,
			expectedError:       nil,
		},
		{
			name:                "CSV with one invalid line",
			csv:                 []byte("playerOne,50\nplayerTwo,0,anotherValue\nplayerThree,25"),
			expectedRecords:     nil,
			expectedRecordCount: 0,
			expectedError:       errors.New("[3] could not parse any data in CSV format"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			records, err := parseCSV(testCase.csv)
			if err != nil {
				if testCase.expectedError != nil {
					if err.Error() != testCase.expectedError.Error() {
						t.Errorf("err actual %v | expected %v", err, testCase.expectedError)
					}
				} else {
					t.Errorf("err actual %v | expected %v", err, testCase.expectedError)
				}
			}

			if len(records) != testCase.expectedRecordCount {
				t.Errorf("got %d playerCount | expected %d playerCount", len(records), testCase.expectedRecordCount)
			}

			if !reflect.DeepEqual(records, testCase.expectedRecords) {
				t.Errorf("got dataSlice %v | expected %v", records, testCase.expectedRecords)
			}
		})
	}
}
