package main

import (
	"testing"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// The end screen (spec F19) reads the result and its reason from the view:
// each seat sees its own win or loss, and why the game ended.
func TestViewReportsOutcomeAndReasonPerViewer(t *testing.T) {
	fire, lightning := engine.DemoDecks()
	g := engine.NewGame(1, fire, lightning)

	before := toGameView(g, 0)
	if before.Outcome != "" || before.Reason != "" {
		t.Fatalf("unfinished game: outcome %q reason %q, want both empty", before.Outcome, before.Reason)
	}

	if err := g.Submit(engine.Concede{Player: 0}); err != nil {
		t.Fatalf("concede: %v", err)
	}
	for viewer, want := range []string{"loss", "win"} {
		v := toGameView(g, viewer)
		if v.Outcome != want || v.Reason != "concede" {
			t.Fatalf("viewer %d: outcome %q reason %q, want %q concede", viewer, v.Outcome, v.Reason, want)
		}
	}
}
