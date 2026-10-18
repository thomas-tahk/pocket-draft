package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// viewTestDeck builds 20 Basics whose IDs all start with prefix, so a JSON
// search for the prefix finds any of that player's cards.
func viewTestDeck(prefix string) []engine.Card {
	deck := make([]engine.Card, 20)
	for i := range deck {
		deck[i] = engine.Card{ID: fmt.Sprintf("%s-%02d", prefix, i), Name: prefix, Stage: engine.Basic, HP: 60, Type: engine.Fire}
	}
	return deck
}

func TestViewForPlayer0NeverContainsPlayer1HandCardIDs(t *testing.T) {
	// Arrange
	g := engine.NewGame(7, viewTestDeck("mine"), viewTestDeck("theirs"))
	theirHand := len(g.S.Players[1].Hand)

	// Act
	v := toGameView(g, 0)
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	// Assert
	if strings.Contains(string(raw), "theirs-") {
		t.Fatalf("player 0's view leaks player 1's cards: %s", raw)
	}
	if got := v.Players[1].HandCount; got != theirHand {
		t.Fatalf("opponent hand count = %d, want %d", got, theirHand)
	}
	if got := len(v.Players[0].Hand); got != len(g.S.Players[0].Hand) {
		t.Fatalf("own hand has %d cards, want %d", got, len(g.S.Players[0].Hand))
	}
}
