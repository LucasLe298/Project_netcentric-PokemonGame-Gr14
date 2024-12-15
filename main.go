package main

import (
	"PROJECT_NETCENTRIC/pokebat"
	"PROJECT_NETCENTRIC/pokecat"
	"PROJECT_NETCENTRIC/pokedex"
	"strings"
	"github.com/gocolly/colly"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	// Register API endpoints
	http.HandleFunc("/api/pokedex", enableCORS(getPokedexHandler))
	http.HandleFunc("/api/battle", enableCORS(pokebat.BattleHandler))
	http.HandleFunc("/api/addPlayer", enableCORS(pokecat.AddPlayerHandler))
	http.HandleFunc("/api/movePlayer", enableCORS(pokecat.MovePlayerHandler))
	http.HandleFunc("/api/worldState", enableCORS(pokecat.WorldStateHandler))
	http.HandleFunc("/api/levelUp", enableCORS(levelUpHandler))
	http.HandleFunc("/api/destroy", enableCORS(destroyHandler))
	http.HandleFunc("/api/spawnPokemons", enableCORS(spawnPokemonsHandler))
	http.HandleFunc("/api/switchPokemon", enableCORS(switchPokemonHandler))
	http.HandleFunc("/api/players", enableCORS(getPlayersHandler))
	http.HandleFunc("/api/addPlayerData", enableCORS(addPlayerData))
	
	

	// Initialize BattleData
	initBattleData()
	initWorld()

	// Check if the Pokedex JSON file exists
	if _, err := os.Stat("pokedex.json"); os.IsNotExist(err) {
		fmt.Println("Crawling data...")
		url := "https://pokemondb.net/pokedex/national"
		pokemons, err := pokedex.CrawlPokemonData(url)
		if err != nil {
			log.Printf("Crawling failed: %v. Starting the server without Pokémon data.", err)
			// Allow server to run even if crawling fails
		} else {
			// Save fetched data
			err = pokedex.SaveToJSON("pokedex.json", pokemons)
			if err != nil {
				log.Printf("Failed to save Pokémon data to JSON: %v", err)
			} else {
				fmt.Println("Pokémon data crawled and saved successfully!")
			}
		}
		// Start the PokeCat spawner
		pokecat.World.StartSpawner()
	}

	////////////////////////////////////////////////////////////////////////////////////////////////
	// Kiểm tra hoặc tạo tệp `battle_data.json`
	filePath := "battle_data.json"
	battleData, err := pokebat.LoadOrCreateBattleData(filePath)
	if err != nil {
		log.Fatalf("Không thể tải hoặc tạo dữ liệu trận đấu: %v", err)
	}

	// Gán dữ liệu vào biến toàn cục BattleData
	pokebat.BattleData = battleData
	log.Println("Battle Data is loaded successfully!")

	/////////////////////////////////////////////////////////////////////////////////////////////
	// Check if the Pokedex JSON file exists
    if _, err := os.Stat("pokedex.json"); os.IsNotExist(err) {
        fmt.Println("Crawling data...")
        url := "https://pokemondb.net/pokedex/national"
        pokemons, err := pokedex.CrawlPokemonData(url)
        if err != nil {
            log.Printf("Crawling failed: %v. Starting the server without Pokémon data.", err)
            // Allow server to run even if crawling fails
        } else {
            // Save fetched data
            err = pokedex.SaveToJSON("pokedex.json", pokemons)
            if err != nil {
                log.Printf("Failed to save Pokémon data to JSON: %v", err)
            } else {
                fmt.Println("Pokémon data crawled and saved successfully!")
            }
        }
    }

    // Update Pokémon images in the Pokedex
    updatePokemonImages()
	
	// Run the server
	fmt.Println("Server is running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// enableCORS adds CORS headers to API responses
func enableCORS(h http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
        if r.Method == "OPTIONS" {
            return
        }
        h.ServeHTTP(w, r)
    }
}

// Function to crawl images and update the Pokedex
func updatePokemonImages() {
    // Load existing Pokedex
    pokemons, err := pokedex.LoadPokedex("pokedex.json")
    if err != nil {
        log.Printf("Failed to load Pokedex: %v", err)
        return
    }

    // Use Colly to crawl image URLs
    url := "https://pokemondb.net/pokedex/national"
    c := colly.NewCollector()

    // Map to store name-image mapping
    imageMap := make(map[string]string)

    // Scrape each Pokémon name and image
    c.OnHTML(".infocard", func(e *colly.HTMLElement) {
        name := e.ChildText(".ent-name")
        image := e.ChildAttr("img", "src")
        if name != "" && image != "" {
            imageMap[strings.ToLower(name)] = image
        }
    })

    // Start crawling
    fmt.Println("Crawling Pokémon images...")
    err = c.Visit(url)
    if err != nil {
        log.Printf("Error during crawling: %v", err)
        return
    }

    // Update images in Pokedex
    for i := range pokemons {
        name := strings.ToLower(pokemons[i].Name)
        if imgURL, exists := imageMap[name]; exists {
            pokemons[i].Image = imgURL
            log.Printf("Updated image for Pokémon: %s", pokemons[i].Name)
        } else {
            log.Printf("No image found for Pokémon: %s", pokemons[i].Name)
        }
    }

    // Save updated Pokedex
    err = pokedex.SaveToJSON("pokedex.json", pokemons)
    if err != nil {
        log.Printf("Failed to save updated Pokedex: %v", err)
    } else {
        fmt.Println("Pokedex updated with images successfully!")
    }
}

func initBattleData() {
	// Initialize BattleData with default players
	pokebat.BattleData = pokebat.Battle{
		PlayerA: pokebat.Player{Name: "PlayerA", Pokemons: []pokedex.Pokemon{}, Active: 0},
		PlayerB: pokebat.Player{Name: "PlayerB", Pokemons: []pokedex.Pokemon{}, Active: 0},
		Turn:    0,
	}
}

func initWorld() {
	pokecat.World = pokecat.Pokeworld{
		Players:  make(map[string]*pokecat.Player),
		Pokemons: []pokecat.PokemonInWorld{},
	}
}

// getPlayersHandler retrieves player information and saves it to a JSON file
func getPlayersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Prepare response data
	response := map[string]interface{}{
		"player_a": pokebat.BattleData.PlayerA,
		"player_b": pokebat.BattleData.PlayerB,
	}

	// Save the players' data to a JSON file
	err := savePlayersToJSON("players_data.json", response)
	if err != nil {
		http.Error(w, "Error saving players data to JSON file", http.StatusInternalServerError)
		log.Printf("Error saving players data to JSON file: %v", err)
		return
	}

	// Return the players' data in the response
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		log.Printf("Error encoding players information response: %v", err)
	}
}


