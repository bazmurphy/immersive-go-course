package main

import (
	"errors"
	"reflect"
	"testing"
)

func TestParseJSON(t *testing.T) {
	testCases := []struct {
		name                string
		json                []byte
		expectedError       error
		expectedRecords     []Record
		expectedRecordCount int
	}{
		{
			name:                "valid json",
			json:                []byte(`[{"name":"playerOne","high_score":50},{"name":"playerTwo","high_score":0},{"name":"playerThree","high_score":25}]`),
			expectedRecords:     []Record{{Name: "playerOne", HighScore: 50}, {Name: "playerTwo", HighScore: 0}, {Name: "playerThree", HighScore: 25}},
			expectedRecordCount: 3,
			expectedError:       nil,
		},
		{
			name:                "invalid json",
			json:                []byte(`[{"name":"","high_score":50}`),
			expectedRecords:     nil,
			expectedRecordCount: 0,
			expectedError:       errors.New("[3] could not parse any data in JSON format"),
		},
		{
			name:                "empty json",
			json:                []byte(`[]`),
			expectedRecords:     []Record{},
			expectedRecordCount: 0,
			expectedError:       nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var jsonParser JSONParser

			records, err := jsonParser.Parse(testCase.json)
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
