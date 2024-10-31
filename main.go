package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rickrollrumble/random-pokemon-publisher/services/aws"
	"github.com/rickrollrumble/random-pokemon-publisher/services/pokemon"
	"github.com/rs/zerolog"
	"golang.org/x/exp/rand"
)

func main() {
	lambda.Start(publish)
}

func publish() (string, error) {
	logger := zerolog.New(os.Stdout)

	var pokemonToPublish int
	var publishErr error
	for {
		rand.Seed(uint64(time.Now().Unix()))
		pokemonToPublish = rand.Intn(1025) + 1
		alreadyPublished, readErr := readHistory(context.Background(), pokemonToPublish)
		if readErr != nil {
			logger.Err(readErr).Msgf("failed to check if pokemon #%d has been published already; may be double-published", pokemonToPublish)
		}
		if !alreadyPublished {
			publishErr = pokemon.CreatePost(pokemonToPublish)

			if publishErr != nil {
				publishErr = fmt.Errorf("failed to publish pokemon #d: %w", pokemonToPublish, publishErr)
				break
			}

			logger.Info().Msg("successfully created a post on Bluesky")

			if err := updateHistory(context.Background(), pokemonToPublish); err != nil {
				logger.Err(err).Msg("failed to save the published pokemon to the history; this pokemon may be published again")
			}

			return fmt.Sprintf("successfully published pokemon #%d", pokemonToPublish), nil
		}
	}
	return "", publishErr
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
