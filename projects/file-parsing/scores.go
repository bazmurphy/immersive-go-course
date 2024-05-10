package main

import "errors"

func getHighestLowestScoringPlayers(records []Record) (Record, Record, error) {
	// check if the records are empty
	if len(records) == 0 {
		return Record{}, Record{}, errors.New("[1] the parsed records contains no players")
	}

	// create two variables to hold the highest/lowest scoring players, and set them as the first player in the slice
	highestScoringPlayer := records[0]
	lowestScoringPlayer := records[0]

	// loop over the slice of structs and establish the highest/lowest scoring players
	for _, player := range records {
		if player.HighScore > highestScoringPlayer.HighScore {
			highestScoringPlayer = player
		} else if player.HighScore < lowestScoringPlayer.HighScore {
			lowestScoringPlayer = player
		}
	}

	return highestScoringPlayer, lowestScoringPlayer, nil
}
