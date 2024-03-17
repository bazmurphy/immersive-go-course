package main

import (
	"errors"
	"testing"
)

func TestParseJSON(t *testing.T) {
	t.Run("valid json", func(t *testing.T) {
		// create a valid json "file"
		validJSON := []byte(`[{"name":"playerOne","high_score":50},{"name":"playerTwo","high_score":0},{"name":"playerThree","high_score":25}]`)

		// initialise an empty slice of structs (players)
		var dataSlice []Player

		// try to parse the json
		err := parseJSON(validJSON, &dataSlice)

		// there should be no error
		if err != nil {
			t.Errorf("got %v | want no error", err)
		}

		// check the number of structs (players) in the data slice
		if len(dataSlice) != 3 {
			t.Errorf("got %d players | want %d players", len(dataSlice), 3)
		}
	})

	t.Run("invalid json - one player", func(t *testing.T) {
		// create an invalid json "file"
		invalidJSON := []byte(`[{"name":,"high_score":50}`)

		// initialise an empty slice of structs (players)
		var dataSlice []Player

		// try to parse the json
		err := parseJSON(invalidJSON, &dataSlice)

		wantErr := errors.New("[3] could not parse any data in JSON format")

		if err == nil {
			t.Errorf("got no error | want %q", wantErr)
		}

		if len(dataSlice) != 0 {
			t.Errorf("got %d | want %d", len(dataSlice), 0)
		}
	})
}
