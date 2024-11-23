package main

import (
	"github.com/rickrollrumble/random-pokemon-publisher/services/pokemon"
	"github.com/rs/zerolog/log"
)

func main() {
	res, err := pokemon.Publish()

	if err != nil {
		log.Err(err).Msg(err.Error())
	} else {
		log.Info().Msg(res)
	}
}
