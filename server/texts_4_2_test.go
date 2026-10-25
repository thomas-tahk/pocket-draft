package main

import (
	"testing"

	"github.com/thomas-tahk/pocket-draft/engine"
)

const (
	eevee     = "B1-184"  // Boosted Evolution
	vaporeon  = "B3a-008" // evolves from Eevee
	alolanMuk = "B2-097"  // Power of Alchemy
)

// eeveeGame is a game just past setup with an Eevee as each Active and a
// Vaporeon in the hand of the player on turn 1.
func eeveeGame(t *testing.T) (*engine.Game, int) {
	t.Helper()
	g := engine.NewGame(7, deckOfCard(realCard(t, eevee), 20), deckOfCard(realCard(t, eevee), 20))
	doSetup(t, g)
	p := g.S.Active
	g.S.Players[p].Hand = append(g.S.Players[p].Hand, realCard(t, vaporeon))
	return g, p
}

// T8: "As long as this Pokémon is in the Active Spot, it can evolve during
// your first turn or the turn you play it."
func TestBoostedEvolutionEvolvesOnTheFirstTurn(t *testing.T) {
	// Arrange
	useRealCards(t)
	g, p := eeveeGame(t)

	// Act
	err := g.Submit(engine.Evolve{Player: p, HandCardID: vaporeon, Target: 0})

	// Assert
	if err != nil {
		t.Fatalf("Eevee in the Active Spot can't evolve on turn 1: %v", err)
	}
	if g.S.Players[p].Active.Card.Name != "Vaporeon" {
		t.Errorf("Active = %s, want Vaporeon", g.S.Players[p].Active.Card.Name)
	}
}

// Only "in the Active Spot": an Eevee on the Bench still waits a turn.
func TestBoostedEvolutionNeedsTheActiveSpot(t *testing.T) {
	// Arrange
	useRealCards(t)
	g, p := eeveeGame(t)
	pl := &g.S.Players[p]
	pl.Bench[engine.BenchLeft] = &engine.InPlay{Card: realCard(t, eevee), Energy: map[engine.Energy]int{}, PlacedTurn: g.S.Turn}

	// Act
	err := g.Submit(engine.Evolve{Player: p, HandCardID: vaporeon, Target: 1 + engine.BenchLeft})

	// Assert
	if err == nil {
		t.Error("a Benched Eevee evolved the turn it was played")
	}
}

// "Basic Pokémon in play (both yours and your opponent's) have no Abilities."
// turns Boosted Evolution off, from either side.
func TestPowerOfAlchemyTurnsOffBoostedEvolution(t *testing.T) {
	for _, side := range []int{0, 1} {
		// Arrange
		useRealCards(t)
		g, p := eeveeGame(t)
		owner := (p + side) % 2
		g.S.Players[owner].Bench[engine.BenchBehind] = &engine.InPlay{Card: realCard(t, alolanMuk), Energy: map[engine.Energy]int{}}

		// Act
		err := g.Submit(engine.Evolve{Player: p, HandCardID: vaporeon, Target: 0})

		// Assert
		if err == nil {
			t.Errorf("Muk on side %d: Eevee evolved on turn 1 with no Abilities", owner)
		}
	}
}

// Every printing of an Ability's wording gets that Ability, named as printed.
func TestAbilityTextMapsEveryPrinting(t *testing.T) {
	useRealCards(t)
	for _, id := range []string{"B1-184", "P-B-011", "P-B-054", "B2-097", "B2-173"} {
		c := realCard(t, id)
		if c.Ability == nil || c.Ability.Name != rawByID[id].Ability.Name {
			t.Errorf("%s (%s) Ability = %+v, want %q", id, c.Name, c.Ability, rawByID[id].Ability.Name)
		}
	}
}
