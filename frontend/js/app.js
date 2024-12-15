document.addEventListener("DOMContentLoaded", () => {
    const API_BASE = "http://localhost:8080/api";

    // DOM Elements
    const menuScreen = document.getElementById("menu-screen");
    const battleScreen = document.getElementById("battle-screen");
    const pokemonWorldScreen = document.getElementById("pokemon-world-screen");
    const yourPokemonScreen = document.getElementById("your-pokemon-screen");
    const playerAImg = document.getElementById("player-a-img");
    const playerBImg = document.getElementById("player-b-img");
    const playerAHPBar = document.getElementById("player-a-hp-bar");
    const playerBHPBar = document.getElementById("player-b-hp-bar");
    const battleLog = document.getElementById("battle-log");
    const pokemonWorldMap = document.getElementById("pokemon-world-map");
    const pokemonList = document.getElementById("pokemon-list");

    // Global Data
    let battleData = null;

    /** --------------------------------------------------------------------
     *  Battle Functionality
     * ------------------------------------------------------------------- */
    const logMessage = (message) => {
        const logEntry = document.createElement("p");
        logEntry.textContent = message;
        battleLog.appendChild(logEntry);
    };

    const updateHPBar = (bar, currentHP, maxHP) => {
        bar.style.width = `${(currentHP / maxHP) * 100}%`;
    };

    const loadBattleData = async () => {
        try {
            const response = await fetch(`${API_BASE}/battle`, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({
                    player_a: {
                        name: "Ash",
                        pokemons: [
                            {
                                name: "Pikachu",
                                image: "https://img.pokemondb.net/sprites/home/normal/pikachu.png",
                                type: ["Electric"],
                                hp: 100,
                                attack: 50,
                                defense: 30,
                                special_attack: 40,
                                special_defense: 20,
                                speed: 90,
                            },
                        ],
                    },
                    player_b: {
                        name: "Opponent",
                        pokemons: [
                            {
                                name: "Charizard",
                                image: "https://img.pokemondb.net/sprites/home/normal/charizard.png",
                                type: ["Fire"],
                                hp: 120,
                                attack: 70,
                                defense: 50,
                                special_attack: 60,
                                special_defense: 30,
                                speed: 100,
                            },
                        ],
                    },
                }),
            });

            if (!response.ok) throw new Error("Failed to load battle data");
            battleData = await response.json();
            updateBattleUI();
        } catch (error) {
            console.error("Error loading battle data:", error);
        }
    };

    const updateBattleUI = () => {
        if (!battleData) return;

        const playerA = battleData.player_a.pokemons[battleData.player_a.active];
        const playerB = battleData.player_b.pokemons[battleData.player_b.active];

        playerAImg.src = playerA.image;
        playerBImg.src = playerB.image;

        updateHPBar(playerAHPBar, playerA.hp, 100);
        updateHPBar(playerBHPBar, playerB.hp, 100);
    };

    const attack = async () => {
        try {
            const response = await fetch(`${API_BASE}/battle/attack`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
            });

            if (!response.ok) throw new Error("Attack failed");

            const result = await response.json();
            battleData = result.battle;
            updateBattleUI();
            logMessage(result.log);

            if (result.winner) {
                alert(`${result.winner} wins the battle!`);
                resetBattle();
            }
        } catch (error) {
            console.error("Error attacking:", error);
        }
    };

    const switchPokemon = async () => {
        try {
            const response = await fetch(`${API_BASE}/switchPokemon`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
            });

            if (!response.ok) throw new Error("Switch Pokémon failed");

            const result = await response.json();
            battleData = result.battle;
            updateBattleUI();
            logMessage("Switched Pokémon!");
        } catch (error) {
            console.error("Error switching Pokémon:", error);
        }
    };

    /** --------------------------------------------------------------------
     *  Pokémon World Functionality (Catch Pokémon)
     * ------------------------------------------------------------------- */
    let gridSize = 10; // Default grid size (10x10)
    let playerPosition = { x: 4, y: 4 }; // Initial player position
    let pokemonsOnGrid = []; // Store Pokémon positions and details

    const renderPokemonWorld = () => {
        pokemonWorldMap.innerHTML = ""; // Clear the grid

        // Create grid
        for (let row = 0; row < gridSize; row++) {
            for (let col = 0; col < gridSize; col++) {
                const cell = document.createElement("div");
                cell.classList.add("grid-cell");

                // Check if this cell is the player's position
                if (playerPosition.x === col && playerPosition.y === row) {
                    const playerMarker = document.createElement("div");
                    playerMarker.classList.add("player");
                    cell.appendChild(playerMarker);
                }

                // Check if this cell contains a Pokémon
                pokemonsOnGrid.forEach((pokemon) => {
                    if (pokemon.x === col && pokemon.y === row) {
                        const pokemonMarker = document.createElement("div");
                        pokemonMarker.classList.add("pokemon");
                        pokemonMarker.style.backgroundImage = `url(${pokemon.image})`;
                        cell.appendChild(pokemonMarker);
                    }
                });

                pokemonWorldMap.appendChild(cell);
            }
        }
    };

    const loadPokemonWorld = async () => {
        try {
            const response = await fetch(`${API_BASE}/pokedex`);
            if (!response.ok) throw new Error("Failed to load Pokémon data");

            const pokedex = await response.json();
            pokemonsOnGrid = []; // Reset Pokémon positions

            // Spawn random Pokémon on the grid
            for (let i = 0; i < 10; i++) { // Spawn 10 random Pokémon
                const randomPokemon = pokedex[Math.floor(Math.random() * pokedex.length)];
                pokemonsOnGrid.push({
                    name: randomPokemon.name,
                    image: randomPokemon.image,
                    x: Math.floor(Math.random() * gridSize),
                    y: Math.floor(Math.random() * gridSize),
                });
            }

            renderPokemonWorld();
        } catch (error) {
            console.error("Error loading Pokémon World:", error);
        }
    };

    const movePlayer = (direction) => {
        let previousPosition = { ...playerPosition };
    
        switch (direction) {
            case "up":
                if (playerPosition.y > 0) playerPosition.y -= 1;
                break;
            case "down":
                if (playerPosition.y < gridSize - 1) playerPosition.y += 1;
                break;
            case "left":
                if (playerPosition.x > 0) playerPosition.x -= 1;
                break;
            case "right":
                if (playerPosition.x < gridSize - 1) playerPosition.x += 1;
                break;
        }
    
        // Check if player collided with a Pokémon
        const collidedPokemonIndex = pokemonsOnGrid.findIndex(
            (pokemon) => pokemon.x === playerPosition.x && pokemon.y === playerPosition.y
        );
    
        if (collidedPokemonIndex !== -1) {
            catchPokemon(collidedPokemonIndex); // Trigger Pokémon catching logic
        } else {
            renderPokemonWorld(); // Just render the grid if no collision occurred
        }
    };

    const catchPokemon = (index) => {
        const pokemon = pokemonsOnGrid[index];
    
        fetch(`${API_BASE}/catchPokemon`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({ name: pokemon.name }),
        })
            .then((response) => {
                if (!response.ok) throw new Error("Failed to catch Pokémon");
                return response.json();
            })
            .then((result) => {
                alert(result.message); // Notify the user of the result
                pokemonsOnGrid.splice(index, 1); // Remove the caught Pokémon from the grid
                renderPokemonWorld(); // Re-render the grid
            })
            .catch((error) => {
                console.error("Error catching Pokémon:", error);
                alert("Failed to catch Pokémon. Please try again.");
            });
    };

    // Add keyboard controls for player movement
    document.addEventListener("keydown", (event) => {
        switch (event.key) {
            case "ArrowUp":
                movePlayer("up");
                break;
            case "ArrowDown":
                movePlayer("down");
                break;
            case "ArrowLeft":
                movePlayer("left");
                break;
            case "ArrowRight":
                movePlayer("right");
                break;
        }
    });

    // Add button controls for player movement
    document.getElementById("move-up").addEventListener("click", () => movePlayer("up"));
    document.getElementById("move-down").addEventListener("click", () => movePlayer("down"));
    document.getElementById("move-left").addEventListener("click", () => movePlayer("left"));
    document.getElementById("move-right").addEventListener("click", () => movePlayer("right"));

    // Load Pokémon World on screen activation
    loadPokemonWorld();

