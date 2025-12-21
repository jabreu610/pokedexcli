package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/jabreu610/pokedexcli/repl"
)

type cliCommand struct {
	name        string
	description string
	callback    func() error
}

var commands map[string]cliCommand

func commandExit() error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp() error {
	fmt.Print("Welcome to the Pokedex!\nUsage:\n\n")
	for _, command := range commands {
		helpMessage := fmt.Sprintf("%s: %s", command.name, command.description)
		fmt.Println(helpMessage)
	}
	return nil
}

func init() {
	commands = map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
	}
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)

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
		if err := command.callback(); err != nil {
			fmt.Println(err)
		}
	}
}
