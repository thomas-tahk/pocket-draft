package main

import (
	"testing"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// Spec T4/T6: the view carries both the now and next energy, and the energy
// discarded from the zone, so the screen can show all three.
func TestViewShowsEnergyNowNextAndDiscarded(t *testing.T) {
	// Arrange
	g := engine.NewGame(7, viewTestDeck("mine"), viewTestDeck("theirs"))
	pl := &g.S.Players[0]
	pl.EnergyZone, pl.EnergyNext = engine.Fire, engine.Water
	pl.DiscardedEnergy = []engine.Energy{engine.Grass}

	// Act
	v := toGameView(g, 1).Players[0]

	// Assert
	if v.EnergyZone != "Fire" || v.EnergyNext != "Water" {
		t.Errorf("zone now=%q next=%q, want Fire/Water", v.EnergyZone, v.EnergyNext)
	}
	if len(v.DiscardedEnergy) != 1 || v.DiscardedEnergy[0] != "Grass" {
		t.Errorf("discarded energy %v, want [Grass]", v.DiscardedEnergy)
	}
}