// Back Button for Pokémon World
document.getElementById("back-to-menu-from-world-btn").addEventListener("click", () => {
    pokemonWorldScreen.classList.add("hidden");
    menuScreen.classList.remove("hidden");
});


    /** --------------------------------------------------------------------
     *  Your Pokémon Functionality
     * ------------------------------------------------------------------- */
   const loadPlayerPokemons = async () => {
    try {
        // Fetch player data from the backend
        const response = await fetch(`${API_BASE}/players`);
        if (!response.ok) throw new Error("Failed to fetch player Pokémon");

        const data = await response.json();
        renderPlayerPokemons(data.player_a.pokemons);
    } catch (error) {
        console.error("Error fetching player Pokémon:", error);
        alert("Failed to load your Pokémon. Please try again.");
    }
};

const renderPlayerPokemons = (pokemons) => {
    // Clear existing Pokémon list
    pokemonList.innerHTML = "";

    // Check if the player has no Pokémon
    if (pokemons.length === 0) {
        const emptyMessage = document.createElement("p");
        emptyMessage.textContent = "You don't have any Pokémon yet!";
        pokemonList.appendChild(emptyMessage);
        return;
    }

    // Render each Pokémon
    pokemons.forEach((pokemon) => {
        const pokemonCard = document.createElement("div");
        pokemonCard.classList.add("pokemon-card");

        // Pokémon Image
        const pokemonImage = document.createElement("img");
        pokemonImage.src = pokemon.image || "https://via.placeholder.com/100";
        pokemonImage.alt = pokemon.name;
        pokemonImage.classList.add("pokemon-card-image");

        // Pokémon Name
        const pokemonName = document.createElement("h3");
        pokemonName.textContent = pokemon.name;

        // Pokémon Details
        const pokemonDetails = document.createElement("p");
        pokemonDetails.textContent = `Level: ${pokemon.level || 1}, Type: ${pokemon.type.join(", ")}`;

        // Append everything to the card
        pokemonCard.appendChild(pokemonImage);
        pokemonCard.appendChild(pokemonName);
        pokemonCard.appendChild(pokemonDetails);

        // Append the card to the Pokémon list
        pokemonList.appendChild(pokemonCard);
    });
};

