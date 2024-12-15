package pokedex

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"github.com/PuerkitoBio/goquery"
)

type Pokemon struct {
	Name            string   `json:"name"`
	Image          string   `json:"image"`
	Type            []string `json:"type"`
	Level           int      `json:"level"`
	AccumulatedExp  int      `json:"accumulated_exp"`
	Attack          int      `json:"attack"`
	Defense         int      `json:"defense"`
	Speed           int      `json:"speed"`
	SpecialAttack   int      `json:"special_attack"`
	SpecialDefense  int      `json:"special_defense"`
	HP              int      `json:"hp"`
	EV              float64  `json:"ev"`
}

type PokemonStats struct {
	HP             int `json:"hp"`
	Attack         int `json:"attack"`
	Defense        int `json:"defense"`
	SpecialAttack  int `json:"special_attack"`
	SpecialDefense int `json:"special_defense"`
	Speed          int `json:"speed"`
}

// LoadPokedex reads pokedex data from JSON file
func LoadPokedex(filePath string) ([]Pokemon, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var pokedex []Pokemon
	err = json.NewDecoder(file).Decode(&pokedex)
	return pokedex, err
}

// SaveToJSON saves data to a JSON file
func SaveToJSON(filename string, data []Pokemon) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// CrawlPokemonData fetches Pokémon data from the given URL
func CrawlPokemonData(url string) ([]Pokemon, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 HTTP status: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var pokemons []Pokemon

	// Updated selector for Pokémon cards
	doc.Find(".infocard").Each(func(i int, card *goquery.Selection) {
		// Extract Pokémon name
		name := card.Find(".ent-name").Text()
		if name == "" {
			log.Printf("Skipping card %d with no name", i)
			return
		}

		// Extract Pokémon types
		var types []string
		card.Find(".itype").Each(func(j int, typeSel *goquery.Selection) {
			types = append(types, typeSel.Text())
		})

		// Extract Pokémon image URL
		imageURL, exists := card.Find("img").Attr("src")
		if !exists {
			log.Printf("No image URL found for Pokémon: %s", name)
			imageURL = "" // Leave empty if no image found
		} else {
			log.Printf("Found image URL for Pokémon %s: %s", name, imageURL)
		}

		// Get the detailed stats URL
		statsURL, exists := card.Find(".ent-name").Attr("href")
		if !exists {
			log.Printf("No stats URL for Pokémon: %s", name)
			return
		}

		// Append the base URL to the relative stats URL
		statsURL = "https://pokemondb.net" + statsURL

		// Fetch stats from the detailed page
		stats, err := GetPokemonStats(statsURL)
		if err != nil {
			log.Printf("Failed to fetch stats for Pokémon %s: %v", name, err)
			return
		}

		// Append the Pokémon to the list
		pokemons = append(pokemons, Pokemon{
			Name:           name,
			Image:          imageURL,
			Type:           types,
			Level:          1, 
			HP:             stats.HP,
			Attack:         stats.Attack,
			Defense:        stats.Defense,
			SpecialAttack:  stats.SpecialAttack,
			SpecialDefense: stats.SpecialDefense,
			Speed:          stats.Speed,
		})
		log.Printf("Successfully added Pokémon: %s", name)
	})

	if len(pokemons) == 0 {
		return nil, fmt.Errorf("no Pokémon data found, check the CSS selectors or website structure")
	}

	return pokemons, nil
}

// GetPokemonStats fetches base stats for a Pokémon from its stats page
func GetPokemonStats(url string) (PokemonStats, error) {
	resp, err := http.Get(url)
	if err != nil {
		return PokemonStats{}, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return PokemonStats{}, fmt.Errorf("received non-200 HTTP status: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return PokemonStats{}, fmt.Errorf("failed to parse HTML: %w", err)
	}

	stats := PokemonStats{}

	// Updated selector for the stats table
	doc.Find(".vitals-table tbody tr").Each(func(index int, row *goquery.Selection) {
		statName := strings.TrimSpace(row.Find("th").Text())
		statValue := strings.TrimSpace(row.Find("td.cell-num").First().Text())

		// Skip non-numeric rows
		if statValue == "" {
			log.Printf("Skipping stat: %s (non-numeric value)", statName)
			return
		}

		// Convert statValue to an integer
		value, err := strconv.Atoi(statValue)
		if err != nil {
			log.Printf("Failed to parse stat value for %s: %v", statName, err)
			return
		}

		// Map the stat name to the appropriate field in PokemonStats
		switch statName {
		case "HP":
			stats.HP = value
		case "Attack":
			stats.Attack = value
		case "Defense":
			stats.Defense = value
		case "Sp. Atk":
			stats.SpecialAttack = value
		case "Sp. Def":
			stats.SpecialDefense = value
		case "Speed":
			stats.Speed = value
		default:
			log.Printf("Skipping unknown stat: %s", statName)
		}
	})

	if stats == (PokemonStats{}) {
		return stats, fmt.Errorf("no valid stats found for URL: %s", url)
	}

	return stats, nil
}

// LevelUp increases a Pokémon's level based on accumulated experience
func (p *Pokemon) LevelUp() {
	requiredExp := p.Level * 2
	for p.AccumulatedExp >= requiredExp {
		p.Level++
		p.AccumulatedExp -= requiredExp

		p.Attack = int(float64(p.Attack) * (1 + p.EV))
		p.Defense = int(float64(p.Defense) * (1 + p.EV))
		p.SpecialAttack = int(float64(p.SpecialAttack) * (1 + p.EV))
		p.SpecialDefense = int(float64(p.SpecialDefense) * (1 + p.EV))
		p.HP = int(float64(p.HP) * (1 + p.EV))

		requiredExp = p.Level * 2
	}
}

func (p *Pokemon) Destroy(target *Pokemon) error {
	// Nếu cùng loại (Type), chuyển kinh nghiệm và tăng cấp
	if len(p.Type) > 0 && len(target.Type) > 0 && p.Type[0] == target.Type[0] {
		target.AccumulatedExp += p.AccumulatedExp
		p.AccumulatedExp = 0
		target.LevelUp() // Tăng cấp Pokémon mục tiêu
		log.Printf("%s transferred experience to %s and leveled up!", p.Name, target.Name)
	} else {
		// Nếu khác loại, chỉ hạ gục Pokémon nguồn
		log.Printf("%s defeated %s. No experience transferred as types differ.", target.Name, p.Name)
	}
	return nil
}
