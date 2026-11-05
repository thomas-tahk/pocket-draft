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

// #139 review: Yes on a targeted on-evolve Ability asks which Pokémon.
// Milotic's Healing Ripples with two damaged [W] Pokémon.
func TestHealingRipplesYesAsksForATarget(t *testing.T) {
	g, p := abGame(t)
	g.S.Turn = 3
	feebas := put(g, p, 0, realCard(t, "A4a-021"))
	feebas.PlacedTurn, feebas.Damage = 1, 20
	sobble := put(g, p, 1, realCard(t, "B3-048"))
	sobble.Damage = 30
	g.S.Players[p].Hand = []engine.Card{realCard(t, "A4a-022")}
	mustSubmit(t, g, engine.Evolve{Player: p, HandCardID: "A4a-022", Target: 0})
	mustSubmit(t, g, engine.ChooseOption{Player: p, Index: 0}) // Use Healing Ripples
	pr := g.S.Pending
	if pr == nil || pr.Kind != engine.PromptChoose || len(pr.Labels) != 2 {
		t.Fatalf("want a 2-target choice after Yes, got %+v", pr)
	}
	mustSubmit(t, g, engine.ChooseOption{Player: p, Index: 1}) // the Bench Sobble
	if sobble.Damage != 0 || feebas.Damage != 20 {
		t.Errorf("Sobble damage %d (want 0), Milotic damage %d (want 20)", sobble.Damage, feebas.Damage)
	}
}

// #139 review: Rare Candy puts the Stage 2 from your hand, so it prompts too.
func TestRareCandyEvolvePromptsOnEvolveAbility(t *testing.T) {
	g, p := itemGame(t, "A4b-314")
	g.S.Turn = 3
	g.S.Players[p].Hand = append(g.S.Players[p].Hand, realCard(t, "B3-050")) // Inteleon
	sobble := onBoard(t, g, p, 1, "B3-048")
	sobble.PlacedTurn = 1
	playItem(t, g, p, "A4b-314", engine.Spot{Player: p, Pos: 1})
	pr := g.S.Pending
	if pr == nil || pr.Kind != engine.PromptChoose || pr.Title != "Use Swift Shot?" {
		t.Fatalf("no Swift Shot prompt after Rare Candy; pending=%+v", pr)
	}
	mustSubmit(t, g, engine.ChooseOption{Player: p, Index: 0})
	if d := g.S.Players[1-p].Active.Damage; d != 30 {
		t.Errorf("foe damage = %d, want 30", d)
	}
}
