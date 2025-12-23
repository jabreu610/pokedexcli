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

const BaseUrlLocationArea = "https://pokeapi.co/api/v2/location-area"

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

		data, err := io.ReadAll(res.Body)
		if err != nil {
			return out, err
		}
		d = data
		cache.Add(url, data)
	}

	if err := json.Unmarshal(d, &out); err != nil {
		return out, err
	}
	return out, nil
}
