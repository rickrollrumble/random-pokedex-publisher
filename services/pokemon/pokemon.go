package pokemon

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/rickrollrumble/random-pokemon-publisher/services/bluesky"
	"github.com/rickrollrumble/random-pokemon-publisher/services/cloud/gcp"
	"github.com/rs/zerolog"
	"github.com/vicanso/go-charts/v2"
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

	postText := fmt.Sprintf("Today's #Pokemon of the day is %s\n\nType: %s\n\n",
		titleCaser.String(strings.ToLower(pokemon.Name)),
		strings.Join(types[:], "/"),
	)

	stats := make(map[string]float64)

	for _, pokemonStat := range pokemon.Stats {
		stats[pokemonStat.Stat.Name] = float64(pokemonStat.BaseStat)
	}

	statsChart, statChartErr := createStatsChart(stats, pokemon.Name)
	if statChartErr != nil {
		return fmt.Errorf("failed to upload stats chart: %w", statChartErr)
	}

	flavorText, flavorTextErr := getFlavorText(id)
	if flavorTextErr != nil {
		return fmt.Errorf("failed to get flavor text for pokemon %d: %w", id, flavorTextErr)
	}

	postText += flavorText

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
		{
			Alt:   fmt.Sprintf("radar chart of the stats of the pokemon %s", titleCaser.String(strings.ToLower(pokemon.Name))),
			Image: statsChart,
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

func createStatsChart(stats map[string]float64, name string) (bluesky.RespImageUpload, error) {
	// a map does not necessarily have the same order of keys every time
	// this causes the stats to be in a random order during each run and causes the charts to
	// not be standardized.
	// by initializing the stat name, the map lookup happens in a fixed order no matter the
	// order of the keys itself.
	statNames := []string{"hp", "attack", "defense", "special-attack", "special-defense", "speed"}

	statValues := [][]float64{{}}

	bst := float64(0)

	for i, statName := range statNames {
		statNames[i] = fmt.Sprintf("%s [%.0f]", statName, stats[statName])
		statValues[0] = append(statValues[0], stats[statName])
	}

	max_stat_val := float64(0xFF)
	chart, err := charts.RadarRender(
		statValues,
		charts.SVGTypeOption(),
		charts.TitleOptionFunc(charts.TitleOption{
			Text: fmt.Sprintf("%s base stat total - %.0f", name, bst),
			Left: charts.PositionCenter,
		}),
		charts.RadarIndicatorOptionFunc(
			statNames,
			[]float64{
				max_stat_val,
				max_stat_val,
				max_stat_val,
				max_stat_val,
				max_stat_val,
				max_stat_val,
			}),
	)
	if err != nil {
		return bluesky.RespImageUpload{}, fmt.Errorf("failed to create stats chart: %w", err)
	}
	buf, err := chart.Bytes()
	if len(buf) > 1000000 {
		return bluesky.RespImageUpload{}, fmt.Errorf("image too large")
	}

	uploadedImage, uploadedImageErr := bluesky.UploadImage(context.Background(), buf)
	if uploadedImageErr != nil {
		return bluesky.RespImageUpload{}, uploadedImageErr
	}

	return uploadedImage, nil
}
