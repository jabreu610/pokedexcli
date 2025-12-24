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
	expectedCommands := []string{"exit", "help", "map", "mapb", "explore"}

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
