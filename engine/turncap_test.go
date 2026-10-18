package engine

import "testing"

// endTurnAt fast-forwards a set-up game to the given turn with a 2–0 score,
// then ends that turn.
func endTurnAt(t *testing.T, g *Game, turn int) {
	t.Helper()
	g.S.Turn = turn
	g.S.Players[g.S.Active].Points = 2
	if err := g.Submit(EndTurn{Player: g.S.Active}); err != nil {
		t.Fatalf("end turn: %v", err)
	}
}

func assertTurnCapTie(t *testing.T, g *Game) {
	t.Helper()
	if g.S.Phase != PhaseOver {
		t.Fatalf("phase = %v, want over", g.S.Phase)
	}
	if g.S.Winner != -1 || g.S.Reason != ReasonTurnCap {
		t.Fatalf("winner %d reason %v, want tie by turn cap", g.S.Winner, g.S.Reason)
	}
	for p := 0; p < 2; p++ {
		if got := g.S.Outcome(p); got != Tie {
			t.Errorf("%s outcome = %v, want Tie", side(p), got)
		}
	}
}

// M3/M4: against the bot (the default), the game ties when turn 50 ends, even at 2–0.
func TestBotGameTiesWhenTurn50Ends(t *testing.T) {
	d0, d1 := twoDecks()
	g := NewGame(42, d0, d1)
	doSetup(t, g)

	endTurnAt(t, g, 49)
	if g.S.Phase == PhaseOver {
		t.Fatal("game ended after turn 49; cap is 50 against the bot")
	}
	endTurnAt(t, g, 50)
	assertTurnCapTie(t, g)
}

// M3: against a player, the game ties when turn 30 ends, even at 2–0.
func TestPvPGameTiesWhenTurn30Ends(t *testing.T) {
	d0, d1 := twoDecks()
	g := NewGame(42, d0, d1)
	g.S.Mode = VsPlayer
	doSetup(t, g)

	endTurnAt(t, g, 29)
	if g.S.Phase == PhaseOver {
		t.Fatal("game ended after turn 29; cap is 30 against a player")
	}
	endTurnAt(t, g, 30)
	assertTurnCapTie(t, g)
}

// A decided game reports Win/Loss with its reason.
func TestConcedeOutcomeIsWinLoss(t *testing.T) {
	d0, d1 := twoDecks()
	g := NewGame(42, d0, d1)
	doSetup(t, g)
	p := g.S.Active
	if err := g.Submit(Concede{Player: p}); err != nil {
		t.Fatal(err)
	}
	if g.S.Outcome(p) != Loss || g.S.Outcome(1-p) != Win || g.S.Reason != ReasonConcede {
		t.Fatalf("outcomes %v/%v reason %v", g.S.Outcome(p), g.S.Outcome(1-p), g.S.Reason)
	}
}
