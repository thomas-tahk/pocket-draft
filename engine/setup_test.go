package engine

import "testing"

func placeSetup(t *testing.T, g *Game, p int) {
	t.Helper()
	a, b := pickSetup(g.S.Players[p].Hand)
	if err := g.Submit(SetupPlace{Player: p, ActiveCardID: a, BenchCardIDs: b}); err != nil {
		t.Fatalf("P%d setup: %v", p+1, err)
	}
}

// Spec S5: setup is simultaneous, so either player may press Ready first, and
// the game starts only once both have.
func TestSetupEitherPlayerMayBeReadyFirst(t *testing.T) {
	// Arrange
	d0, d1 := twoDecks()
	g := NewGame(3, d0, d1)

	// Act
	placeSetup(t, g, 1)

	// Assert
	if !g.S.Ready(1) || g.S.Ready(0) {
		t.Fatalf("ready = %v/%v, want P2 only", g.S.Ready(0), g.S.Ready(1))
	}
	if g.S.Phase != PhaseSetup {
		t.Fatalf("phase = %v, want setup until both are Ready", g.S.Phase)
	}
	placeSetup(t, g, 0)
	if g.S.Phase != PhaseMain || g.S.Turn != 1 || g.S.Active != g.S.FirstPlayer {
		t.Fatalf("after both Ready: phase %v turn %d active %d, want main/1/%d", g.S.Phase, g.S.Turn, g.S.Active, g.S.FirstPlayer)
	}
}

// Ready is final: a player who is Ready can't place again.
func TestSetupReadyTwiceIsRejected(t *testing.T) {
	// Arrange
	d0, d1 := twoDecks()
	g := NewGame(3, d0, d1)
	placeSetup(t, g, 0)
	a, b := pickSetup(g.S.Players[0].Hand)

	// Act
	err := g.Submit(SetupPlace{Player: 0, ActiveCardID: a, BenchCardIDs: b})

	// Assert
	if err == nil {
		t.Fatal("a second SetupPlace from a Ready player was accepted")
	}
}
