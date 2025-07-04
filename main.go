package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PancakeMash/pokedexcli/internal/pokecache"
)

// Init the cliCommand at the package level so that callback functions can access it.
var cli_cmd map[string]cliCommand
var caught_pkm map[string]Pokemon

func main() {

	cache := pokecache.NewCache(10 * time.Second)
	caught_pkm = make(map[string]Pokemon)

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
			description: "Visit a specific location by name. Usage: explore <location_name>",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch",
			description: "Attempt to catch a Pokemon by name. Usage: catch <pokemon_name>",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect",
			description: "Inspect a caught Pokemon by name. Usage: inspect <pokemon_name>",
			callback:    commandInspect,
		},
		"pokedex": {
			name:        "pokedex",
			description: "Shows a list of caught Pokemon.",
			callback:    commandPokedex,
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
		if len(input) == 2 {
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

func randRange(min, max int) int {
	return rand.IntN(max-min) + min
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

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d: %s", res.StatusCode, res.Status)
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
		return err
	}

	var EncounterData Encounter
	if err := json.Unmarshal(body, &EncounterData); err != nil {
		return err
	}

	//Display the Pokemon found in the location
	for _, encounter := range EncounterData.PokemonEncounters {
		fmt.Println(encounter.Pokemon.Name)
	}

	return nil
}

func commandCatch(input string, config *Config, cache *pokecache.Cache) error {
	if input == "" {
		fmt.Println("Please provide a Pokemon name to catch.")
		return nil
	}
	url := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s", input)
	fmt.Printf("Throwing a Pokeball at %s...\n", input)

	fmt.Println("Debug: About to fetch data")
	body, err := fetchLocationData(url, cache)
	if err != nil {
		fmt.Printf("Error: %s not found or could not be retrieved.\n", input)
		return nil
	}

	fmt.Println("Debug: Data fetched successfully") // Add this
	var PokemonData Pokemon
	fmt.Println("Debug: About to unmarshal JSON") // Add this
	if err := json.Unmarshal(body, &PokemonData); err != nil {
		fmt.Printf("Debug: JSON unmarshal error: %v\n", err) // Add this
		return err
	}

	fmt.Printf("Debug: %s BaseEXP=%d\n", PokemonData.Name, PokemonData.BaseEXP)

	randRange := randRange(1, 170)
	if randRange >= PokemonData.BaseEXP/2 {
		caught_pkm[PokemonData.Name] = PokemonData
		fmt.Printf("Caught %s!\n", PokemonData.Name)
	} else {
		fmt.Printf("%s escaped!\n", PokemonData.Name)
	}

	return nil
}

func commandInspect(input string, config *Config, cache *pokecache.Cache) error {
	if input == "" {
		fmt.Println("Please provide a Pokemon name to inspect.")
		return nil
	}

	if pokemon, found := caught_pkm[input]; found {
		fmt.Printf("Name: %s \n", pokemon.Name)
		fmt.Printf("Height: %d\n", pokemon.Height)
		fmt.Printf("Weight: %d\n", pokemon.Weight)
		fmt.Println("Stats:")
		for _, stat := range pokemon.Stats {
			fmt.Printf("  %s: %d\n", stat.Stat.Name, stat.BaseStat)
		}
		fmt.Println("Types:")
		for _, typeInfo := range pokemon.Types {
			fmt.Printf("  - %s\n", typeInfo.Type.Name)
		}
	} else {
		fmt.Println("You have not caught that Pokemon yet.")
	}

	return nil
}

func commandPokedex(input string, config *Config, cache *pokecache.Cache) error {
	if len(caught_pkm) == 0 {
		fmt.Println("You haven't caught any Pokemon yet.")
		return nil
	}
	fmt.Println("Your Pokedex:")
	for _, pokemon := range caught_pkm {
		fmt.Printf("  - %s\n", pokemon.Name)
	}

	return nil
}

type LocationArea struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Pokemon struct {
	Name    string `json:"name"`
	BaseEXP int    `json:"base_experience"`
	Height  int    `json:"height"`
	Weight  int    `json:"weight"`
	Stats   []struct {
		BaseStat int `json:"base_stat"`
		Stat     struct {
			Name string `json:"name"`
		} `json:"stat"`
	} `json:"stats"`
	Types []struct {
		Type struct {
			Name string `json:"name"`
		} `json:"type"`
	} `json:"types"`
}

type Encounter struct {
	PokemonEncounters []struct {
		Pokemon Pokemon `json:"pokemon"`
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
