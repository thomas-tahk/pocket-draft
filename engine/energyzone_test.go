package engine

import "testing"

func endTurn(t *testing.T, g *Game) {
	t.Helper()
	if err := g.Submit(EndTurn{Player: g.S.Active}); err != nil {
		t.Fatalf("end turn %d: %v", g.S.Turn, err)
	}
}

// Spec T6: energy you don't attach stays in the zone through the opponent's
// turn, is discarded at the start of your next turn (counted as discarded
// energy), and next turn's energy moves into the large spot.
func TestUnattachedEnergyVisibleThenDiscarded(t *testing.T) {
	d0, d1 := twoDecks()
	g := NewGame(3, d0, d1)
	doSetup(t, g)
	endTurn(t, g)   // turn 1: going first, no energy
	p := g.S.Active // turn 2
	pl := &g.S.Players[p]
	unused, next := pl.EnergyZone, pl.EnergyNext
	if unused == NoEnergy || next == NoEnergy {
		t.Fatalf("turn 2 zone now=%v next=%v, want both rolled", unused, next)
	}

	endTurn(t, g) // turn 2 ends without attaching
	if got := g.S.Players[p].EnergyZone; got != unused {
		t.Fatalf("opponent's turn: zone shows %v, want unused %v", got, unused)
	}
	if n := len(g.S.Players[p].DiscardedEnergy); n != 0 {
		t.Fatalf("opponent's turn: %d energy discarded, want 0", n)
	}

	endTurn(t, g) // turn 4 starts
	pl = &g.S.Players[p]
	if len(pl.DiscardedEnergy) != 1 || pl.DiscardedEnergy[0] != unused {
		t.Fatalf("discarded energy %v, want [%v]", pl.DiscardedEnergy, unused)
	}
	if pl.EnergyZone != next {
		t.Errorf("zone now %v, want last turn's next %v", pl.EnergyZone, next)
	}
	if pl.EnergyNext == NoEnergy {
		t.Error("a new next energy was not rolled")
	}
}

// Spec T5: the player going first gets no energy on turn 1 but the zone still
// shows their next-turn energy; attached energy is never discarded.
func TestFirstTurnShowsNextAndAttachedIsNotDiscarded(t *testing.T) {
	d0, d1 := twoDecks()
	g := NewGame(5, d0, d1)
	doSetup(t, g)
	first := &g.S.Players[g.S.Active]
	if first.EnergyZone != NoEnergy || first.EnergyNext == NoEnergy {
		t.Fatalf("turn 1 zone now=%v next=%v, want none + rolled", first.EnergyZone, first.EnergyNext)
	}
	if g.S.Players[1-g.S.Active].EnergyNext == NoEnergy {
		t.Fatal("second player's next energy is not shown on turn 1")
	}
	endTurn(t, g)
	p := g.S.Active
	if err := g.Submit(AttachEnergy{Player: p, Target: 0}); err != nil {
		t.Fatalf("attach: %v", err)
	}
	endTurn(t, g)
	endTurn(t, g)
	if n := len(g.S.Players[p].DiscardedEnergy); n != 0 {
		t.Errorf("attached energy was discarded: %v", g.S.Players[p].DiscardedEnergy)
	}
}
