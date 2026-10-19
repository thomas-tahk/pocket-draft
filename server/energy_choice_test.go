package main

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// waterSynthDeck is 20 card IDs (10 Water Basics x2) from syntheticDraftPool,
// whose attacks each cost one Water.
func waterSynthDeck() []string {
	var ids []string
	for i := 1; i < 50; i += 5 {
		id := fmt.Sprintf("SYN-%03d", i)
		ids = append(ids, id, id)
	}
	return ids
}

// Spec S2 through POST /api/new: the deck's chosen energy types reach the game,
// and with none chosen the default comes from the deck's attack costs.
func TestStartGameUsesChosenEnergyTypes(t *testing.T) {
	syntheticDraftPool(t)
	seed := uint64(3)
	cases := []struct {
		name   string
		energy []string
		want   []engine.Energy
	}{
		{"default from attack costs", nil, []engine.Energy{engine.Water}},
		{"chosen", []string{"Grass", "Metal"}, []engine.Energy{engine.Grass, engine.Metal}},
	}
	for _, c := range cases {
		view, err := startGame(newReq{You: waterSynthDeck(), Seed: &seed, Energy: c.energy})
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if got := game.S.Players[0].EnergyTypes; !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: engine got %v, want %v", c.name, got, c.want)
		}
		if got := view.Players[0].EnergyTypes; len(got) != len(c.want) || got[0] != c.want[0].String() {
			t.Errorf("%s: view got %v, want %v", c.name, got, c.want)
		}
	}
}

func TestStartGameRejectsBadEnergyChoice(t *testing.T) {
	syntheticDraftPool(t)
	for _, energy := range [][]string{{"Dragon"}, {"Fire", "Water", "Grass", "Metal"}, {"Fire", "Fire"}, {"Plasma"}} {
		_, err := startGame(newReq{You: waterSynthDeck(), Energy: energy})
		if err == nil || err.status != http.StatusBadRequest {
			t.Errorf("%v: want 400, got %v", energy, err)
		}
	}
}
