package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"os"
	"time"

	"github.com/jabreu610/pokedexcli/internal/pokecache"
	"github.com/jabreu610/pokedexcli/internal/pokeclient"
	"github.com/jabreu610/pokedexcli/internal/repl"
)

const defaultInterval = time.Second * 5

type Config struct {
	Next    *string
	Prev    *string
	cache   *pokecache.Cache
	args    []string
	pokedex map[string]pokeclient.Pokemon
}

type cliCommand struct {
	Name        string
	Description string
	Callback    func(*Config) error
}

var commands map[string]cliCommand

func passWithDifficulty(baseExp int) bool {
	// Normalize 36-635 to 0.0-1.0
	normalizedDifficulty := float64(baseExp-36) / float64(635-36)

	// Invert so higher exp = harder (0.9 pass for weakest, 0.1 for strongest)
	passRate := 0.9 - (normalizedDifficulty * 0.8)

	return rand.Float64() < passRate
}

func processLocationAreaResponse(d pokeclient.LocationAreaResponse, c *Config) {
	c.Prev = d.Previous
	c.Next = d.Next
	for _, locArea := range d.Results {
		fmt.Println(locArea.Name)
	}
}

func commandExit(c *Config) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(c *Config) error {
	fmt.Print("Welcome to the Pokedex!\nUsage:\n\n")
	for _, command := range commands {
		helpMessage := fmt.Sprintf("%s: %s", command.Name, command.Description)
		fmt.Println(helpMessage)
	}
	return nil
}

func commandMap(c *Config) error {
	url := pokeclient.BaseUrlLocationArea
	if c.Next != nil {
		url = *c.Next
	}
	res, err := pokeclient.GetLocationAreas(url, c.cache)
	if err != nil {
		return err
	}
	processLocationAreaResponse(res, c)
	return nil
}

func commandMapb(c *Config) error {
	if c.Prev == nil {
		fmt.Println("you're on the first page")
		return nil
	}
	res, err := pokeclient.GetLocationAreas(*c.Prev, c.cache)
	if err != nil {
		return err
	}
	processLocationAreaResponse(res, c)
	return nil
}

func commandExplore(c *Config) error {
	if len(c.args) < 1 {
		return errors.New("Expected one arguement, a location area name. Recieved none")
	}
	pokemon, err := pokeclient.GetPokemonForLocationName(c.args[0], c.cache)
	if err != nil {
		return err
	}
	for _, name := range pokemon {
		fmt.Println(name)
	}
	return nil
}

func commandCatch(c *Config) error {
	if len(c.args) < 1 {
		return errors.New("Expected one arguement, a Pokemon name. Recieved none")
	}
	pokemon, err := pokeclient.GetPokemon(c.args[0], c.cache)
	if errors.Is(err, pokeclient.ErrPokemonNotFound) {
		msg := fmt.Sprintf("Pokemon %s does not exist", c.args[0])
		fmt.Println(msg)
		return nil
	}
	if err != nil {
		return err
	}
	catchIntro := fmt.Sprintf("Throwing a Pokeball at %s...", pokemon.Name)
	fmt.Println(catchIntro)
	caught := passWithDifficulty(pokemon.BaseExperience)
	if caught {
		caughtMsg := fmt.Sprintf("%s was caught!", pokemon.Name)
		fmt.Println(caughtMsg)
		c.pokedex[pokemon.Name] = pokemon
	} else {
		failedMsg := fmt.Sprintf("%s escaped!", pokemon.Name)
		fmt.Println(failedMsg)
	}
	return nil
}

func init() {
	commands = map[string]cliCommand{
		"exit": {
			Name:        "exit",
			Description: "Exit the Pokedex",
			Callback:    commandExit,
		},
		"help": {
			Name:        "help",
			Description: "Displays a help message",
			Callback:    commandHelp,
		},
		"map": {
			Name:        "map",
			Description: "Displays location areas",
			Callback:    commandMap,
		},
		"mapb": {
			Name:        "mapb",
			Description: "Displays the previous page of location areas",
			Callback:    commandMapb,
		},
		"explore": {
			Name:        "explore",
			Description: "List Pokemon for a given location area",
			Callback:    commandExplore,
		},
		"catch": {
			Name:        "catch",
			Description: "Attenpt a Pokemon, expects a pokemon name as an argument",
			Callback:    commandCatch,
		},
	}
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	config := Config{
		cache:   pokecache.NewCache(defaultInterval, context.Background()),
		pokedex: map[string]pokeclient.Pokemon{},
	}
	for {
		fmt.Print("Pokedex > ")
		scanner.Scan()
		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			os.Exit(1)
		}
		cleaned := repl.CleanInput(scanner.Text())
		if len(cleaned) == 0 {
			continue
		}
		command, ok := commands[cleaned[0]]
		if len(cleaned) > 1 {
			config.args = cleaned[1:]
		} else {
			config.args = []string{}
		}
		if !ok {
			fmt.Println("Unknown command")
			continue
		}
		if err := command.Callback(&config); err != nil {
			fmt.Println(err)
		}
	}
}
