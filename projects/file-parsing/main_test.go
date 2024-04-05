package main

import (
	"errors"
	"reflect"
	"testing"
)

func TestParseJSON(t *testing.T) {
	type testCaseJSON struct {
		name                string
		json                []byte
		expectedError       error
		expectedDataSlice   []Player
		expectedPlayerCount int
	}

	testCases := []testCaseJSON{
		{
			name:                "valid json",
			json:                []byte(`[{"name":"playerOne","high_score":50},{"name":"playerTwo","high_score":0},{"name":"playerThree","high_score":25}]`),
			expectedError:       nil,
			expectedDataSlice:   []Player{{Name: "playerOne", HighScore: 50}, {Name: "playerTwo", HighScore: 0}, {Name: "playerThree", HighScore: 25}},
			expectedPlayerCount: 3,
		},
		{
			name:                "invalid json",
			json:                []byte(`[{"name":"","high_score":50}`),
			expectedError:       errors.New("[3] could not parse any data in JSON format"),
			expectedDataSlice:   []Player{},
			expectedPlayerCount: 0,
		},
		{
			name:                "empty json",
			json:                []byte(`[]`),
			expectedError:       nil,
			expectedDataSlice:   []Player{},
			expectedPlayerCount: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// why if i use the top line does the test fail but not the bottom line are they not the same?
			// var dataSlice []Player
			dataSlice := []Player{}

			err := parseJSON(tc.json, &dataSlice)

			// i don't like this weird nested error logic i am doing here... what is more elegant?
			if err != nil {
				if tc.expectedError != nil {
					if err.Error() != tc.expectedError.Error() {
						t.Errorf("got error %v | expected error %v", err, tc.expectedError)
					}
				} else {
					t.Errorf("got error %v | expected error %v", err, tc.expectedError)
				}
			}

			if !reflect.DeepEqual(dataSlice, tc.expectedDataSlice) {
				t.Errorf("got dataSlice %v | expected %v", dataSlice, tc.expectedDataSlice)
			}

			if len(dataSlice) != tc.expectedPlayerCount {
				t.Errorf("got %d playerCount | expected %d playerCount", len(dataSlice), tc.expectedPlayerCount)
			}
		})
	}
}

func TestParseRepeatedJSON(t *testing.T) {
	type testCaseRepeatedJSON struct {
		name                string
		repeatedjson        []byte
		expectedError       error
		expectedDataSlice   []Player
		expectedPlayerCount int
	}

	testCases := []testCaseRepeatedJSON{
		{
			name: "valid repeatedJSON",
			repeatedjson: []byte(`{"name":"playerOne","high_score":50}
{"name":"playerTwo","high_score":0}
{"name":"playerThree","high_score":25}`),
			expectedError: nil,
			expectedDataSlice: []Player{
				{Name: "playerOne", HighScore: 50},
				{Name: "playerTwo", HighScore: 0},
				{Name: "playerThree", HighScore: 25},
			},
			expectedPlayerCount: 3,
		},
		{
			name: "valid repeatedJSON with empty lines and comments",
			repeatedjson: []byte(`
# Comment1
{"name":"playerOne","high_score":50}

# Comment2 # Comment3
{"name":"playerTwo","high_score":0}

`),
			expectedError: nil,
			expectedDataSlice: []Player{
				{Name: "playerOne", HighScore: 50},
				{Name: "playerTwo", HighScore: 0},
			},
			expectedPlayerCount: 2,
		},
		{
			name:                "invalid repeatedJSON",
			repeatedjson:        []byte(`{"name":"playerOne","high_score":50{"name":"playerTwo","high_score":}`),
			expectedError:       nil,
			expectedDataSlice:   []Player{},
			expectedPlayerCount: 0,
		},
		{
			name:                "empty repeatedJSON",
			repeatedjson:        []byte(""),
			expectedError:       nil,
			expectedDataSlice:   []Player{},
			expectedPlayerCount: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// why if i use the top line does the test fail but not the bottom line are they not the same?
			// var dataSlice []Player
			dataSlice := []Player{}

			err := parseRepeatedJSON(tc.repeatedjson, &dataSlice)

			if err != nil {
				t.Errorf("got error %v | expected error %v", err, tc.expectedError)
			}

			if len(dataSlice) != tc.expectedPlayerCount {
				t.Errorf("got %d playerCount | expected %d playerCount", len(dataSlice), tc.expectedPlayerCount)
			}

			if !reflect.DeepEqual(dataSlice, tc.expectedDataSlice) {
				t.Errorf("got dataSlice %v | expected %v", dataSlice, tc.expectedDataSlice)
			}
		})
	}
}

func TestParseCSV(t *testing.T) {
	type testCaseCSV struct {
		name                string
		csv                 []byte
		expectedError       error
		expectedDataSlice   []Player
		expectedPlayerCount int
	}

	testCases := []testCaseCSV{
		{
			name:          "valid CSV",
			csv:           []byte("playerOne,50\nplayerTwo,0\nplayerThree,25"),
			expectedError: nil,
			expectedDataSlice: []Player{
				{Name: "playerOne", HighScore: 50},
				{Name: "playerTwo", HighScore: 0},
				{Name: "playerThree", HighScore: 25},
			},
			expectedPlayerCount: 3,
		},
		{
			name:                "CSV with one invalid line",
			csv:                 []byte("playerOne,50\nplayerTwo,0,anotherValue\nplayerThree,25"),
			expectedError:       errors.New("[3] could not parse any data in CSV format"),
			expectedDataSlice:   []Player{},
			expectedPlayerCount: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// why if i use the top line does the test fail but not the bottom line are they not the same?
			// var dataSlice []Player
			dataSlice := []Player{}

			err := parseCSV(tc.csv, &dataSlice)

			// i don't like this weird nested error logic i am doing here... what is more elegant?
			if err != nil {
				if tc.expectedError != nil {
					if err.Error() != tc.expectedError.Error() {
						t.Errorf("got error %v | expected error %v", err, tc.expectedError)
					}
				} else {
					t.Errorf("got error %v | expected error %v", err, tc.expectedError)
				}
			}

			if len(dataSlice) != tc.expectedPlayerCount {
				t.Errorf("got %d playerCount | expected %d playerCount", len(dataSlice), tc.expectedPlayerCount)
			}

			if !reflect.DeepEqual(dataSlice, tc.expectedDataSlice) {
				t.Errorf("got dataSlice %v | expected %v", dataSlice, tc.expectedDataSlice)
			}
		})
	}
}
