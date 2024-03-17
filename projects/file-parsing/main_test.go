package main

import (
	"errors"
	"testing"
)

func TestParseJSON(t *testing.T) {
	t.Run("valid json", func(t *testing.T) {
		validJSON := []byte(
			`[
				{"name":"playerOne","high_score":50},
				{"name":"playerTwo","high_score":0},
				{"name":"playerThree","high_score":25}
			]`)

		var dataSlice []Player

		err := parseJSON(validJSON, &dataSlice)

		if err != nil {
			t.Errorf("got %v | want no error", err)
		}

		if len(dataSlice) != 3 {
			t.Errorf("got %d players | want %d players", len(dataSlice), 3)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		invalidJSON := []byte(
			`[{"name":"","high_score":50}`)

		var dataSlice []Player

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

func TestParseRepeatedJSON(t *testing.T) {
	t.Run("valid repeatedJSON", func(t *testing.T) {
		validRepeatedJSON := []byte(
			`{"name":"playerOne","high_score":50}
			# Comment
			{"name":"playerTwo","high_score":0}
			# Comment # Comment
			{"name":"playerThree","high_score":25}`)

		var dataSlice []Player

		err := parseRepeatedJSON(validRepeatedJSON, &dataSlice)

		if err != nil {
			t.Errorf("got %v | want no error", err)
		}

		if len(dataSlice) != 3 {
			t.Errorf("got %d, want %d", len(dataSlice), 3)
		}
	})

	t.Run("invalid repeatedJSON", func(t *testing.T) {
		invalidRepeatedJSON := []byte(`{"name":"playerOne","high_score":50{"name":"playerTwo","high_score":}`)

		var dataSlice []Player

		err := parseRepeatedJSON(invalidRepeatedJSON, &dataSlice)

		wantErr := errors.New("[3] could not parse any data in repeated JSON format")

		if err == nil {
			t.Errorf("got no error | want %q", wantErr)
		}

		if len(dataSlice) != 0 {
			t.Errorf("got %d | want %d", len(dataSlice), 0)
		}
	})
}
