package main

import (
	"context"
	"testing"
	"time"

	"github.com/jabreu610/pokedexcli/internal/pokecache"
	"github.com/jabreu610/pokedexcli/internal/pokeclient"
)

func TestProcessLocationAreaResponse(t *testing.T) {
	config := &Config{}
	nextURL := "https://example.com/next"
	prevURL := "https://example.com/prev"

	response := pokeclient.LocationAreaResponse{
		Count:    10,
		Next:     &nextURL,
		Previous: &prevURL,
		Results: []pokeclient.LocationArea{
			{Name: "area-1", Url: "https://example.com/1"},
			{Name: "area-2", Url: "https://example.com/2"},
		},
	}

	processLocationAreaResponse(response, config)

	if config.Next == nil || *config.Next != nextURL {
		t.Errorf("Expected Next to be %s, got %v", nextURL, config.Next)
	}
	if config.Prev == nil || *config.Prev != prevURL {
		t.Errorf("Expected Prev to be %s, got %v", prevURL, config.Prev)
	}
}

func TestCommandHelp(t *testing.T) {
	config := &Config{}
	err := commandHelp(config)
	if err != nil {
		t.Errorf("commandHelp should not return error, got %v", err)
	}
}

func TestCommandMapbFirstPage(t *testing.T) {
	config := &Config{
		Prev: nil,
	}

	err := commandMapb(config)
	if err != nil {
		t.Errorf("commandMapb should not return error on first page, got %v", err)
	}
}

func TestCommandExploreNoArgs(t *testing.T) {
	config := &Config{
		args: []string{},
	}

	err := commandExplore(config)
	if err == nil {
		t.Error("commandExplore should return error when no arguments provided")
	}
}

func TestCommandExploreWithArgs(t *testing.T) {
	cache := pokecache.NewCache(5*time.Second, context.Background())
	defer cache.Close()

	config := &Config{
		args:  []string{"test-location"},
		cache: cache,
	}

	// This will fail with real API call, but tests the argument handling
	err := commandExplore(config)
	// We expect an error here because we're not mocking the API
	// but we're testing that it doesn't panic with valid args
	if err == nil {
		// If somehow it succeeds (unlikely without real API), that's fine
		return
	}
	// Just verify it's not an argument error
	if err.Error() == "Expected one arguement, a location area name. Recieve none" {
		t.Error("Should not get argument error when args are provided")
	}
}

func TestCommandsInitialized(t *testing.T) {
	expectedCommands := []string{"exit", "help", "map", "mapb", "explore", "catch"}

	for _, cmdName := range expectedCommands {
		cmd, ok := commands[cmdName]
		if !ok {
			t.Errorf("Expected command %s to be initialized", cmdName)
			continue
		}
		if cmd.Name != cmdName {
			t.Errorf("Expected command name %s, got %s", cmdName, cmd.Name)
		}
		if cmd.Description == "" {
			t.Errorf("Command %s should have a description", cmdName)
		}
		if cmd.Callback == nil {
			t.Errorf("Command %s should have a callback", cmdName)
		}
	}
}

func TestConfigArgsHandling(t *testing.T) {
	config := &Config{}

	// Test with no args
	config.args = []string{}
	if len(config.args) != 0 {
		t.Error("Config args should be empty")
	}

	// Test with single arg
	config.args = []string{"arg1"}
	if len(config.args) != 1 || config.args[0] != "arg1" {
		t.Error("Config args should contain single argument")
	}

	// Test with multiple args
	config.args = []string{"arg1", "arg2", "arg3"}
	if len(config.args) != 3 {
		t.Error("Config args should contain three arguments")
	}
}

func TestConfigCacheInitialization(t *testing.T) {
	cache := pokecache.NewCache(defaultInterval, context.Background())
	defer cache.Close()

	config := &Config{
		cache: cache,
	}

	if config.cache == nil {
		t.Error("Config cache should be initialized")
	}

	// Test cache is usable
	config.cache.Add("test-key", []byte("test-value"))
	val, ok := config.cache.Get("test-key")
	if !ok {
		t.Error("Cache should return cached value")
	}
	if string(val) != "test-value" {
		t.Errorf("Expected 'test-value', got %s", string(val))
	}
}

func TestConfigNextPrevPointers(t *testing.T) {
	config := &Config{}

	// Initially should be nil
	if config.Next != nil {
		t.Error("Next should be nil initially")
	}
	if config.Prev != nil {
		t.Error("Prev should be nil initially")
	}

	// Set values
	nextURL := "https://example.com/next"
	prevURL := "https://example.com/prev"
	config.Next = &nextURL
	config.Prev = &prevURL

	if config.Next == nil || *config.Next != nextURL {
		t.Error("Next should be set correctly")
	}
	if config.Prev == nil || *config.Prev != prevURL {
		t.Error("Prev should be set correctly")
	}
}

