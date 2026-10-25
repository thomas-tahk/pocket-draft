package engine

import (
	"reflect"
	"testing"
)

// retreatGame starts a main-phase game whose current player has an Active
// with retreat cost 2 holding energy, and one Pokémon behind it on the bench.
func retreatGame(t *testing.T, energy map[Energy]int) (*Game, int) {
	t.Helper()
	d0, d1 := twoDecks()
	g := NewGame(42, d0, d1)
	doSetup(t, g)
	p := g.S.Active
	pl := &g.S.Players[p]
	clearBench(pl)
	pl.Active = &InPlay{Card: legalBasic("R", "Retreater", 2), Energy: energy}
	pl.Bench[BenchBehind] = newInPlay(legalBasic("B", "Bencher", 1), 0)
	return g, p
}

// T9: with mixed energy the player chooses which to discard; the options are
// the distinct sets of energy that pay the cost, attachment order ignored.
func TestRetreatWithMixedEnergyPrompts(t *testing.T) {
	// Arrange
	g, p := retreatGame(t, map[Energy]int{Fire: 1, Water: 2})

	// Act
	if err := g.Submit(Retreat{Player: p, BenchIndex: BenchBehind}); err != nil {
		t.Fatalf("retreat: %v", err)
	}

	// Assert
	pr := g.S.Pending
	if pr == nil || pr.Player != p || pr.Kind != PromptRetreatEnergy {
		t.Fatalf("want a retreat-energy prompt for %s, got %+v", side(p), pr)
	}
	assertMoves(t, pr.Options, []Event{
		RetreatDiscard{Player: p, Energy: []Energy{Fire, Water}},
		RetreatDiscard{Player: p, Energy: []Energy{Water, Water}},
	})
	if g.S.Players[p].Active.Card.ID != "R" {
		t.Error("the Active should not move until the energy is chosen")
	}
}

// T9: the chosen energy is discarded, counted as discarded energy, and the
// swap completes.
func TestRetreatDiscardsTheChosenEnergy(t *testing.T) {
	// Arrange
	g, p := retreatGame(t, map[Energy]int{Fire: 1, Water: 2})
	before := len(g.S.Players[p].DiscardedEnergy)
	if err := g.Submit(Retreat{Player: p, BenchIndex: BenchBehind}); err != nil {
		t.Fatalf("retreat: %v", err)
	}

	// Act
	if err := g.Submit(RetreatDiscard{Player: p, Energy: []Energy{Fire, Water}}); err != nil {
		t.Fatalf("discard: %v", err)
	}

	// Assert
	pl := g.S.Players[p]
	if g.S.Pending != nil || pl.Active.Card.ID != "B" || pl.Bench[BenchBehind].Card.ID != "R" {
		t.Fatalf("retreat should have swapped Active and slot; pending %+v", g.S.Pending)
	}
	if left := pl.Bench[BenchBehind].Energy; left[Fire] != 0 || left[Water] != 1 {
		t.Errorf("energy left = %v, want one Water", left)
	}
	if got := pl.DiscardedEnergy[before:]; !reflect.DeepEqual(got, []Energy{Fire, Water}) {
		t.Errorf("discarded energy = %v, want [Fire Water]", got)
	}
}

// T9: with only one way to pay (mono energy) the engine picks for the player.
func TestRetreatWithMonoEnergyResolvesItself(t *testing.T) {
	// Arrange
	g, p := retreatGame(t, map[Energy]int{Water: 3})
	before := len(g.S.Players[p].DiscardedEnergy)

	// Act
	if err := g.Submit(Retreat{Player: p, BenchIndex: BenchBehind}); err != nil {
		t.Fatalf("retreat: %v", err)
	}

	// Assert
	pl := g.S.Players[p]
	if g.S.Pending != nil {
		t.Fatalf("a one-option retreat prompted: %+v", g.S.Pending)
	}
	if pl.Active.Card.ID != "B" || pl.Bench[BenchBehind].Energy[Water] != 1 {
		t.Error("retreat should have swapped, leaving one Water on the retreated Pokémon")
	}
	if got := pl.DiscardedEnergy[before:]; !reflect.DeepEqual(got, []Energy{Water, Water}) {
		t.Errorf("discarded energy = %v, want [Water Water]", got)
	}
	if len(g.Events) == 0 {
		t.Fatal("no events logged")
	}
	if _, ok := g.Events[len(g.Events)-1].(Retreat); !ok {
		t.Error("an automatic energy choice must not be logged as an event")
	}
}
