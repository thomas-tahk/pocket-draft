package main

import (
	"testing"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// F17: the view says why each of the viewer's hand cards, and the zone's
// energy, has nowhere to go, so the screen flashes the engine's reason.
func TestViewGivesReasonsForCardsWithNowhereToGo(t *testing.T) {
	// Arrange: your turn 3, energy already attached, Bench full.
	g := engine.NewGame(7, viewTestDeck("mine"), viewTestDeck("theirs"))
	g.S.Phase, g.S.Pending, g.S.Turn, g.S.Active = engine.PhaseMain, nil, 3, 0
	pl := &g.S.Players[0]
	mon := func() *engine.InPlay { return &engine.InPlay{Card: pl.Hand[0], Energy: map[engine.Energy]int{}} }
	pl.Active = mon()
	for s := range pl.Bench {
		pl.Bench[s] = mon()
	}
	g.S.Players[1].Active = mon()
	pl.EnergyZone, pl.EnergyUsed = engine.Fire, true
	card := pl.Hand[0].ID

	// Act
	v := toGameView(g, 0)

	// Assert
	if got := v.Players[0].HandBlocked[card]; got != "your Bench is full" {
		t.Errorf("handBlocked[%s] = %q, want the full-Bench reason", card, got)
	}
	if got := v.Players[0].EnergyBlocked; got != "energy was already attached this turn" {
		t.Errorf("energyBlocked = %q", got)
	}
	if len(v.Players[1].HandBlocked) != 0 || v.Players[1].EnergyBlocked != "" {
		t.Error("the opponent's side carries reasons")
	}

	// A free slot and fresh energy: nothing is blocked.
	pl.Bench[engine.BenchLeft], pl.EnergyUsed = nil, false
	v = toGameView(g, 0)
	if len(v.Players[0].HandBlocked) != 0 || v.Players[0].EnergyBlocked != "" {
		t.Errorf("playable cards carry reasons: %v / %q", v.Players[0].HandBlocked, v.Players[0].EnergyBlocked)
	}
}
