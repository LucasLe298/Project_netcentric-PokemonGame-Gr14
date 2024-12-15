package pokecat

import (
	"encoding/json"
	"net/http"
)

// AddPlayerHandler adds a new player to the world
func AddPlayerHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	World.AddPlayer(request.ID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "player added"})
}

// MovePlayerHandler moves a player in the world
func MovePlayerHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ID        string `json:"id"`
		Direction string `json:"direction"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	World.MovePlayer(request.ID, request.Direction)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "player moved"})
}

// WorldStateHandler retrieves the current state of the world
func WorldStateHandler(w http.ResponseWriter, r *http.Request) {
	World.mu.Lock()
	defer World.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(World)
}

func catchPokemonHandler(w http.ResponseWriter, r *http.Request) {
    var request struct {
        PlayerID string `json:"player_id"`
        PokemonName string `json:"pokemon_name"`
    }

    if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    World.mu.Lock()
    defer World.mu.Unlock()

    player, exists := World.Players[request.PlayerID]
    if !exists {
        http.Error(w, "Player not found", http.StatusNotFound)
        return
    }

    // Find Pokémon on the same grid as the player
    for i, pokemon := range World.Pokemons {
        if pokemon.X == player.X && pokemon.Y == player.Y && pokemon.Pokemon.Name == request.PokemonName {
            player.Pokemons = append(player.Pokemons, pokemon.Pokemon)

            // Remove Pokémon from the world
            World.Pokemons = append(World.Pokemons[:i], World.Pokemons[i+1:]...)

            response := map[string]string{"message": "Successfully caught " + request.PokemonName}
            json.NewEncoder(w).Encode(response)
            return
        }
    }

    http.Error(w, "Pokémon not found or not on the same position", http.StatusNotFound)
}