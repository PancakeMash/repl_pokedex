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
	"time"

	"github.com/PancakeMash/pokedexcli/internal/pokecache"
)

// Init the cliCommand at the package level so that callback functions can access it.
var cli_cmd map[string]cliCommand

func main() {

	cache := pokecache.NewCache(10 * time.Second)

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
		"explore": {
			name:        "explore",
			description: "Explore a specific location by name. Usage: explore <location_name>",
			callback:    commandExplore,
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

		name := ""
		if len(input) == 2 && first_word == "explore" {
			name = input[1]
		}

		cmd.callback(name, config, cache)

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

func commandExit(input string, config *Config, cache *pokecache.Cache) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(input string, config *Config, cache *pokecache.Cache) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	for _, cmd := range cli_cmd {
		fmt.Printf("%s: %s\n", cmd.name, cmd.description)
	}

	return nil
}

func fetchLocationData(url string, cache *pokecache.Cache) ([]byte, error) {

	if cachedData, found := cache.Get(url); found {
		fmt.Println("Using cached data!")
		return cachedData, nil
	}

	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	cache.Add(url, body)
	return body, nil
}

func commandMap(input string, config *Config, cache *pokecache.Cache) error {
	var url string
	if config.nextURL == "" {
		url = "https://pokeapi.co/api/v2/location-area/"
	} else {
		url = config.nextURL
	}

	body, err := fetchLocationData(url, cache)
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

func commandMapBack(input string, config *Config, cache *pokecache.Cache) error {
	var url string
	if config.previousURL == "" {
		fmt.Println("you're on the first page")
		return nil
	} else {
		url = config.previousURL
	}

	body, err := fetchLocationData(url, cache)
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

func commandExplore(input string, config *Config, cache *pokecache.Cache) error {
	if input == "" {
		fmt.Println("Please provide a location name to explore.")
		return nil
	}

	fmt.Printf("Exploring %s ...\n", input)
	url := fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%s", input)

	body, err := fetchLocationData(url, cache)
	if err != nil {
		log.Fatal(err)
	}

	var EncounterData Encounter
	if err := json.Unmarshal(body, &EncounterData); err != nil {
		log.Fatal(err)
	}

	//Display the Pokemon found in the location
	for _, encounter := range EncounterData.PokemonEncounters {
		fmt.Println(encounter.Pokemon.Name)
	}

	return nil
}

type LocationArea struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Encounter struct {
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
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
	callback    func(string, *Config, *pokecache.Cache) error
}
