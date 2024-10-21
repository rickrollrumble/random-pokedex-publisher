package pokemon

type RespPokemon struct {
	Name    string  `json:"name"`
	Sprites Sprites `json:"sprites"`
	Stats   []Stats `json:"stats"`
	Types   []Types `json:"types"`
}
type OfficialArtwork struct {
	FrontDefault string `json:"front_default"`
	FrontShiny   string `json:"front_shiny"`
}

type Other struct {
	OfficialArtwork OfficialArtwork `json:"official-artwork"`
}

type Sprites struct {
	Other Other `json:"other"`
}
type Stat struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}
type Stats struct {
	BaseStat int  `json:"base_stat"`
	Effort   int  `json:"effort"`
	Stat     Stat `json:"stat"`
}
type Type struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}
type Types struct {
	Slot int  `json:"slot"`
	Type Type `json:"type"`
}
type RespPokemonSpecies struct {
	FlavorTextEntries []FlavorTextEntries `json:"flavor_text_entries"`
}
type Language struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}
type Version struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}
type FlavorTextEntries struct {
	FlavorText string   `json:"flavor_text"`
	Language   Language `json:"language"`
	Version    Version  `json:"version"`
}
