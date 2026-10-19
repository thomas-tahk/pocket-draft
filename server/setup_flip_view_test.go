package main

import (
	"testing"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// Spec S4: the setup screen shows who flipped, how the coin landed and who
// goes first, each from the viewer's seat, and both sides' first energy.
func TestViewShowsOpeningFlipAndFirstEnergyAtSetup(t *testing.T) {
	for seed := uint64(1); seed <= 8; seed++ {
		// Arrange
		g := engine.NewGame(seed, viewTestDeck("mine"), viewTestDeck("theirs"))
		flip := g.S.OpeningFlip

		// Act
		v := toGameView(g, 0)

		// Assert
		if v.OpeningFlip.Flipper != flip.Flipper || v.OpeningFlip.Heads != flip.Heads {
			t.Fatalf("seed %d: view flip %+v, want %+v", seed, v.OpeningFlip, flip)
		}
		if v.FirstPlayer != g.S.FirstPlayer {
			t.Fatalf("seed %d: view first player %d, want %d", seed, v.FirstPlayer, g.S.FirstPlayer)
		}
		if v.Players[0].EnergyNext == "" || v.Players[1].EnergyNext == "" {
			t.Fatalf("seed %d: setup view hides first energy: %q / %q", seed, v.Players[0].EnergyNext, v.Players[1].EnergyNext)
		}
	}
}
