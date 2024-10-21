package pokemon

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/go-resty/resty/v2"
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
