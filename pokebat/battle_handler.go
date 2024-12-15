package pokebat

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"fmt"
)

var battleMutex sync.Mutex

// BattleHandler handles the `/api/battle` endpoint
func BattleHandler(w http.ResponseWriter, r *http.Request) {
	// Lock mutex to ensure thread-safety
	battleMutex.Lock()
	defer battleMutex.Unlock()

	var battle Battle

	// Decode the request body
	err := json.NewDecoder(r.Body).Decode(&battle)
	if err != nil {
		log.Println("Error decoding request:", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate that both players have Pokémon
	if len(battle.PlayerA.Pokemons) == 0 || len(battle.PlayerB.Pokemons) == 0 {
		log.Println("One or both players have no Pokémon")
		http.Error(w, "Both players must have Pokémon", http.StatusBadRequest)
		return
	}

	// Determine attacker and defender based on turn
	var attacker, defender *Player
	if battle.Turn == 0 {
		attacker = &battle.PlayerA
		defender = &battle.PlayerB
	} else {
		attacker = &battle.PlayerB
		defender = &battle.PlayerA
	}

	activeAttacker := &attacker.Pokemons[attacker.Active]
	activeDefender := &defender.Pokemons[defender.Active]

	// Perform attack and calculate damage
	damage := CalculateDamage(activeAttacker, activeDefender)
	log.Printf("Damage dealt by %s to %s: %d", activeAttacker.Name, activeDefender.Name, damage)

	// Check if the defender's Pokémon is defeated
	if activeDefender.HP <= 0 {
		log.Printf("%s's Pokémon %s is defeated!", defender.Name, activeDefender.Name)
		defender.Active++ // Move to the next Pokémon

		if defender.Active >= len(defender.Pokemons) {
			// If the defender has no more Pokémon, the attacker wins
			w.Header().Set("Content-Type", "application/json")
			response := map[string]interface{}{
				"winner": attacker.Name,
				"damage": damage,
				"message": attacker.Name + " wins the battle!",
			}
			json.NewEncoder(w).Encode(response)
			log.Printf("Battle ended. Winner: %s", attacker.Name)
			return
		}

		// Log next Pokémon for the defender
		log.Printf("%s switches to their next Pokémon: %s", defender.Name, defender.Pokemons[defender.Active].Name)
	}

	// Switch turn to the other player
	SwitchTurn(&battle)

	// Prepare response with the updated battle state
	response := map[string]interface{}{
		"battle": map[string]interface{}{
			"player_a": map[string]interface{}{
				"name":     battle.PlayerA.Name,
				"active":   battle.PlayerA.Active,
				"pokemons": battle.PlayerA.Pokemons,
			},
			"player_b": map[string]interface{}{
				"name":     battle.PlayerB.Name,
				"active":   battle.PlayerB.Active,
				"pokemons": battle.PlayerB.Pokemons,
			},
			"turn": battle.Turn,
		},
		"damage": damage,
		"message": activeAttacker.Name + " dealt " + fmt.Sprintf("%d", damage) + " damage to " + activeDefender.Name,
	}

	// Send the response as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
