package pokeclient

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/jabreu610/pokedexcli/internal/pokecache"
)

type PokemonEntry struct {
	Name string `json:"name"`
}

type EncounterEntry struct {
	Pokemon PokemonEntry `json:"pokemon"`
}

type LocationAreaByNameResponse struct {
	PokemonEncounters []EncounterEntry `json:"pokemon_encounters"`
}

func GetPokemonForLocationName(name string, cache *pokecache.Cache) ([]string, error) {
	resParsed := LocationAreaByNameResponse{}
	out := []string{}
	var d []byte
	fullUrl := BaseUrlLocationArea + "/" + name
	d, ok := cache.Get(fullUrl)
	if !ok {
		res, err := http.Get(fullUrl)
		if err != nil {
			return out, err
		}
		defer res.Body.Close()

		d, err = io.ReadAll(res.Body)
		if err != nil {
			return out, err
		}
		cache.Add(fullUrl, d)
	}

	if err := json.Unmarshal(d, &resParsed); err != nil {
		return out, err
	}
	for _, entry := range resParsed.PokemonEncounters {
		out = append(out, entry.Pokemon.Name)
	}
	return out, nil
}
