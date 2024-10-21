package main

import (
	"context"
	"os"

	"github.com/rickrollrumble/random-pokemon-publisher/services/bluesky"
	"github.com/rs/zerolog"
)

func main() {
	logger := zerolog.New(os.Stdout)

	bskyClient, err := bluesky.CreateNewSession()
	if err != nil {
		logger.Fatal().Msg("failed to create Bluesky client")
	}

	bskyPost := bluesky.PostParams{
		Text: "this post was created using Bluesky's API. Track this project here: ",
		Link: "http://github.com/rickrollrumble/random-pokemon-publisher",
	}
	postErr := bluesky.SendPost(context.WithValue(context.Background(), "session", bskyClient), bskyPost)

	if postErr != nil {
		logger.Fatal().Msgf("failed to send post to Bluesky: %v", postErr)
	}

	logger.Info().Msg("successfully created a post on Bluesky")
}
