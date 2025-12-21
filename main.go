package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/jabreu610/pokedexcli/repl"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Pokedex > ")
		scanner.Scan()
		cleaned := repl.CleanInput(scanner.Text())
		echo := fmt.Sprintf("Your command was: %s", cleaned[0])
		fmt.Println(echo)
	}
}
