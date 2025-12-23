package pokeclient

import (
	"encoding/json"
	"io"
	"net/http"
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

func GetLocationAreas(url string) (LocationAreaResponse, error) {
	res, err := http.Get(url)
	out := LocationAreaResponse{}
	if err != nil {
		return out, err
	}
	defer res.Body.Close()

	d, err := io.ReadAll(res.Body)
	if err != nil {
		return out, err
	}

	if err := json.Unmarshal(d, &out); err != nil {
		return out, err
	}
	return out, nil
}
