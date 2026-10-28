package main

import (
	"testing"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// Alolan Muk's Power of Alchemy: "Basic Pokémon in play (both yours and your
// opponent's) have no Abilities." Eevee ex is a Basic, so while Muk is in play
// Veevee ’volve does nothing and Eevee ex can't take an Eeveelution (owner
// 2026-10-28, decision 5: follow the card text).
func TestMukTurnsOffVeeveeVolve(t *testing.T) {
	for _, tc := range []struct {
		name    string
		mukSide int // -1: no Muk; 0: the mover's Bench; 1: the opponent's Bench
	}{{"no Muk", -1}, {"Muk on your Bench", 0}, {"Muk on the opponent's Bench", 1}} {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			useRealCards(t)
			g, p := eeveeExGame(t, sylveonOne)
			if tc.mukSide >= 0 {
				owner := p
				if tc.mukSide == 1 {
					owner = 1 - p
				}
				put(g, owner, 1, realCard(t, "B2-097")) // Alolan Muk
			}

			// Act
			err := g.Submit(engine.Evolve{Player: p, HandCardID: sylveonOne, Target: 0})
			offered := false
			for _, mv := range g.LegalMoves(p) {
				if ev, ok := mv.(engine.Evolve); ok && ev.HandCardID == sylveonOne {
					offered = true
				}
			}

			// Assert
			if tc.mukSide < 0 {
				if err != nil {
					t.Fatalf("without Muk, Eevee ex can't evolve into Sylveon: %v", err)
				}
				if got := g.S.Players[p].Active.Card.Name; got != "Sylveon" {
					t.Errorf("Active = %s, want Sylveon", got)
				}
				return
			}
			if err == nil {
				t.Error("Eevee ex evolved into Sylveon with Alolan Muk in play")
			}
			if offered {
				t.Error("LegalMoves offers Eevee ex -> Sylveon with Alolan Muk in play")
			}
		})
	}
}
