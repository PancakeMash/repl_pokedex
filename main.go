package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Init the cliCommand at the package level so that callback functions can access it.
var cli_cmd map[string]cliCommand

func main() {

	cli_cmd = map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"help": {
			name:        "help",
			description: "Show this help message",
			callback:    commandHelp,
		},
	}

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Pokedex > ")
		scanner.Scan()
		input := cleanInput(scanner.Text())
		if len(input) == 0 {
			continue
		}
		first_word := input[0]
		cmd, ok := cli_cmd[first_word]
		if !ok {
			fmt.Println("Unknown command")
			continue
		}
		cmd.callback()

	}

}

func cleanInput(text string) []string {

	words := []string{}
	text_split := strings.Split(text, " ")
	for _, word := range text_split {
		word = strings.TrimSpace(word)
		if word != "" {
			words = append(words, word)
		}
	}

	return words
}

func commandExit() error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp() error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	for _, cmd := range cli_cmd {
		fmt.Printf("%s: %s\n", cmd.name, cmd.description)
	}

	return nil
}

type cliCommand struct {
	name        string
	description string
	callback    func() error
}
