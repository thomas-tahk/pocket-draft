package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// readyUp plays the bot's setup steps for p until p is Ready.
func readyUp(t *testing.T, g *engine.Game, p int) {
	t.Helper()
	for !g.S.Ready(p) {
		ev, ok := bestMove(g, g.LegalMoves(p))
		if !ok {
			t.Fatalf("P%d setup: bot found no move", p+1)
		}
		if err := g.Submit(ev); err != nil {
			t.Fatalf("P%d setup: %v", p+1, err)
		}
	}
}

func viewJSON(t *testing.T, g *engine.Game, viewer int) (gameView, string) {
	t.Helper()
	v := toGameView(g, viewer)
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return v, string(raw)
}

// Spec S5: the opponent's setup picks are hidden until both players are Ready,
// whichever of them is Ready first.
func TestNeitherViewShowsTheOthersPicksUntilBothReady(t *testing.T) {
	for first := 0; first < 2; first++ {
		// Arrange
		names := [2]string{"Alphamon", "Bravomon"}
		g := engine.NewGame(7, viewTestDeck(names[0]), viewTestDeck(names[1]))
		other := 1 - first

		// Act
		readyUp(t, g, first)

		// Assert: neither card ids nor names (narration included) leak.
		v, raw := viewJSON(t, g, other)
		if strings.Contains(raw, names[first]) {
			t.Fatalf("P%d Ready first: P%d's view leaks the picks: %s", first+1, other+1, raw)
		}
		if !v.Players[first].Ready {
			t.Fatalf("P%d's view should still show P%d is Ready", other+1, first+1)
		}
		if own, _ := viewJSON(t, g, first); own.Players[first].Active == nil {
			t.Fatalf("P%d can't see their own Active", first+1)
		}

		readyUp(t, g, other)
		for viewer := 0; viewer < 2; viewer++ {
			v, _ := viewJSON(t, g, viewer)
			if v.Players[0].Active == nil || v.Players[1].Active == nil {
				t.Fatalf("both Ready: P%d's view is missing an Active", viewer+1)
			}
		}
	}
}
