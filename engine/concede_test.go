package engine

import "testing"

func assertConcededBy(t *testing.T, g *Game, p int) {
	t.Helper()
	if g.S.Phase != PhaseOver || g.S.Reason != ReasonConcede {
		t.Fatalf("phase=%v reason=%v, want over by concede", g.S.Phase, g.S.Reason)
	}
	if g.S.Outcome(p) != Loss || g.S.Outcome(1-p) != Win {
		t.Fatalf("outcome for conceder = %v, opponent = %v; want Loss/Win", g.S.Outcome(p), g.S.Outcome(1-p))
	}
	if g.S.Pending != nil {
		t.Errorf("pending prompt %+v should be cleared", g.S.Pending)
	}
}

// Spec E7: concede is allowed at any time — including on the opponent's turn.
func TestConcedeDuringOpponentsTurnEndsGame(t *testing.T) {
	d0, d1 := twoDecks()
	g := NewGame(42, d0, d1)
	doSetup(t, g)
	p := 1 - g.S.Active

	if err := g.Submit(Concede{Player: p}); err != nil {
		t.Fatalf("concede on opponent's turn should be legal: %v", err)
	}
	assertConcededBy(t, g, p)
}

// Concede is accepted while a setup prompt is pending, from either player.
func TestConcedeDuringSetupPrompt(t *testing.T) {
	d0, d1 := twoDecks()
	for p := 0; p < 2; p++ {
		g := NewGame(42, d0, d1)
		if g.S.Pending == nil {
			t.Fatal("expected a setup prompt at game start")
		}
		if err := g.Submit(Concede{Player: p}); err != nil {
			t.Fatalf("player %d concede during setup: %v", p, err)
		}
		assertConcededBy(t, g, p)
	}
}

// Concede is accepted while the other player owes a new-Active prompt.
func TestConcedeDuringOpponentsPrompt(t *testing.T) {
	d0, d1 := twoDecks()
	g := NewGame(5, d0, d1)
	doSetup(t, g)
	p := g.S.Active
	opp := 1 - p
	oppActive := g.S.Players[opp].Active
	oppActive.Poisoned = true
	oppActive.Damage = oppActive.Card.HP - 5
	if err := g.Submit(EndTurn{Player: p}); err != nil {
		t.Fatalf("end turn: %v", err)
	}
	if g.S.Pending == nil || g.S.Pending.Player != opp {
		t.Fatalf("expected a prompt for %s, got %+v", side(opp), g.S.Pending)
	}

	if err := g.Submit(Concede{Player: p}); err != nil {
		t.Fatalf("concede while opponent owes a prompt: %v", err)
	}
	assertConcededBy(t, g, p)
}

// A concede off-turn survives replay.
func TestConcedeOffTurnReplays(t *testing.T) {
	d0, d1 := twoDecks()
	g := NewGame(42, d0, d1)
	doSetup(t, g)
	if err := g.Submit(Concede{Player: 1 - g.S.Active}); err != nil {
		t.Fatal(err)
	}
	r, err := Replay(42, d0, d1, g.Events)
	if err != nil {
		t.Fatalf("replay: %v", err)
	}
	if r.S.Winner != g.S.Winner || r.S.Reason != ReasonConcede {
		t.Errorf("replay winner=%d reason=%v, want %d concede", r.S.Winner, r.S.Reason, g.S.Winner)
	}
}

// Only a seated player can concede.
func TestConcedeRejectsUnknownPlayer(t *testing.T) {
	d0, d1 := twoDecks()
	g := NewGame(42, d0, d1)
	doSetup(t, g)
	if err := g.Submit(Concede{Player: 2}); err == nil {
		t.Error("concede by a non-player should be refused")
	}
}
