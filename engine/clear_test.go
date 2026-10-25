package engine

import "testing"

// C9: retreating clears every condition on the Pokémon that left the Active spot.
func TestRetreatClearsConditions(t *testing.T) {
	// Arrange
	g, me := conditionGame()
	me.Inflict(Poison)
	me.Inflict(Burn)
	me.Inflict(Confused)

	// Act
	if err := g.Submit(Retreat{Player: 0, BenchIndex: BenchBehind}); err != nil {
		t.Fatalf("retreat: %v", err)
	}

	// Assert
	if g.S.Players[0].Active == me {
		t.Fatal("retreat did not swap")
	}
	if got := me.Conditions(); len(got) != 0 {
		t.Errorf("benched Pokémon kept %v", got)
	}
}

// C9: evolving clears every condition.
func TestEvolveClearsConditions(t *testing.T) {
	for _, s := range []Status{Asleep, Paralyzed, Confused} {
		// Arrange
		g, me := conditionGame()
		pl := &g.S.Players[0]
		pl.Hand = []Card{legalEvo("E", "Evo", "Me")}
		me.Inflict(Poison)
		me.Inflict(Burn)
		me.Inflict(s)

		// Act
		if err := g.Submit(Evolve{Player: 0, HandCardID: "E", Target: 0}); err != nil {
			t.Fatalf("%v: evolve: %v", s, err)
		}

		// Assert
		if got := pl.Active.Conditions(); len(got) != 0 {
			t.Errorf("%v: evolved Pokémon kept %v", s, got)
		}
		if !hasMove[UseAttack](g.LegalMoves(0)) && len(pl.Active.Card.Attacks) > 0 {
			t.Errorf("%v: evolved Pokémon still can't attack", s)
		}
	}
}
