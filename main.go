package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/rickrollrumble/random-pokemon-publisher/services/pokemon"
	"github.com/rs/zerolog"
	"golang.org/x/exp/rand"
)

func main() {
	logger := zerolog.New(os.Stdout)

	pokeHistory, err := readHistory()
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

func readHistory() (map[int]bool, error) {
	numbers := make(map[int]bool)

	fileContents, err := os.ReadFile("published_pokemon.txt")

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
