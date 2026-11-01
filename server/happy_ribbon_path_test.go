package main

import (
	"testing"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// #139: Eevee / Eevee ex -> Sylveon ex from hand must offer Happy Ribbon (draw 2),
// as a yes/no prompt (owner decision 2026-10-25), answered with Yes.
func TestHappyRibbonOffersAfterEvolveFromHand(t *testing.T) {
	for _, basic := range []string{"B4-131", "A4b-287"} { // Eevee, Eevee ex
		g, p := abGame(t)
		g.S.Turn = 3
		g.S.Players[p].Deck = []engine.Card{realCard(t, "A4b-069"), realCard(t, "A4b-069")}
		m := put(g, p, 0, realCard(t, basic))
		m.PlacedTurn = 1
		g.S.Players[p].Hand = []engine.Card{realCard(t, "B2-229")}
		var evo engine.Event
		for _, ev := range g.LegalMoves(p) {
			if e, ok := ev.(engine.Evolve); ok {
				evo = e
			}
		}
		if evo == nil {
			t.Fatalf("%s: no legal Evolve", basic)
		}
		mustSubmit(t, g, evo)
		pr := g.S.Pending
		if pr == nil || pr.Kind != engine.PromptChoose || pr.Player != p {
			t.Fatalf("%s: no Happy Ribbon prompt after evolving; pending=%+v", basic, pr)
		}
		mustSubmit(t, g, engine.ChooseOption{Player: p, Index: 0}) // "Use Happy Ribbon"
		if g.S.Pending != nil {
			t.Fatalf("%s: prompt left pending after answering", basic)
		}
		if got := len(g.S.Players[p].Hand); got != 2 {
			t.Errorf("%s: hand = %d, want 2", basic, got)
		}
	}
}
