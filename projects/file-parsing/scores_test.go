package main

import (
	"fmt"
	"testing"
)

// TODO: write these tests
func TestGetHighestLowestScoringPlayers(t *testing.T) {
	testCases := []struct {
		name    string
		records []Record
	}{
		{
			name:    "placeholder",
			records: []Record{},
		},
	}

	for _, testCase := range testCases {
		highestScoringPlayer, lowestScoringPlayer, err := getHighestLowestScoringPlayers(testCase.records)
		if err != nil {
			// TODO: placeholder
			fmt.Println(err)
		}
		// TODO: placeholder
		fmt.Println(highestScoringPlayer, lowestScoringPlayer)
	}
}
