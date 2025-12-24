package pokeclient

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/jabreu610/pokedexcli/internal/pokecache"
)

type LocationArea struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type LocationAreaResponse struct {
	Count    int            `json:"count"`
	Next     *string        `json:"next"`
	Previous *string        `json:"previous"`
	Results  []LocationArea `json:"results"`
}

var BaseUrlLocationArea string = "https://pokeapi.co/api/v2/location-area"

func GetLocationAreas(url string, cache *pokecache.Cache) (LocationAreaResponse, error) {
	out := LocationAreaResponse{}
	var d []byte
	d, ok := cache.Get(url)
	if !ok {
		res, err := http.Get(url)
		if err != nil {
			return out, err
		}
		defer res.Body.Close()

		d, err = io.ReadAll(res.Body)
		if err != nil {
			return out, err
		}
		cache.Add(url, d)
	}

	if err := json.Unmarshal(d, &out); err != nil {
		return out, err
	}
	return out, nil
}
