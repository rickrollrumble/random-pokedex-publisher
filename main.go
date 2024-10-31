package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/rickrollrumble/random-pokemon-publisher/services/aws"
	"github.com/rickrollrumble/random-pokemon-publisher/services/pokemon"
	"github.com/rs/zerolog"
	"golang.org/x/exp/rand"
)

func main() {
	logger := zerolog.New(os.Stdout)

	var pokemonToPublish int
	for {
		rand.Seed(uint64(time.Now().Unix()))
		pokemonToPublish = rand.Intn(1025) + 1
		alreadyPublished, readErr := readHistory(context.Background(), pokemonToPublish)
		if readErr != nil {
			logger.Err(readErr).Msgf("failed to check if pokemon #%d has been published already; may be double-published", pokemonToPublish)
		}
		if !alreadyPublished {
			postErr := pokemon.CreatePost(pokemonToPublish)

			if postErr != nil {
				logger.Fatal().Msgf("failed to send post to Bluesky: %v", postErr)
			}

			logger.Info().Msg("successfully created a post on Bluesky")

			if err := updateHistory(context.Background(), pokemonToPublish); err != nil {
				logger.Err(err).Msg("failed to save the published pokemon to the history; this pokemon may be published again")
			}

			break
		}
	}
}

func alreadyPublished(pokemonNum int, previouslyPublished map[int]bool) bool {
	return previouslyPublished[pokemonNum]
}

func updateHistory(ctx context.Context, num int) error {
	return aws.CreateFile(ctx, fmt.Sprintf("%d", num))
}

func readHistory(ctx context.Context, pokemonNumber int) (bool, error) {
	return aws.FileExists(ctx, fmt.Sprintf("%d", pokemonNumber))
}
