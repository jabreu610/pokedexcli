package pokeclient

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/jabreu610/pokedexcli/internal/pokecache"
)

type Pokemon struct {
	Name           string `json:"name"`
	BaseExperience int    `json:"base_experience"`
}

var BaseUrlPokemon string = "https://pokeapi.co/api/v2/pokemon"

var ErrPokemonNotFound error = errors.New("pokemon not found")

func GetPokemon(name string, cache *pokecache.Cache) (Pokemon, error) {
	var p Pokemon
	var d []byte
	fullUrl := BaseUrlPokemon + "/" + name
	d, ok := cache.Get(fullUrl)
	if !ok {
		res, err := http.Get(fullUrl)
		if err != nil {
			return p, err
		}
		if res.StatusCode == 404 {
			return p, ErrPokemonNotFound
		}
		defer res.Body.Close()

		d, err = io.ReadAll(res.Body)
		if err != nil {
			return p, err
		}
		cache.Add(fullUrl, d)
	}

	if err := json.Unmarshal(d, &p); err != nil {
		return p, err
	}
	return p, nil
}
