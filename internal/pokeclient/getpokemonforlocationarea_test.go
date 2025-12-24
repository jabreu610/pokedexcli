package pokeclient_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jabreu610/pokedexcli/internal/pokecache"
	"github.com/jabreu610/pokedexcli/internal/pokeclient"
)

func TestGetPokemonForLocationName(t *testing.T) {
	tests := []struct {
		name            string
		locationName    string
		serverResponse  string
		serverStatus    int
		expectError     bool
		expectedPokemon []string
	}{
		{
			name:         "successful response with multiple pokemon",
			locationName: "canalave-city-area",
			serverResponse: `{
				"pokemon_encounters": [
					{"pokemon": {"name": "tentacool"}},
					{"pokemon": {"name": "tentacruel"}},
					{"pokemon": {"name": "staryu"}}
				]
			}`,
			serverStatus:    http.StatusOK,
			expectError:     false,
			expectedPokemon: []string{"tentacool", "tentacruel", "staryu"},
		},
		{
			name:         "empty pokemon list",
			locationName: "empty-area",
			serverResponse: `{
				"pokemon_encounters": []
			}`,
			serverStatus:    http.StatusOK,
			expectError:     false,
			expectedPokemon: []string{},
		},
		{
			name:           "server error",
			locationName:   "error-area",
			serverResponse: "",
			serverStatus:   http.StatusInternalServerError,
			expectError:    true,
		},
		{
			name:           "invalid json",
			locationName:   "invalid-area",
			serverResponse: `{invalid json}`,
			serverStatus:   http.StatusOK,
			expectError:    true,
		},
		{
			name:         "malformed json structure",
			locationName: "malformed-area",
			serverResponse: `{
				"pokemon_encounters": "not an array"
			}`,
			serverStatus: http.StatusOK,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify the URL path includes the location name
				expectedPath := "/" + tt.locationName
				if !strings.HasSuffix(r.URL.Path, expectedPath) {
					t.Errorf("Expected path to end with %s, got %s", expectedPath, r.URL.Path)
				}

				w.WriteHeader(tt.serverStatus)
				w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			cache := pokecache.NewCache(5*time.Second, context.Background())
			defer cache.Close()

			// Replace the base URL for testing
			originalBaseURL := pokeclient.BaseUrlLocationArea
			pokeclient.BaseUrlLocationArea = server.URL
			defer func() {
				pokeclient.BaseUrlLocationArea = originalBaseURL
			}()

			result, err := pokeclient.GetPokemonForLocationName(tt.locationName, cache)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			if !tt.expectError {
				if len(result) != len(tt.expectedPokemon) {
					t.Errorf("Expected %d pokemon, got %d", len(tt.expectedPokemon), len(result))
				}
				for i, pokemon := range tt.expectedPokemon {
					if i >= len(result) {
						break
					}
					if result[i] != pokemon {
						t.Errorf("Expected pokemon[%d] to be %s, got %s", i, pokemon, result[i])
					}
				}
			}
		})
	}
}

func TestGetPokemonForLocationNameCaching(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		response := `{
			"pokemon_encounters": [
				{"pokemon": {"name": "pikachu"}}
			]
		}`
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	cache := pokecache.NewCache(5*time.Second, context.Background())
	defer cache.Close()

	originalBaseURL := pokeclient.BaseUrlLocationArea
	pokeclient.BaseUrlLocationArea = server.URL
	defer func() {
		pokeclient.BaseUrlLocationArea = originalBaseURL
	}()

	// First call - should hit the server
	result1, err := pokeclient.GetPokemonForLocationName("test-area", cache)
	if err != nil {
		t.Fatalf("First call failed: %v", err)
	}

	if callCount != 1 {
		t.Errorf("Expected 1 server call, got %d", callCount)
	}

	// Second call - should use cache
	result2, err := pokeclient.GetPokemonForLocationName("test-area", cache)
	if err != nil {
		t.Fatalf("Second call failed: %v", err)
	}

	if callCount != 1 {
		t.Errorf("Expected cache to be used (1 server call), got %d calls", callCount)
	}

	// Verify both results are the same
	if len(result1) != len(result2) {
		t.Errorf("Cached result differs from original")
	}
	if len(result1) > 0 && result1[0] != result2[0] {
		t.Errorf("Cached pokemon name differs: %s vs %s", result1[0], result2[0])
	}
}

func TestGetPokemonForLocationNameNilCache(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"pokemon_encounters": [
				{"pokemon": {"name": "charmander"}}
			]
		}`
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	originalBaseURL := pokeclient.BaseUrlLocationArea
	pokeclient.BaseUrlLocationArea = server.URL
	defer func() {
		pokeclient.BaseUrlLocationArea = originalBaseURL
	}()

	// Should handle nil cache gracefully
	result, err := pokeclient.GetPokemonForLocationName("test-area", nil)
	if err != nil {
		t.Fatalf("Expected no error with nil cache, got %v", err)
	}

	if len(result) != 1 {
		t.Errorf("Expected 1 pokemon, got %d", len(result))
	}
	if len(result) > 0 && result[0] != "charmander" {
		t.Errorf("Expected 'charmander', got %s", result[0])
	}
}

func TestGetPokemonForLocationNameURLConstruction(t *testing.T) {
	requestedPath := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.Path
		response := `{"pokemon_encounters": []}`
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	cache := pokecache.NewCache(5*time.Second, context.Background())
	defer cache.Close()

	originalBaseURL := pokeclient.BaseUrlLocationArea
	pokeclient.BaseUrlLocationArea = server.URL
	defer func() {
		pokeclient.BaseUrlLocationArea = originalBaseURL
	}()

	locationName := "viridian-forest"
	_, err := pokeclient.GetPokemonForLocationName(locationName, cache)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedPath := "/" + locationName
	if requestedPath != expectedPath {
		t.Errorf("Expected request path %s, got %s", expectedPath, requestedPath)
	}
}
