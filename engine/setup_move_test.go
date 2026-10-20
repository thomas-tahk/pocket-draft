package engine

import "testing"

// Owner 2026-10-19 (8): before Ready a placed card can go back to the hand or
// move between spots; none of it is possible after Ready.
func TestSetupPlacedCardCanBeReturnedAndMoved(t *testing.T) {
	// Arrange
	d0, d1 := twoDecks()
	g := NewGame(3, d0, d1)
	a, bench := pickSetup(g.S.Players[0].Hand)
	mustSetup(t, g, SetupActive{Player: 0, CardID: a})
	mustSetup(t, g, SetupBench{Player: 0, CardID: bench[0], Slot: BenchLeft})
	handBefore := len(g.S.Players[0].Hand)
	pl := &g.S.Players[0]

	// Act + Assert: a Bench card goes back to the hand
	mustSetup(t, g, SetupReturn{Player: 0, Pos: 1})
	if pl.Bench[BenchLeft] != nil || len(pl.Hand) != handBefore+1 {
		t.Fatalf("return: bench %v, hand %d want %d", pl.Bench[BenchLeft], len(pl.Hand), handBefore+1)
	}

	// Act + Assert: the Active moves to the Bench, leaving no Active
	mustSetup(t, g, SetupMove{Player: 0, From: 0, To: 3})
	if pl.Active != nil || pl.Bench[BenchRight] == nil || pl.Bench[BenchRight].Card.ID != a {
		t.Fatalf("move: active %v bench-right %v", pl.Active, pl.Bench[BenchRight])
	}
	if err := g.Submit(SetupReady{Player: 0}); err == nil {
		t.Fatal("Ready without an Active was accepted")
	}

	// Act + Assert: Bench to Active, and an occupied target swaps
	mustSetup(t, g, SetupMove{Player: 0, From: 3, To: 0})
	mustSetup(t, g, SetupBench{Player: 0, CardID: bench[0], Slot: BenchLeft})
	mustSetup(t, g, SetupMove{Player: 0, From: 0, To: 1})
	if pl.Active.Card.ID != bench[0] || pl.Bench[BenchLeft].Card.ID != a {
		t.Fatalf("swap: active %s bench-left %s", pl.Active.Card.ID, pl.Bench[BenchLeft].Card.ID)
	}

	// Assert: refused after Ready
	mustSetup(t, g, SetupReady{Player: 0})
	for _, ev := range []Event{SetupReturn{Player: 0, Pos: 0}, SetupMove{Player: 0, From: 0, To: 1}} {
		if err := g.Submit(ev); err == nil {
			t.Fatalf("%T accepted after Ready", ev)
		}
	}
}

func TestSetupReturnAndMoveNeedAPlacedCard(t *testing.T) {
	d0, d1 := twoDecks()
	g := NewGame(3, d0, d1)
	for _, ev := range []Event{SetupReturn{Player: 0, Pos: 2}, SetupMove{Player: 0, From: 1, To: 2}, SetupMove{Player: 0, From: 0, To: 9}} {
		if err := g.Submit(ev); err == nil {
			t.Fatalf("%T on an empty spot was accepted", ev)
		}
	}
}

func TestSetupLegalMovesListReturnAndMove(t *testing.T) {
	d0, d1 := twoDecks()
	g := NewGame(3, d0, d1)
	a, _ := pickSetup(g.S.Players[0].Hand)
	mustSetup(t, g, SetupActive{Player: 0, CardID: a})
	var returns, moves int
	for _, ev := range g.LegalMoves(0) {
		switch ev.(type) {
		case SetupReturn:
			returns++
		case SetupMove:
			moves++
		}
	}
	if returns != 1 || moves != 3 {
		t.Fatalf("legal returns %d moves %d, want 1 and 3 (Active to each empty Bench slot)", returns, moves)
	}
}

func mustSetup(t *testing.T, g *Game, ev Event) {
	t.Helper()
	if err := g.Submit(ev); err != nil {
		t.Fatalf("%T: %v", ev, err)
	}
}
