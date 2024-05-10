package main

import "errors"

func getHighestLowestScoringPlayers(records []Record) (Record, Record, error) {
	// check if the slice of structs is empty
	if len(records) == 0 {
		return Record{}, Record{}, errors.New("[1] the parsed data contains no players")
	}

	// create two variables to hold the highest/lowest scoring players, and set them as the first player in the slice
	playerWithHighestScore := records[0]
	playerWithLowestScore := records[0]

	// loop over the slice of structs and establish the highest/lowest scoring players
	for _, player := range records {
		if player.HighScore > playerWithHighestScore.HighScore {
			playerWithHighestScore = player
		} else if player.HighScore < playerWithLowestScore.HighScore {
			playerWithLowestScore = player
		}
	}

	return playerWithHighestScore, playerWithLowestScore, nil
}
