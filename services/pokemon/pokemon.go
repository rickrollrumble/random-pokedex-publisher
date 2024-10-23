package pokemon

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/rickrollrumble/random-pokemon-publisher/services/bluesky"
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

func CreatePost(id int) error {
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

	postText := fmt.Sprintf("Today's Pokemon of the day is %s\n\nType: %s\n\n%s\n\n",
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

	sprite, err := formatSprite(pokemon.Sprites.Other.OfficialArtwork.FrontDefault, pokemon.Name)
	if err != nil {
		return fmt.Errorf("failed to fetch sprite for pokemon")
	}

	post.Image = []bluesky.ImageDetails{sprite}

	return bluesky.SendPost(context.Background(), post)
}

func formatSprite(url, pokemonName string) (bluesky.ImageDetails, error) {
	var image bluesky.ImageDetails
	resp, err := resty.New().R().Get(url)
	if err != nil {
		return image, fmt.Errorf("failed to get sprite for pokemon")
	}

	if len(resp.Body()) > 1000000 {
		return image, fmt.Errorf("image too large")
	}

	image.Alt = fmt.Sprintf("official artwork of the pokemon %s", pokemonName)
	image.Image.MimeType = resp.Header().Get("Content-Type")
	image.Image.Size = len(resp.Body())
	image.Image.Type = "app.bsky.embed.images"
	image.Image.Ref.Link = string(resp.Body())

	return image, nil
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
