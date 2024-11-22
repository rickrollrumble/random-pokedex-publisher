package pokebot

import (
	"fmt"
	"net/http"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/rickrollrumble/random-pokemon-publisher/services/pokemon"
)

func init() {
	functions.HTTP("Publish", publish)
}

func publish(w http.ResponseWriter, r *http.Request) {
	resp, err := pokemon.Publish()
	if err != nil {
		fmt.Fprintln(w, err.Error())
	}
	fmt.Fprintln(w, resp)
}
