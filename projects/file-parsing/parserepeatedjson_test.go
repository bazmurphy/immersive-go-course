package main

import (
	"reflect"
	"testing"
)

func TestParseRepeatedJSON(t *testing.T) {
	testCases := []struct {
		name                string
		repeatedjson        []byte
		expectedRecords     []Record
		expectedRecordCount int
		expectedError       error
	}{
		{
			name: "valid repeatedJSON",
			repeatedjson: []byte(`{"name":"playerOne","high_score":50}
{"name":"playerTwo","high_score":0}
{"name":"playerThree","high_score":25}`),
			expectedRecords: []Record{
				{Name: "playerOne", HighScore: 50},
				{Name: "playerTwo", HighScore: 0},
				{Name: "playerThree", HighScore: 25},
			},
			expectedRecordCount: 3,
			expectedError:       nil,
		},
		{
			name: "valid repeatedJSON with empty lines and comments",
			repeatedjson: []byte(`
# Comment1
{"name":"playerOne","high_score":50}

# Comment2 # Comment3
{"name":"playerTwo","high_score":0}

`),
			expectedRecords: []Record{
				{Name: "playerOne", HighScore: 50},
				{Name: "playerTwo", HighScore: 0},
			},
			expectedRecordCount: 2,
			expectedError:       nil,
		},
		{
			name:                "invalid repeatedJSON",
			repeatedjson:        []byte(`{"name":"playerOne","high_score":50{"name":"playerTwo","high_score":}`),
			expectedRecords:     nil,
			expectedRecordCount: 0,
			expectedError:       nil,
		},
		{
			name:                "empty repeatedJSON",
			repeatedjson:        []byte{},
			expectedRecords:     nil,
			expectedRecordCount: 0,
			expectedError:       nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			records, err := parseRepeatedJSON(testCase.repeatedjson)
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
				t.Errorf("recordCount actual %v | expected %v", len(records), testCase.expectedRecordCount)
			}

			if !reflect.DeepEqual(records, testCase.expectedRecords) {
				t.Errorf("records actual %v | expected %v", records, testCase.expectedRecords)
			}
		})
	}
}
