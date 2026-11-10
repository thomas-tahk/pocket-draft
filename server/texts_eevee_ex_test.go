package main

import (
	"testing"

	"github.com/thomas-tahk/pocket-draft/engine"
)

const (
	eeveeExID  = "A3b-056" // Eevee ex, Veevee ’volve
	sylveonOne = "A3b-033" // Sylveon, Stage 1 from Eevee
	ivysaurID  = "A1-002"  // Stage 1 from Bulbasaur
)

// eeveeExGame: Eevee ex as the mover's Active, placed on an earlier turn, on a
// turn past the first, with the evolution cards in hand.
func eeveeExGame(t *testing.T, hand ...string) (*engine.Game, int) {
	t.Helper()
	g := engine.NewGame(7, deckOfCard(realCard(t, eeveeExID), 20), deckOfCard(realCard(t, eeveeExID), 20))
	doSetup(t, g)
	p := g.S.Active
	g.S.Turn = 5
	g.S.Players[p].Active.PlacedTurn = 1
	for _, id := range hand {
		g.S.Players[p].Hand = append(g.S.Players[p].Hand, realCard(t, id))
	}
	return g, p
}

// "This Pokémon can evolve into any Pokémon that evolves from Eevee if you play
// it from your hand onto this Pokémon."
func TestVeeveeVolveEvolvesIntoAnEeveelution(t *testing.T) {
	// Arrange
	useRealCards(t)
	g, p := eeveeExGame(t, sylveonOne)

	// Act
	err := g.Submit(engine.Evolve{Player: p, HandCardID: sylveonOne, Target: 0})

	// Assert
	if err != nil {
		t.Fatalf("Eevee ex can't evolve into Sylveon: %v", err)
	}
	if got := g.S.Players[p].Active.Card.Name; got != "Sylveon" {
		t.Errorf("Active = %s, want Sylveon", got)
	}
}

// The parenthetical: not the turn it was played, nor the first turn.
func TestVeeveeVolveStillWaitsATurn(t *testing.T) {
	// Arrange
	useRealCards(t)
	g, p := eeveeExGame(t, sylveonOne)
	g.S.Players[p].Active.PlacedTurn = g.S.Turn

	// Act
	err := g.Submit(engine.Evolve{Player: p, HandCardID: sylveonOne, Target: 0})

	// Assert
	if err == nil {
		t.Error("Eevee ex evolved the turn it was played")
	}
}

// It takes Pokémon that evolve from Eevee only, not any Stage 1.
func TestVeeveeVolveIsOnlyForEeveelutions(t *testing.T) {
	// Arrange
	useRealCards(t)
	g, p := eeveeExGame(t, ivysaurID)

	// Act
	err := g.Submit(engine.Evolve{Player: p, HandCardID: ivysaurID, Target: 0})

	// Assert
	if err == nil {
		t.Error("Eevee ex evolved into Ivysaur")
	}
}

// The evolution is offered as a legal move, so the screen and the bot see it.
func TestVeeveeVolveIsALegalMove(t *testing.T) {
	// Arrange
	useRealCards(t)
	g, p := eeveeExGame(t, sylveonOne)

	// Act
	found := false
	for _, mv := range g.LegalMoves(p) {
		if ev, ok := mv.(engine.Evolve); ok && ev.HandCardID == sylveonOne && ev.Target == 0 {
			found = true
		}
	}

	// Assert
	if !found {
		t.Error("LegalMoves has no Evolve of Eevee ex into Sylveon")
	}
}
