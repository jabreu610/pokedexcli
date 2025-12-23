package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/jabreu610/pokedexcli/pokeclient"
	"github.com/jabreu610/pokedexcli/repl"
)

type Config struct {
	next *string
	prev *string
}

type cliCommand struct {
	Name        string
	Description string
	Callback    func(*Config) error
}

var commands map[string]cliCommand

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
	if c.next != nil {
		url = *c.next
	}
	res, err := pokeclient.GetLocationAreas(url)
	if err != nil {
		return err
	}
	c.prev = res.Previous
	c.next = res.Next
	for _, locArea := range res.Results {
		fmt.Println(locArea.Name)
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
	}
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	config := Config{}
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
		if !ok {
			fmt.Println("Unknown command")
			continue
		}
		if err := command.Callback(&config); err != nil {
			fmt.Println(err)
		}
	}
}