//Get Pokemon's Information
func getPokedexHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request to fetch Pokedex") // Add log
	file, err := os.Open("pokedex.json")
	if err != nil {
		log.Printf("Error reading Pokedex file: %v", err) // Add log
		http.Error(w, "Error reading Pokedex file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Decode data from JSON file
	var pokemons []pokedex.Pokemon
	err = json.NewDecoder(file).Decode(&pokemons)
	if err != nil {
		http.Error(w, "Error decoding Pokedex data", http.StatusInternalServerError)
		return
	}

	// Return JSON data
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pokemons)
}

////////////////////////////////////////////////////////////////
// addPlayerData handles adding a new player's information to players_data.json
func addPlayerData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var newPlayer struct {
		Name     string           `json:"name"`
		Pokemons []pokedex.Pokemon `json:"pokemons"`
	}
	err := json.NewDecoder(r.Body).Decode(&newPlayer)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		log.Printf("Error decoding request body: %v", err)
		return
	}

	playersData := map[string]interface{}{}
	file, err := os.Open("players_data.json")
	if err == nil {
		defer file.Close()
		json.NewDecoder(file).Decode(&playersData)
	} else {
		log.Printf("Creating new players_data.json: %v", err)
	}

	playerKey := fmt.Sprintf("player_%d", len(playersData)+1)
	playersData[playerKey] = newPlayer

	err = savePlayersToJSON("players_data.json", playersData)
	if err != nil {
		http.Error(w, "Error saving players data to JSON file", http.StatusInternalServerError)
		log.Printf("Error saving players data: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"message": fmt.Sprintf("Player %s added successfully!", newPlayer.Name),
	}
	json.NewEncoder(w).Encode(response)
}

func savePlayersToJSON(filename string, data map[string]interface{}) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}
///////////////////////////////////////////////////////////////

// spawnPokemonsHandler spawns new Pokémon in the world
func spawnPokemonsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	pokecat.World.SpawnPokemons()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Pokémon spawned successfully!",
	})
}



