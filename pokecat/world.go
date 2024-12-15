package pokecat

import (
	"math/rand"
	"sync"
	"time"

	"PROJECT_NETCENTRIC/pokedex"
)

// Global variable to hold the world state
var World Pokeworld

type Pokeworld struct {
	mu       sync.Mutex
	Players  map[string]*Player       // Players in the world
	Pokemons []PokemonInWorld         // Pokémon currently in the world
}

// Configurations for the Pokeworld
const (
	GridSize      = 1000
	SpawnInterval = 1 * time.Minute
	DespawnTime   = 5 * time.Minute
	MaxPokemons   = 200
)

// Player structure representing a player in the world
type Player struct {
	ID       string            `json:"id"`
	X        int               `json:"x"`
	Y        int               `json:"y"`
	Pokemons []pokedex.Pokemon `json:"pokemons"`
}

// PokemonInWorld represents a spawned Pokémon in the world
type PokemonInWorld struct {
	Pokemon pokedex.Pokemon
	X       int
	Y       int
	Spawned time.Time
}

// AddPlayer adds a new player to the Pokeworld
func (w *Pokeworld) AddPlayer(id string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.Players[id] = &Player{
		ID:       id,
		X:        rand.Intn(GridSize),
		Y:        rand.Intn(GridSize),
		Pokemons: []pokedex.Pokemon{},
	}
}

// MovePlayer moves a player in a specified direction
func (w *Pokeworld) MovePlayer(id, direction string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	player, exists := w.Players[id]
	if !exists {
		return
	}

	switch direction {
	case "up":
		if player.Y > 0 {
			player.Y--
		}
	case "down":
		if player.Y < GridSize-1 {
			player.Y++
		}
	case "left":
		if player.X > 0 {
			player.X--
		}
	case "right":
		if player.X < GridSize-1 {
			player.X++
		}
	}

	// Capture Pokémon if player steps on it
	for i, p := range w.Pokemons {
		if p.X == player.X && p.Y == player.Y {
			if len(player.Pokemons) < MaxPokemons {
				player.Pokemons = append(player.Pokemons, p.Pokemon)
			}
			// Remove the captured Pokémon from the world
			w.Pokemons = append(w.Pokemons[:i], w.Pokemons[i+1:]...)
			break
		}
	}
}

// SpawnPokemons generates new Pokémon in the world
func (w *Pokeworld) SpawnPokemons() {
	w.mu.Lock()
	defer w.mu.Unlock()

	for i := 0; i < 50; i++ {
		pokemon := pokedex.Pokemon{
			Name:           "RandomPokemon",
			Type:           []string{"Normal"}, // Fixed: Use slice for Type
			Level:          rand.Intn(50) + 1,
			HP:             rand.Intn(100) + 50,
			Attack:         rand.Intn(50) + 10,
			Defense:        rand.Intn(50) + 10,
			SpecialAttack:  rand.Intn(50) + 10,
			SpecialDefense: rand.Intn(50) + 10,
			Speed:          rand.Intn(50) + 10,
		}

		w.Pokemons = append(w.Pokemons, PokemonInWorld{
			Pokemon: pokemon,
			X:       rand.Intn(GridSize),
			Y:       rand.Intn(GridSize),
			Spawned: time.Now(),
		})
	}
}

// CleanupPokemons removes Pokémon that are too old
func (w *Pokeworld) CleanupPokemons() {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := time.Now()
	filtered := []PokemonInWorld{}
	for _, p := range w.Pokemons {
		if now.Sub(p.Spawned) < DespawnTime {
			filtered = append(filtered, p)
		}
	}
	w.Pokemons = filtered
}

// StartSpawner manages spawning and cleanup of Pokémon
func (w *Pokeworld) StartSpawner() {
	go func() {
		for {
			time.Sleep(SpawnInterval)
			w.SpawnPokemons()
			w.CleanupPokemons()
		}
	}()
}
