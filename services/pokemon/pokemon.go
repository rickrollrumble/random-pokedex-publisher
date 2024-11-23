package pokemon

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/rickrollrumble/random-pokemon-publisher/services/bluesky"
	"github.com/rickrollrumble/random-pokemon-publisher/services/cloud/gcp"
	"github.com/rs/zerolog"
	"golang.org/x/exp/rand"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func getPokemon(id int) (RespPokemon, error) {
	client := resty.New().SetBaseURL("https://pokeapi.co/api/v2")

	req := client.R()

	var pokemon RespPokemon

	resp, err := req.Get(fmt.Sprintf("pokemon/%d", id))
	if err != nil {
		return pokemon, fmt.Errorf("failed to make request to fetch pokemon: %w", err)
	}

	if resp.IsError() {
		return pokemon, fmt.Errorf("received an error while trying to fetch pokemon")
	}

	unmarshalErr := json.Unmarshal(resp.Body(), &pokemon)
	if unmarshalErr != nil {
		return pokemon, fmt.Errorf("fetched GET pokemon response is not a valid JSON: %w", unmarshalErr)
	}

	if reflect.ValueOf(pokemon).IsZero() {
		return pokemon, fmt.Errorf("could not populate pokemon data from response")
	}

	return pokemon, nil
}

func getFlavorText(id int) (string, error) {
	client := resty.New().SetBaseURL("https://pokeapi.co/api/v2")

	req := client.R()

	flavorText := ""

	resp, err := req.Get(fmt.Sprintf("pokemon-species/%d", id))
	if err != nil {
		return flavorText, fmt.Errorf("failed to make request to get flavor text of pokemon: %w", err)
	}

	if resp.IsError() {
		return flavorText, fmt.Errorf("received an error while trying to fetch flavor text")
	}

	var speciesResp RespPokemonSpecies

	unmarshalErr := json.Unmarshal(resp.Body(), &speciesResp)

	if unmarshalErr != nil {
		return flavorText, fmt.Errorf("fetched GET pokemon species response is not a valid JSON: %w", unmarshalErr)
	}

	if reflect.ValueOf(speciesResp).IsZero() {
		return flavorText, fmt.Errorf("could not populate species from response")
	}

	for _, flavorTextEntry := range speciesResp.FlavorTextEntries {
		if flavorTextEntry.Language.Name == "en" {
			flavorText = flavorTextEntry.FlavorText
			break
		}
	}

	if flavorText == "" {
		return flavorText, fmt.Errorf("English flavor text not found for this species")
	}

	return flavorText, nil
}

func getSprite(imageUrl string) ([]byte, error) {
	resp, err := resty.New().R().Get(imageUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sprite for pokemon: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("received an error while fetching pokemon sprite")
	}

	return resp.Body(), nil
}

func createPost(id int) error {
	pokemon, err := getPokemon(id)
	types := []string{}

	titleCaser := cases.Title(language.Und)

	for _, pokemonType := range pokemon.Types {
		types = append(types, titleCaser.String(strings.ToLower(pokemonType.Type.Name)))
	}

	flavorText, flavorTextErr := getFlavorText(id)
	if flavorTextErr != nil {
		return fmt.Errorf("failed to get flavor text for pokemon %d: %w", id, flavorTextErr)
	}

	postText := fmt.Sprintf("Today's #Pokemon of the day is %s\n\nType: %s\n\n%s\n\n",
		titleCaser.String(strings.ToLower(pokemon.Name)),
		strings.Join(types[:], "/"),
		flavorText,
	)

	bst := 0
	for _, pokemonStat := range pokemon.Stats {
		bst += pokemonStat.BaseStat
		postText += fmt.Sprintf("%s\n", fmt.Sprintf("%s - %d", pokemonStat.Stat.Name, pokemonStat.BaseStat))
	}

	postText += fmt.Sprintf("bst - %d\n", bst)

	post := bluesky.PostParams{
		Text: postText,
	}

	sprite, err := formatSprite(pokemon.Sprites.Other.OfficialArtwork.FrontDefault)
	if err != nil {
		return fmt.Errorf("failed to fetch sprite for pokemon: %w", err)
	}

	post.Images = []bluesky.ImageDetails{
		{
			Alt:   fmt.Sprintf("official artwork of the pokemon %s", titleCaser.String(strings.ToLower(pokemon.Name))),
			Image: sprite,
		},
	}

	return bluesky.SendPost(context.Background(), post)
}

func formatSprite(url string) (bluesky.RespImageUpload, error) {
	resp, err := resty.New().R().Get(url)
	if err != nil {
		return bluesky.RespImageUpload{}, fmt.Errorf("failed to get sprite for pokemon: %w", err)
	}

	if len(resp.Body()) > 1000000 {
		return bluesky.RespImageUpload{}, fmt.Errorf("image too large")
	}

	uploadedImage, uploadedImageErr := bluesky.UploadImage(context.Background(), resp.Body())
	if uploadedImageErr != nil {
		return bluesky.RespImageUpload{}, fmt.Errorf("failed to upload sprite: %w", uploadedImageErr)
	}

	return uploadedImage, nil
}

func formatStat(statName string, statVal int) string {
	// Replace hyphens and underscores with spaces
	statName = strings.ReplaceAll(statName, "-", " ")
	statName = strings.ReplaceAll(statName, "_", " ")

	// Remove any other special characters using a regular expression
	re := regexp.MustCompile(`[^a-zA-Z0-9\s]+`)
	cleanStatName := re.ReplaceAllString(statName, "")

	// Convert the stat name to title case using golang.org/x/text/cases
	titleCaser := cases.Title(language.Und) // Using language.Und for language-agnostic title casing
	titleCaseStatName := titleCaser.String(strings.ToLower(cleanStatName))

	// Format the string so that the stat name is left-aligned and the stat value is right-aligned
	// within a total width of 25 characters.
	return fmt.Sprintf("%-20s%5d", titleCaseStatName, statVal)
}

var bucket = gcp.Bucket{}

func Publish() (string, error) {
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
			publishErr = createPost(pokemonToPublish)

			if publishErr != nil {
				publishErr = fmt.Errorf("failed to publish pokemon #%d: %w", pokemonToPublish, publishErr)
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
	return bucket.CreateFile(ctx, fmt.Sprintf("%d", num))
}

func readHistory(ctx context.Context, pokemonNumber int) (bool, error) {
	return bucket.FileExists(ctx, fmt.Sprintf("%d", pokemonNumber))
}
