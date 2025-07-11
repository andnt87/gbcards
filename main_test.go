package main

import (
	"fmt"
	"math/rand"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func dummyTexture() rl.Texture2D {
	return rl.Texture2D{}
}

// helper to initialize dummy player textures.
// It creates a map keyed by card names "cardSpades2" to "cardSpades14".
func dummyPlayerTextures(dummyTex rl.Texture2D) map[string]rl.Texture2D {
	textures := make(map[string]rl.Texture2D)
	for i := 2; i < 15; i++ {
		key := fmt.Sprintf("cardSpades%d", i)
		textures[key] = dummyTex
	}
	return textures
}

func TestCreatePlayerDeckOneTableCard(t *testing.T) {
	// Set a deterministic random seed
	rand.Seed(42)

	// Create a dummy texture and initialize the game with one table card (value: 12)
	dummyTex := dummyTexture()
	g := &Game{
		// TableTextures is not used in CreatePlayerDeck so it can be empty
		TableTextures: make(map[string]rl.Texture2D),
		TableDeck: []Card{
			{Name: "cardHearts12", Value: 12, Texture: dummyTex},
		},
		PlayerDeck:     make([]Card, 0),
		PlayerTextures: dummyPlayerTextures(dummyTex),
	}

	// Call the function to generate the player deck
	g.Level = make([]Level, 1)
	g.LevelSet = 0 // Set the level set to 0 for testing
	g.Level[g.LevelSet].NumTableCards = 3
	g.CreatePlayerDeck()

	// Sum the generated player cards
	sum := 0
	for _, card := range g.PlayerDeck {
		sum += card.Value
	}

	// Expect the total sum equals the table card value (12)
	if sum < 12 {
		t.Errorf("Expected sum of player deck to be at least 12, got %d", sum)
	}
}

func TestCreatePlayerDeckMultipleTableCards(t *testing.T) {
	// Set a deterministic random seed
	rand.Seed(100)

	dummyTex := dummyTexture()
	// Create two table cards with known values (12 and 14)
	g := &Game{
		TableTextures: make(map[string]rl.Texture2D),
		TableDeck: []Card{
			{Name: "cardHearts12", Value: 12, Texture: dummyTex},
			{Name: "cardHearts14", Value: 14, Texture: dummyTex},
		},
		PlayerDeck:     make([]Card, 0),
		PlayerTextures: dummyPlayerTextures(dummyTex),
	}

	g.Level = make([]Level, 1)
	g.LevelSet = 0 // Set the level set to 0 for testing
	g.Level[g.LevelSet].NumTableCards = 3
	g.CreatePlayerDeck()

	// Sum player cards
	sum := 0
	for _, card := range g.PlayerDeck {
		sum += card.Value
	}

	expectedSum := 12 + 14
	if sum < expectedSum {
		t.Errorf("Expected sum of player deck to be at least %d, got %d", expectedSum, sum)
	}
}
