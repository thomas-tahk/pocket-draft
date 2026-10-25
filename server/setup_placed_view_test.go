package main

import (
	"strings"
	"testing"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// Owner 2026-10-19 (8): during setup you see how many cards the opponent has
// placed, face down; names stay hidden until both are Ready.
func TestSetupViewShowsOpponentPlacedCountWithoutNames(t *testing.T) {
	// Arrange
	g := engine.NewGame(7, viewTestDeck("Alphamon"), viewTestDeck("Bravomon"))
	a, bench := pickBasics(g.S.Players[1].Hand)
	for _, ev := range []engine.Event{
		engine.SetupActive{Player: 1, CardID: a},
		engine.SetupBench{Player: 1, CardID: bench, Slot: engine.BenchLeft},
	} {
		if err := g.Submit(ev); err != nil {
			t.Fatal(err)
		}
	}

	// Act
	v, raw := viewJSON(t, g, 0)

	// Assert
	if got := faceDownCount(v.Players[1]); got != 2 || faceDownCount(v.Players[0]) != 0 {
		t.Fatalf("face-down = %d/%d, want 0/2", faceDownCount(v.Players[0]), got)
	}
	if strings.Contains(raw, "Bravomon") {
		t.Fatalf("view leaks the opponent's picks: %s", raw)
	}
}

// Taking a card back lowers the count the opponent sees, and the history does
// not name the card.
func TestSetupViewCountFollowsReturnsAndMoves(t *testing.T) {
	g := engine.NewGame(7, viewTestDeck("Alphamon"), viewTestDeck("Bravomon"))
	a, bench := pickBasics(g.S.Players[1].Hand)
	for _, ev := range []engine.Event{
		engine.SetupActive{Player: 1, CardID: a},
		engine.SetupBench{Player: 1, CardID: bench, Slot: engine.BenchLeft},
		engine.SetupMove{Player: 1, From: 1, To: 3},
		engine.SetupReturn{Player: 1, Pos: 3},
	} {
		if err := g.Submit(ev); err != nil {
			t.Fatal(err)
		}
	}

	v, raw := viewJSON(t, g, 0)

	if got := faceDownCount(v.Players[1]); got != 1 {
		t.Fatalf("face-down = %d, want 1", got)
	}
	if strings.Contains(raw, "Bravomon") {
		t.Fatalf("view leaks the opponent's picks: %s", raw)
	}
}

// The bot places during setup, before the human is Ready: the first bot move
// of a fresh game is its own, never the human's.
func TestBotPlacesBeforeHumanIsReady(t *testing.T) {
	// Arrange
	game = engine.NewGame(7, viewTestDeck("Alphamon"), viewTestDeck("Bravomon"))

	// Act
	ev, ok := botChoice(game)

	// Assert
	if !ok {
		t.Fatal("bot found no setup move")
	}
	if p := eventPlayer(ev); p != 1 {
		t.Fatalf("first bot move in setup was for P%d, want the bot's seat P2", p+1)
	}
}

func faceDownCount(p playerView) int {
	n := 0
	for _, m := range append([]*inPlayView{p.Active}, p.Bench...) {
		if m != nil && m.FaceDown {
			n++
		}
	}
	return n
}

func pickBasics(hand []engine.Card) (active, bench string) {
	for _, c := range hand {
		switch {
		case !c.IsBasicPokemon():
		case active == "":
			active = c.ID
		case bench == "":
			bench = c.ID
		}
	}
	return active, bench
}

func eventPlayer(ev engine.Event) int { return fromEvent(ev).Player }