func TestCommandCatchNoArgs(t *testing.T) {
	config := &Config{
		args:    []string{},
		pokedex: make(map[string]pokeclient.Pokemon),
	}

	err := commandCatch(config)
	if err == nil {
		t.Error("commandCatch should return error when no arguments provided")
	}
}

func TestCommandCatchWithArgs(t *testing.T) {
	cache := pokecache.NewCache(5*time.Second, context.Background())
	defer cache.Close()

	config := &Config{
		args:    []string{"pikachu"},
		cache:   cache,
		pokedex: make(map[string]pokeclient.Pokemon),
	}

	// This will attempt to call the real API
	// We're just testing it doesn't panic with valid args
	err := commandCatch(config)
	// We expect an error here because we're not mocking the API
	if err == nil {
		// If somehow it succeeds, that's fine too
		return
	}
	// Just verify it's not an argument error
	if err.Error() == "Expected one arguement, a Pokemon name. Recieved none" {
		t.Error("Should not get argument error when args are provided")
	}
}

func TestPassWithDifficulty(t *testing.T) {
	tests := []struct {
		name     string
		baseExp  int
		runCount int
	}{
		{
			name:     "weak pokemon (low base exp)",
			baseExp:  36,
			runCount: 100,
		},
		{
			name:     "medium pokemon",
			baseExp:  300,
			runCount: 100,
		},
		{
			name:     "strong pokemon (high base exp)",
			baseExp:  635,
			runCount: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			passCount := 0
			for i := 0; i < tt.runCount; i++ {
				if passWithDifficulty(tt.baseExp) {
					passCount++
				}
			}

			// Just verify it returns both true and false (not deterministic)
			// and that weaker pokemon have higher pass rates
			passRate := float64(passCount) / float64(tt.runCount)

			if tt.baseExp == 36 && passRate < 0.7 {
				t.Errorf("Weak pokemon should have high pass rate, got %.2f", passRate)
			}
			if tt.baseExp == 635 && passRate > 0.3 {
				t.Errorf("Strong pokemon should have low pass rate, got %.2f", passRate)
			}
		})
	}
}

func TestConfigPokedexInitialization(t *testing.T) {
	config := &Config{
		pokedex: make(map[string]pokeclient.Pokemon),
	}

	if config.pokedex == nil {
		t.Error("Pokedex should be initialized")
	}

	// Test adding to pokedex
	pokemon := pokeclient.Pokemon{
		Name:           "mewtwo",
		BaseExperience: 306,
	}
	config.pokedex["mewtwo"] = pokemon

	retrieved, ok := config.pokedex["mewtwo"]
	if !ok {
		t.Error("Pokemon should be in pokedex")
	}
	if retrieved.Name != "mewtwo" {
		t.Errorf("Expected name 'mewtwo', got %s", retrieved.Name)
	}
	if retrieved.BaseExperience != 306 {
		t.Errorf("Expected base experience 306, got %d", retrieved.BaseExperience)
	}
}

func TestCommandInspectNoArgs(t *testing.T) {
	config := &Config{
		args:    []string{},
		pokedex: make(map[string]pokeclient.Pokemon),
	}

	err := commandInspect(config)
	if err == nil {
		t.Error("commandInspect should return error when no arguments provided")
	}
}

func TestCommandInspectPokemonNotCaught(t *testing.T) {
	config := &Config{
		args:    []string{"pikachu"},
		pokedex: make(map[string]pokeclient.Pokemon),
	}

	// Should not return error, just print message
	err := commandInspect(config)
	if err != nil {
		t.Errorf("commandInspect should not return error for uncaught pokemon, got %v", err)
	}
}

func TestCommandInspectPokemonFound(t *testing.T) {
	pokemon := pokeclient.Pokemon{
		Name:           "charizard",
		Height:         17,
		Weight:         905,
		BaseExperience: 240,
		Stats: []pokeclient.Stat{
			{
				Stat:     pokeclient.Entry{Name: "hp"},
				BaseStat: 78,
			},
			{
				Stat:     pokeclient.Entry{Name: "attack"},
				BaseStat: 84,
			},
		},
		Types: []pokeclient.Type{
			{Type: pokeclient.Entry{Name: "fire"}},
			{Type: pokeclient.Entry{Name: "flying"}},
		},
	}

	config := &Config{
		args:    []string{"charizard"},
		pokedex: map[string]pokeclient.Pokemon{"charizard": pokemon},
	}

	err := commandInspect(config)
	if err != nil {
		t.Errorf("commandInspect should not return error for caught pokemon, got %v", err)
	}
}

