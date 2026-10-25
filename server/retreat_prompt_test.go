package main

import (
	"reflect"
	"testing"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// T9: a retreat paid from mixed energy pauses on a prompt whose answers the
// view lists as postable moves; posting one completes the retreat.
func TestRetreatEnergyPromptIsAnsweredThroughTheView(t *testing.T) {
	// Arrange
	fire, lightning := engine.DemoDecks()
	game = engine.NewGame(1, fire, lightning)
	for game.S.Phase == engine.PhaseSetup {
		p := game.S.Pending.Player
		if resp := applyMove(toGameView(game, p).Players[p].LegalMoves[0]); !resp.OK {
			t.Fatalf("setup refused: %s", resp.Error)
		}
	}
	p := game.S.Active
	pl := &game.S.Players[p]
	pl.Active.Card.RetreatCost = 1
	pl.Active.Energy = map[engine.Energy]int{engine.Fire: 1, engine.Water: 1}
	pl.Bench[engine.BenchLeft] = &engine.InPlay{Card: pl.Active.Card, Energy: map[engine.Energy]int{}}
	if resp := applyMove(moveReq{Type: "Retreat", Player: p, BenchIndex: engine.BenchLeft}); !resp.OK {
		t.Fatalf("retreat refused: %s", resp.Error)
	}

	// Act
	v := toGameView(game, p)

	// Assert
	if v.Pending == nil || v.Pending.Kind != "retreat_energy" {
		t.Fatalf("want a retreat_energy prompt, got %+v", v.Pending)
	}
	moves := v.Players[p].LegalMoves
	want := []moveReq{
		{Type: "RetreatDiscard", Player: p, Energy: []string{"Fire"}},
		{Type: "RetreatDiscard", Player: p, Energy: []string{"Water"}},
		{Type: "Concede", Player: p},
	}
	if !reflect.DeepEqual(moves, want) {
		t.Fatalf("legal moves = %+v, want %+v", moves, want)
	}
	if resp := applyMove(moves[1]); !resp.OK || resp.State.Pending != nil {
		t.Fatalf("answer refused or prompt left: %+v", resp)
	}
	if game.S.Players[p].Bench[engine.BenchLeft].Energy[engine.Fire] != 1 {
		t.Error("the Fire energy should have stayed on the retreated Pokémon")
	}
}
