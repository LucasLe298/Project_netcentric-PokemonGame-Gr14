package pokedex

import (
	"encoding/json"
	"net/http"
)

func GetPokedexHandler(w http.ResponseWriter, r *http.Request) {
	pokedex, err := LoadPokedex("pokedex.json")
	if err != nil {
		http.Error(w, "Error loading pokedex", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pokedex)
}
