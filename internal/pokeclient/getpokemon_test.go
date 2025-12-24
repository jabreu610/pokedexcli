package pokeclient_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jabreu610/pokedexcli/internal/pokecache"
	"github.com/jabreu610/pokedexcli/internal/pokeclient"
)

func TestGetPokemon(t *testing.T) {
	tests := []struct {
		name              string
		pokemonName       string
		serverResponse    string
		serverStatus      int
		expectError       bool
		expectedName      string
		expectedBaseExp   int
		expectNotFoundErr bool
	}{
		{
			name:        "successful response",
			pokemonName: "pikachu",
			serverResponse: `{
				"name": "pikachu",
				"base_experience": 112
			}`,
			serverStatus:    http.StatusOK,
			expectError:     false,
			expectedName:    "pikachu",
			expectedBaseExp: 112,
		},
		{
			name:        "pokemon with high base experience",
			pokemonName: "blissey",
			serverResponse: `{
				"name": "blissey",
				"base_experience": 635
			}`,
			serverStatus:    http.StatusOK,
			expectError:     false,
			expectedName:    "blissey",
			expectedBaseExp: 635,
		},
		{
			name:        "pokemon with low base experience",
			pokemonName: "magikarp",
			serverResponse: `{
				"name": "magikarp",
				"base_experience": 40
			}`,
			serverStatus:    http.StatusOK,
			expectError:     false,
			expectedName:    "magikarp",
			expectedBaseExp: 40,
		},
		{
			name:              "pokemon not found - 404",
			pokemonName:       "fakemon",
			serverResponse:    ``,
			serverStatus:      http.StatusNotFound,
			expectError:       true,
			expectNotFoundErr: true,
		},
		{
			name:           "server error",
			pokemonName:    "error-pokemon",
			serverResponse: "",
			serverStatus:   http.StatusInternalServerError,
			expectError:    true,
		},
		{
			name:        "invalid json",
			pokemonName: "invalid",
			serverResponse: `{
				"name": "invalid",
				"base_experience": "not a number"
			}`,
			serverStatus: http.StatusOK,
			expectError:  true,
		},
		{
			name:           "malformed json",
			pokemonName:    "malformed",
			serverResponse: `{invalid json}`,
			serverStatus:   http.StatusOK,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.serverStatus)
				w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			cache := pokecache.NewCache(5*time.Second, context.Background())
			defer cache.Close()

			originalBaseURL := pokeclient.BaseUrlPokemon
			pokeclient.BaseUrlPokemon = server.URL
			defer func() {
				pokeclient.BaseUrlPokemon = originalBaseURL
			}()

			result, err := pokeclient.GetPokemon(tt.pokemonName, cache)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			if tt.expectNotFoundErr {
				if !errors.Is(err, pokeclient.ErrPokemonNotFound) {
					t.Errorf("Expected ErrPokemonNotFound, got %v", err)
				}
			}
			if !tt.expectError {
				if result.Name != tt.expectedName {
					t.Errorf("Expected name %s, got %s", tt.expectedName, result.Name)
				}
				if result.BaseExperience != tt.expectedBaseExp {
					t.Errorf("Expected base experience %d, got %d", tt.expectedBaseExp, result.BaseExperience)
				}
			}
		})
	}
}

func TestGetPokemonCaching(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		response := `{
			"name": "charizard",
			"base_experience": 240
		}`
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	cache := pokecache.NewCache(5*time.Second, context.Background())
	defer cache.Close()

	originalBaseURL := pokeclient.BaseUrlPokemon
	pokeclient.BaseUrlPokemon = server.URL
	defer func() {
		pokeclient.BaseUrlPokemon = originalBaseURL
	}()

	// First call - should hit the server
	result1, err := pokeclient.GetPokemon("charizard", cache)
	if err != nil {
		t.Fatalf("First call failed: %v", err)
	}

	if callCount != 1 {
		t.Errorf("Expected 1 server call, got %d", callCount)
	}

	// Second call - should use cache
	result2, err := pokeclient.GetPokemon("charizard", cache)
	if err != nil {
		t.Fatalf("Second call failed: %v", err)
	}

	if callCount != 1 {
		t.Errorf("Expected cache to be used (1 server call), got %d calls", callCount)
	}

	// Verify both results are the same
	if result1.Name != result2.Name {
		t.Errorf("Cached name differs: %s vs %s", result1.Name, result2.Name)
	}
	if result1.BaseExperience != result2.BaseExperience {
		t.Errorf("Cached base experience differs: %d vs %d", result1.BaseExperience, result2.BaseExperience)
	}
}

func TestGetPokemonNilCache(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"name": "bulbasaur",
			"base_experience": 64
		}`
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	originalBaseURL := pokeclient.BaseUrlPokemon
	pokeclient.BaseUrlPokemon = server.URL
	defer func() {
		pokeclient.BaseUrlPokemon = originalBaseURL
	}()

	// Should handle nil cache gracefully
	result, err := pokeclient.GetPokemon("bulbasaur", nil)
	if err != nil {
		t.Fatalf("Expected no error with nil cache, got %v", err)
	}

	if result.Name != "bulbasaur" {
		t.Errorf("Expected name 'bulbasaur', got %s", result.Name)
	}
	if result.BaseExperience != 64 {
		t.Errorf("Expected base experience 64, got %d", result.BaseExperience)
	}
}

func TestGetPokemonURLConstruction(t *testing.T) {
	requestedPath := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.Path
		response := `{"name": "test", "base_experience": 100}`
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	cache := pokecache.NewCache(5*time.Second, context.Background())
	defer cache.Close()

	originalBaseURL := pokeclient.BaseUrlPokemon
	pokeclient.BaseUrlPokemon = server.URL
	defer func() {
		pokeclient.BaseUrlPokemon = originalBaseURL
	}()

	pokemonName := "mewtwo"
	_, err := pokeclient.GetPokemon(pokemonName, cache)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedPath := "/" + pokemonName
	if requestedPath != expectedPath {
		t.Errorf("Expected request path %s, got %s", expectedPath, requestedPath)
	}
}

func TestErrPokemonNotFound(t *testing.T) {
	// Verify the error is defined and can be used with errors.Is
	err := pokeclient.ErrPokemonNotFound
	if err == nil {
		t.Error("ErrPokemonNotFound should not be nil")
	}

	if !errors.Is(err, pokeclient.ErrPokemonNotFound) {
		t.Error("errors.Is should work with ErrPokemonNotFound")
	}
}
