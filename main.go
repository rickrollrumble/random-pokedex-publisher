package main

import (
	"os"

	"github.com/rickrollrumble/random-pokemon-publisher/services/pokemon"
	"github.com/rs/zerolog"
)

func main() {
	logger := zerolog.New(os.Stdout)

	postErr := pokemon.CreatePost(1000)

	if postErr != nil {
		logger.Fatal().Msgf("failed to send post to Bluesky: %v", postErr)
	}

	logger.Info().Msg("successfully created a post on Bluesky")
}