// Event Listener for "Your Pokémon" Button
document.getElementById("your-pokemon-btn").addEventListener("click", () => {
    menuScreen.classList.add("hidden");
    yourPokemonScreen.classList.remove("hidden");
    loadPlayerPokemons(); // Load player Pokémon when this screen is shown
});

// Back Button for "Your Pokémon" Screen
document
    .getElementById("back-to-menu-from-your-pokemon-btn")
    .addEventListener("click", () => {
        yourPokemonScreen.classList.add("hidden");
        menuScreen.classList.remove("hidden");
    });

    /** --------------------------------------------------------------------
     *  Event Listeners
     * ------------------------------------------------------------------- */
    // Menu Screen
    document.getElementById("start-game-btn").addEventListener("click", () => {
        menuScreen.classList.add("hidden");
        battleScreen.classList.remove("hidden");
        loadBattleData();
    });

    document.getElementById("catch-pokemon-btn").addEventListener("click", () => {
        menuScreen.classList.add("hidden");
        pokemonWorldScreen.classList.remove("hidden");
        loadPokemonWorld();
    });

    document.getElementById("your-pokemon-btn").addEventListener("click", () => {
        menuScreen.classList.add("hidden");
        yourPokemonScreen.classList.remove("hidden");
        loadPlayerPokemons();
    });

    // Back Buttons
    document.getElementById("back-btn").addEventListener("click", () => {
        battleScreen.classList.add("hidden");
        menuScreen.classList.remove("hidden");
    });

    document
        .getElementById("back-to-menu-from-world-btn")
        .addEventListener("click", () => {
            pokemonWorldScreen.classList.add("hidden");
            menuScreen.classList.remove("hidden");
        });
});
