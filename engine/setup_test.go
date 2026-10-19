package engine

import "testing"

func placeSetup(t *testing.T, g *Game, p int) {
	t.Helper()
	a, b := pickSetup(g.S.Players[p].Hand)
	if err := submitSetup(g, p, a, b); err != nil {
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

// Ready is final: a Ready player can't place more.
func TestSetupNothingAfterReady(t *testing.T) {
	// Arrange
	d0, d1 := twoDecks()
	g := NewGame(3, d0, d1)
	placeSetup(t, g, 0)

	// Act
	moves := g.LegalMoves(0)

	// Assert
	if len(moves) != 1 || moves[0] != (Concede{Player: 0}) {
		t.Fatalf("a Ready player may still %v", moves)
	}
}

// Ready needs an Active.
func TestSetupReadyNeedsAnActive(t *testing.T) {
	// Arrange
	d0, d1 := twoDecks()
	g := NewGame(3, d0, d1)

	// Act
	err := g.Submit(SetupReady{Player: 0})

	// Assert
	if err == nil {
		t.Fatal("Ready without an Active was accepted")
	}
}
