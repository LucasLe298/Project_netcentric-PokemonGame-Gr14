package pokebat

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"PROJECT_NETCENTRIC/pokedex"
)

var BattleData Battle

// Define Player
type Player struct {
	Name     string           `json:"name"`
	Pokemons []pokedex.Pokemon `json:"pokemons"`
	Active   int              `json:"active"`
}

// Define Battle
type Battle struct {
	PlayerA Player `json:"player_a"`
	PlayerB Player `json:"player_b"`
	Turn    int    `json:"turn"`
}

// Load And Create A New Battle
func LoadOrCreateBattleData(filePath string) (Battle, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		defaultData := Battle{
			PlayerA: Player{
				Name: "PlayerA",
				Pokemons: []pokedex.Pokemon{
					{Name: "Bulbasaur", Type: []string{"Grass"}, HP: 100},
					{Name: "Charmander", Type: []string{"Fire"}, HP: 90},
				},
				Active: 0,
			},
			PlayerB: Player{
				Name: "PlayerB",
				Pokemons: []pokedex.Pokemon{
					{Name: "Pidgey", Type: []string{"Flying"}, HP: 80},
					{Name: "Rattata", Type: []string{"Normal"}, HP: 75},
				},
				Active: 0,
			},
			Turn: 0,
		}

		data, _ := json.MarshalIndent(defaultData, "", "  ")
		ioutil.WriteFile(filePath, data, 0644)
		return defaultData, nil
	}

	// Load battle_data file
	file, _ := os.ReadFile(filePath)
	var battle Battle
	json.Unmarshal(file, &battle)
	return battle, nil
}

// Switch Pokemon in a battle
func HandlePlayerSwitch(player *Player, newPokemonName string) error {
	for i, pokemon := range player.Pokemons {
		if pokemon.Name == newPokemonName {
			player.Active = i
			return nil
		}
	}
	return fmt.Errorf("Pokémon not found in player's list: %s", newPokemonName)
}

// Calculate multiplier between 2 Pokémons
func CalculateElementalMultiplier(attackerType string, defenderTypes []string) float64 {
	typeAdvantages := map[string]map[string]float64{
		"Fire": {
			"Grass": 2.0,
			"Water": 0.5,
			"Fire":  1.0,
		},
		"Water": {
			"Fire":  2.0,
			"Grass": 0.5,
			"Water": 1.0,
		},
		"Grass": {
			"Water": 2.0,
			"Fire":  0.5,
			"Grass": 1.0,
		},
	}

	multiplier := 1.0
	for _, defenderType := range defenderTypes {
		if advantage, ok := typeAdvantages[attackerType][defenderType]; ok {
			multiplier *= advantage
		}
	}
	return multiplier
}

// Calculate damage
func CalculateDamage(attacker, defender *pokedex.Pokemon) int {
	if rand.Intn(2) == 0 { // Normal Attack
		damage := attacker.Attack - defender.Defense
		if damage < 0 {
			damage = 0
		}
		defender.HP -= damage
		return damage
	} else { // Special Attack
		elementalMultiplier := CalculateElementalMultiplier(attacker.Type[0], defender.Type)
		damage := int(float64(attacker.SpecialAttack)*elementalMultiplier) - defender.SpecialDefense
		if damage < 0 {
			damage = 0
		}
		defender.HP -= damage
		return damage
	}
}

// SwitchTurn in battle
func SwitchTurn(battle *Battle) {
	if battle.Turn == 0 {
		battle.Turn = 1
	} else {
		battle.Turn = 0
	}
}

// ResolveBattleTurn to conduct attack turn
func ResolveBattleTurn(battle *Battle) (map[string]interface{}, error) {
	var attacker, defender *Player
	if battle.Turn == 0 {
		attacker = &battle.PlayerA
		defender = &battle.PlayerB
	} else {
		attacker = &battle.PlayerB
		defender = &battle.PlayerA
	}

	// Take Pokémon is active
	activeAttacker := &attacker.Pokemons[attacker.Active]
	activeDefender := &defender.Pokemons[defender.Active]

	// Attack
	damage := CalculateDamage(activeAttacker, activeDefender)

	// Check if Defender Pokémon is defeated
	if activeDefender.HP <= 0 {
		log.Printf("%s's Pokémon %s is defeated!", defender.Name, activeDefender.Name)
		defender.Active++
		if defender.Active >= len(defender.Pokemons) {
			// Defender không còn Pokémon, attacker thắng
			return map[string]interface{}{
				"winner": attacker.Name,
				"damage": damage,
			}, nil
		}
	}

	// Chuyển lượt
	SwitchTurn(battle)

	return map[string]interface{}{
		"battle": battle,
		"damage": damage,
	}, nil
}