// switchPokemonHandler handles Pokémon switching during a battle
func switchPokemonHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		PlayerName     string `json:"player_name"`
		NewPokemonName string `json:"new_pokemon_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid Data", http.StatusBadRequest)
		log.Printf("Error: %v", err)
		return
	}

	log.Printf("request for switching Pokemon: player_name=%s, new_pokemon_name=%s", req.PlayerName, req.NewPokemonName)

	var player *pokebat.Player
	if req.PlayerName == "PlayerA" {
		player = &pokebat.BattleData.PlayerA
	} else if req.PlayerName == "PlayerB" {
		player = &pokebat.BattleData.PlayerB
	} else {
		http.Error(w, "Cannot Find Player", http.StatusBadRequest)
		log.Printf("Cannot Find Player: %s", req.PlayerName)
		return
	}

	err := pokebat.HandlePlayerSwitch(player, req.NewPokemonName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("Error Switching Pokémon: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"message": fmt.Sprintf("The Player %s switches to Pokémon %s", req.PlayerName, req.NewPokemonName),
	}
	json.NewEncoder(w).Encode(response)
}

// levelUpHandler handles requests to level up a Pokémon
func levelUpHandler(w http.ResponseWriter, r *http.Request) {
	// Kiểm tra phương thức
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Decode body
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("Invalid request body:", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Println("Received level-up request for Pokémon:", req.Name)

	// Load Pokedex
	pokemons, err := pokedex.LoadPokedex("pokedex.json")
	if err != nil {
		log.Println("Error loading Pokedex:", err)
		http.Error(w, "Error loading Pokedex", http.StatusInternalServerError)
		return
	}

	// Tìm Pokémon cần tăng cấp
	found := false
	for i := range pokemons {
		if pokemons[i].Name == req.Name {
			log.Println("Found Pokémon:", req.Name)
			pokemons[i].LevelUp() 
			found = true
			break
		}
	}


	if !found {
		log.Println("Pokémon not found:", req.Name)
		http.Error(w, "Pokémon not found", http.StatusNotFound)
		return
	}

	// Ghi lại Pokedex sau khi cập nhật
	if err := pokedex.SaveToJSON("pokedex.json", pokemons); err != nil {
		log.Println("Error saving updated Pokedex:", err)
		http.Error(w, "Error saving Pokedex", http.StatusInternalServerError)
		return
	}

	// Phản hồi thành công
	response := map[string]string{
		"message": req.Name + " leveled up successfully!",
	}
	w.Header().Set("Content-Type", "application/json")
	log.Println(response["message"])
	json.NewEncoder(w).Encode(response)
}

// Destroy a pokemon
func destroyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Decode request body
	var req struct {
		Source string `json:"source"`
		Target string `json:"target"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Received destroy request: source=%s, target=%s", req.Source, req.Target)

	// Load Pokedex
	pokemons, err := pokedex.LoadPokedex("pokedex.json")
	if err != nil {
		http.Error(w, "Error loading Pokedex", http.StatusInternalServerError)
		return
	}

	// Find source and target Pokémon
	var source, target *pokedex.Pokemon
	for i := range pokemons {
		if pokemons[i].Name == req.Source {
			source = &pokemons[i]
		} else if pokemons[i].Name == req.Target {
			target = &pokemons[i]
		}
	}

	// Check if both Pokémon exist
	if source == nil {
		http.Error(w, "Source Pokémon not found", http.StatusNotFound)
		return
	}
	if target == nil {
		http.Error(w, "Target Pokémon not found", http.StatusNotFound)
		return
	}

	// Result message
	var message string

	// Destroy logic
	if len(source.Type) > 0 && len(target.Type) > 0 && source.Type[0] == target.Type[0] {
		// If same type, transfer experience and level up
		target.AccumulatedExp += source.AccumulatedExp
		source.AccumulatedExp = 0
		target.LevelUp()
		message = fmt.Sprintf("%s defeated %s and gained experience, leveling up!", target.Name, source.Name)
		log.Println(message)
	} else {
		// If different types, only defeat
		message = fmt.Sprintf("%s defeated %s. No experience transferred.", target.Name, source.Name)
		log.Println(message)
	}

	// Save updated Pokedex
	if err := pokedex.SaveToJSON("pokedex.json", pokemons); err != nil {
		log.Println("Error saving updated Pokedex:", err)
		http.Error(w, "Error saving Pokedex", http.StatusInternalServerError)
		return
	}

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"message": message,
	}
	json.NewEncoder(w).Encode(response)
}