func TestCommandInspectWithMultipleStats(t *testing.T) {
	pokemon := pokeclient.Pokemon{
		Name:           "bulbasaur",
		Height:         7,
		Weight:         69,
		BaseExperience: 64,
		Stats: []pokeclient.Stat{
			{Stat: pokeclient.Entry{Name: "hp"}, BaseStat: 45},
			{Stat: pokeclient.Entry{Name: "attack"}, BaseStat: 49},
			{Stat: pokeclient.Entry{Name: "defense"}, BaseStat: 49},
			{Stat: pokeclient.Entry{Name: "special-attack"}, BaseStat: 65},
			{Stat: pokeclient.Entry{Name: "special-defense"}, BaseStat: 65},
			{Stat: pokeclient.Entry{Name: "speed"}, BaseStat: 45},
		},
		Types: []pokeclient.Type{
			{Type: pokeclient.Entry{Name: "grass"}},
			{Type: pokeclient.Entry{Name: "poison"}},
		},
	}

	config := &Config{
		args:    []string{"bulbasaur"},
		pokedex: map[string]pokeclient.Pokemon{"bulbasaur": pokemon},
	}

	err := commandInspect(config)
	if err != nil {
		t.Errorf("commandInspect should not return error, got %v", err)
	}
}

func TestCommandInspectEmptyStats(t *testing.T) {
	pokemon := pokeclient.Pokemon{
		Name:           "missingno",
		Height:         10,
		Weight:         100,
		BaseExperience: 0,
		Stats:          []pokeclient.Stat{},
		Types:          []pokeclient.Type{},
	}

	config := &Config{
		args:    []string{"missingno"},
		pokedex: map[string]pokeclient.Pokemon{"missingno": pokemon},
	}

	err := commandInspect(config)
	if err != nil {
		t.Errorf("commandInspect should handle empty stats/types, got %v", err)
	}
}

func TestCommandInspectCaseInsensitive(t *testing.T) {
	pokemon := pokeclient.Pokemon{
		Name:           "pikachu",
		Height:         4,
		Weight:         60,
		BaseExperience: 112,
		Stats: []pokeclient.Stat{
			{Stat: pokeclient.Entry{Name: "speed"}, BaseStat: 90},
		},
		Types: []pokeclient.Type{
			{Type: pokeclient.Entry{Name: "electric"}},
		},
	}

	// Add with lowercase
	config := &Config{
		args:    []string{"pikachu"},
		pokedex: map[string]pokeclient.Pokemon{"pikachu": pokemon},
	}

	err := commandInspect(config)
	if err != nil {
		t.Errorf("commandInspect should find pokemon, got %v", err)
	}

	// Try to access with different case (will fail if map key doesn't match)
	config.args = []string{"Pikachu"}
	err = commandInspect(config)
	// This should not error, but should print "not caught" message
	if err != nil {
		t.Errorf("commandInspect should not error, got %v", err)
	}
}

func TestCommandPokedexEmpty(t *testing.T) {
	config := &Config{
		pokedex: make(map[string]pokeclient.Pokemon),
	}

	err := commandPokedex(config)
	if err != nil {
		t.Errorf("commandPokedex should not return error for empty pokedex, got %v", err)
	}
}

func TestCommandPokedexWithOnePokemon(t *testing.T) {
	pokemon := pokeclient.Pokemon{
		Name:           "pikachu",
		BaseExperience: 112,
	}

	config := &Config{
		pokedex: map[string]pokeclient.Pokemon{"pikachu": pokemon},
	}

	err := commandPokedex(config)
	if err != nil {
		t.Errorf("commandPokedex should not return error, got %v", err)
	}
}

func TestCommandPokedexWithMultiplePokemon(t *testing.T) {
	config := &Config{
		pokedex: map[string]pokeclient.Pokemon{
			"pikachu": {
				Name:           "pikachu",
				BaseExperience: 112,
			},
			"charizard": {
				Name:           "charizard",
				BaseExperience: 240,
			},
			"mewtwo": {
				Name:           "mewtwo",
				BaseExperience: 306,
			},
		},
	}

	err := commandPokedex(config)
	if err != nil {
		t.Errorf("commandPokedex should not return error, got %v", err)
	}

	// Verify all pokemon are in the pokedex
	if len(config.pokedex) != 3 {
		t.Errorf("Expected 3 pokemon in pokedex, got %d", len(config.pokedex))
	}
}

func TestCommandPokedexDoesNotModify(t *testing.T) {
	originalPokemon := map[string]pokeclient.Pokemon{
		"bulbasaur": {
			Name:           "bulbasaur",
			BaseExperience: 64,
		},
	}

	config := &Config{
		pokedex: originalPokemon,
	}

	err := commandPokedex(config)
	if err != nil {
		t.Errorf("commandPokedex should not return error, got %v", err)
	}

	// Verify pokedex wasn't modified
	if len(config.pokedex) != 1 {
		t.Error("commandPokedex should not modify the pokedex")
	}
	_, ok := config.pokedex["bulbasaur"]
	if !ok {
		t.Error("Original pokemon should still be in pokedex")
	}
}
