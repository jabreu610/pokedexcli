package pokeclient_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jabreu610/pokedexcli/internal/pokecache"
	"github.com/jabreu610/pokedexcli/internal/pokeclient"
)

func TestGetLocationAreas(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse string
		serverStatus   int
		expectError    bool
		expectedCount  int
		checkNext      bool
		expectedNext   *string
	}{
		{
			name: "successful response with results",
			serverResponse: `{
				"count": 2,
				"next": "https://example.com/next",
				"previous": null,
				"results": [
					{"name": "canalave-city-area", "url": "https://pokeapi.co/api/v2/location-area/1/"},
					{"name": "eterna-city-area", "url": "https://pokeapi.co/api/v2/location-area/2/"}
				]
			}`,
			serverStatus:  http.StatusOK,
			expectError:   false,
			expectedCount: 2,
			checkNext:     true,
		},
		{
			name: "empty results",
			serverResponse: `{
				"count": 0,
				"next": null,
				"previous": null,
				"results": []
			}`,
			serverStatus:  http.StatusOK,
			expectError:   false,
			expectedCount: 0,
		},
		{
			name:           "server error",
			serverResponse: "",
			serverStatus:   http.StatusInternalServerError,
			expectError:    true,
		},
		{
			name:           "invalid json",
			serverResponse: `{invalid json}`,
			serverStatus:   http.StatusOK,
			expectError:    true,
		},
		{
			name:           "malformed json structure",
			serverResponse: `{"count": "not a number"}`,
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

			result, err := pokeclient.GetLocationAreas(server.URL, cache)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			if !tt.expectError {
				if len(result.Results) != tt.expectedCount {
					t.Errorf("Expected %d results, got %d", tt.expectedCount, len(result.Results))
				}
				if tt.checkNext && result.Next == nil {
					t.Error("Expected Next to be set, but it was nil")
				}
			}
		})
	}
}

func TestGetLocationAreasCaching(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		response := `{
			"count": 1,
			"next": null,
			"previous": null,
			"results": [{"name": "test-area", "url": "https://example.com/1"}]
		}`
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	cache := pokecache.NewCache(5*time.Second, context.Background())
	defer cache.Close()

	// First call - should hit the server
	_, err := pokeclient.GetLocationAreas(server.URL, cache)
	if err != nil {
		t.Fatalf("First call failed: %v", err)
	}

	if callCount != 1 {
		t.Errorf("Expected 1 server call, got %d", callCount)
	}

	// Second call - should use cache
	_, err = pokeclient.GetLocationAreas(server.URL, cache)
	if err != nil {
		t.Fatalf("Second call failed: %v", err)
	}

	if callCount != 1 {
		t.Errorf("Expected cache to be used (1 server call), got %d calls", callCount)
	}
}

func TestGetLocationAreasNilCache(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"count": 1,
			"next": null,
			"previous": null,
			"results": [{"name": "test-area", "url": "https://example.com/1"}]
		}`
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	// Should handle nil cache gracefully
	result, err := pokeclient.GetLocationAreas(server.URL, nil)
	if err != nil {
		t.Fatalf("Expected no error with nil cache, got %v", err)
	}

	if len(result.Results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(result.Results))
	}
}

func TestGetLocationAreasResponseFields(t *testing.T) {
	nextURL := "https://example.com/next"
	prevURL := "https://example.com/prev"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"count": 1234,
			"next": "https://example.com/next",
			"previous": "https://example.com/prev",
			"results": [
				{"name": "test-location", "url": "https://example.com/location"}
			]
		}`
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	cache := pokecache.NewCache(5*time.Second, context.Background())
	defer cache.Close()

	result, err := pokeclient.GetLocationAreas(server.URL, cache)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Count != 1234 {
		t.Errorf("Expected count 1234, got %d", result.Count)
	}

	if result.Next == nil || *result.Next != nextURL {
		t.Errorf("Expected Next to be %s, got %v", nextURL, result.Next)
	}

	if result.Previous == nil || *result.Previous != prevURL {
		t.Errorf("Expected Previous to be %s, got %v", prevURL, result.Previous)
	}

	if len(result.Results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(result.Results))
	}

	if result.Results[0].Name != "test-location" {
		t.Errorf("Expected name 'test-location', got %s", result.Results[0].Name)
	}

	if result.Results[0].Url != "https://example.com/location" {
		t.Errorf("Expected url 'https://example.com/location', got %s", result.Results[0].Url)
	}
}
