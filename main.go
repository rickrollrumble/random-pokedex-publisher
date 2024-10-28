package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/rickrollrumble/random-pokemon-publisher/services/aws"
	"github.com/rickrollrumble/random-pokemon-publisher/services/pokemon"
	"github.com/rs/zerolog"
	"golang.org/x/exp/rand"
)

func main() {
	logger := zerolog.New(os.Stdout)

	env, envLoadErr := godotenv.Read(".env")
	if envLoadErr != nil {
		logger.Fatal().Msgf("failed to load environment variables")
	}

	ctx := context.WithValue(context.Background(), "bucket_name", env["S3_BUCKET"])
	ctx = context.WithValue(ctx, "object_key", env["S3_OBJECT"])
	ctx = context.WithValue(ctx, "aws_region", env["AWS_REGION"])
	ctx = context.WithValue(ctx, "aws_key_id", env["AWS_ACCESS_KEY"])
	ctx = context.WithValue(ctx, "aws_secret", env["AWS_SECRET"])

	pokeHistory, err := readHistory(ctx)
	if err != nil {
		logger.Fatal().Msgf("failed to fetch previously published pokemon: %v", err)
	}

	var pokemonToPublish int
	for {
		pokemonToPublish = rand.Intn(1025) + 1
		if !alreadyPublished(pokemonToPublish, pokeHistory) {
			postErr := pokemon.CreatePost(pokemonToPublish)

			if postErr != nil {
				logger.Fatal().Msgf("failed to send post to Bluesky: %v", postErr)
			}

			logger.Info().Msg("successfully created a post on Bluesky")

			if err := saveToFile(pokemonToPublish); err != nil {
				logger.Err(err).Msg("failed to save the published pokemon to the history; this pokemon may be published again")
			}

			break
		}
	}
}

func alreadyPublished(pokemonNum int, previouslyPublished map[int]bool) bool {
	return previouslyPublished[pokemonNum]
}

func saveToFile(num int) error {
	f, err := os.OpenFile("published_pokemon.txt", os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return fmt.Errorf("failed to save to file: %w", err)
	}

	_, err = f.WriteString(fmt.Sprintf("%d\n", num))

	return err
}

func readHistory(ctx context.Context) (map[int]bool, error) {
	numbers := make(map[int]bool)
	fileName := ctx.Value("object_key").(string)

	getFileErr := aws.GetFile(ctx, fileName)
	if getFileErr != nil {
		return nil, fmt.Errorf("failed to read previously published pokemon: %w", getFileErr)
	}

	fileContents, err := os.ReadFile(fileName)

	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to read previously published pokemon: %w", err)
	}

	lines := strings.Split(string(fileContents), "\n")

	for _, line := range lines {
		if num, err := strconv.Atoi(strings.TrimSpace(line)); err == nil {
			numbers[num] = true
		}
	}

	return numbers, nil
}
