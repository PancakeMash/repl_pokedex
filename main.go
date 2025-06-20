package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

// Init the cliCommand at the package level so that callback functions can access it.
var cli_cmd map[string]cliCommand

func main() {

	config := &Config{}

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
		"map": {
			name:        "map",
			description: "Shows the list of 20 locations. Repeat the command to see the next 20 locations.",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Shows the previous 20 locations. Use this to go back in the list of locations.",
			callback:    commandMapBack,
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
		cmd.callback(config)

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

func commandExit(config *Config) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(config *Config) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	for _, cmd := range cli_cmd {
		fmt.Printf("%s: %s\n", cmd.name, cmd.description)
	}

	return nil
}

func commandMap(config *Config) error {
	var url string
	if config.nextURL == "" {
		url = "https://pokeapi.co/api/v2/location-area/"
	} else {
		url = config.nextURL
	}

	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	var PokemonLocation LocationAreaResponse
	if err := json.Unmarshal(body, &PokemonLocation); err != nil {
		log.Fatal(err)
	}

	config.nextURL = PokemonLocation.Next
	config.previousURL = PokemonLocation.Previous

	for _, locationarea := range PokemonLocation.Results {
		fmt.Println(locationarea.Name)
	}

	return nil
}

func commandMapBack(config *Config) error {
	var url string
	if config.previousURL == "" {
		fmt.Println("you're on the first page")
		return nil
	} else {
		url = config.previousURL
	}

	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	var PokemonLocation LocationAreaResponse
	if err := json.Unmarshal(body, &PokemonLocation); err != nil {
		log.Fatal(err)
	}

	config.nextURL = PokemonLocation.Next
	config.previousURL = PokemonLocation.Previous

	for _, locationarea := range PokemonLocation.Results {
		fmt.Println(locationarea.Name)
	}
	return nil
}

type LocationArea struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Config struct {
	nextURL     string
	previousURL string
}

type LocationAreaResponse struct {
	Count    int            `json:"count"`
	Next     string         `json:"next"`
	Previous string         `json:"previous"`
	Results  []LocationArea `json:"results"`
}

type cliCommand struct {
	name        string
	description string
	callback    func(*Config) error
}
