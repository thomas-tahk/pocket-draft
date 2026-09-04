package main

import (
	"strings"
	"testing"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// End-to-end verification for effect-engine slice 1 (docs/superpowers/specs/
// 2026-07-13-effect-engine-slice-1-design.md): real card data goes through the
// actual data bridge (toEngineCard) into a real Game, and the demo criterion
// the spec names — a coin-flip attack, a burn, and a draw all fire and are
// visible on the board — is reproduced the way a player would trigger it,
// not just asserted against an EffectContext in isolation.

// dummyOpponent is a plain Basic with no weakness and no effect, standing in
// for "whatever the other player is doing" — these tests only exercise the
// attacker's side of applyAttack.
func dummyOpponent(t *testing.T) engine.Card {
	t.Helper()
	card, err := toEngineCard(rawCard{
		ID: "DUMMY", Name: "Dummy", CardType: "Colorless", Stage: "Basic", HP: 200,
		Attacks: []rawAttack{{Cost: "C", Name: "Poke", Damage: "10"}},
	})
	if err != nil {
		t.Fatalf("dummy opponent: %v", err)
	}
	return card
}

func deckOfCard(c engine.Card, n int) []engine.Card {
	d := make([]engine.Card, n)
	for i := range d {
		d[i] = c
	}
	return d
}

// doSetup places the first card in each player's opening hand as their Active
// with an empty bench — every deck in these tests is homogeneous, so any card
// in hand is the one under test.
func doSetup(t *testing.T, g *engine.Game) {
	t.Helper()
	for i := 0; i < 2 && g.S.Pending != nil; i++ {
		p := g.S.Pending.Player
		id := g.S.Players[p].Hand[0].ID
		if err := g.Submit(engine.SetupPlace{Player: p, ActiveCardID: id}); err != nil {
			t.Fatalf("setup for player %d: %v", p, err)
		}
	}
}

// waitForAttackerReady drives turns (attaching energy, ending turns for the
// non-attacker) until attacker's Active can pay attack index 0, then returns
// — leaving the actual UseAttack to the caller so a test can measure state
// immediately before it fires. It fails the test if a prompt or a knockout
// intervenes first — these fixtures are built so that never happens within
// the step budget.
func waitForAttackerReady(t *testing.T, g *engine.Game, attacker int) {
	t.Helper()
	for step := 0; step < 20; step++ {
		if g.S.Phase == engine.PhaseOver {
			t.Fatalf("game ended before player %d became ready to attack", attacker)
		}
		if g.S.Pending != nil {
			t.Fatalf("unexpected pending prompt: %+v", g.S.Pending)
		}
		p := g.S.Active
		if p != attacker {
			if err := g.Submit(engine.EndTurn{Player: p}); err != nil {
				t.Fatalf("end turn: %v", err)
			}
			continue
		}
		pl := g.S.Players[p]
		cost := len(pl.Active.Card.Attacks[0].Cost)
		total := 0
		for _, n := range pl.Active.Energy {
			total += n
		}
		if total < cost {
			if pl.EnergyZone != engine.NoEnergy && !pl.EnergyUsed {
				if err := g.Submit(engine.AttachEnergy{Player: p, Target: 0}); err != nil {
					t.Fatalf("attach energy: %v", err)
				}
				continue
			}
			if err := g.Submit(engine.EndTurn{Player: p}); err != nil {
				t.Fatalf("end turn: %v", err)
			}
			continue
		}
		return
	}
	t.Fatalf("player %d never became ready to attack within the step budget", attacker)
}

// Sneasel's real printing: Quick Attack (10+) visibly varies with a coin flip.
func TestEndToEndSneaselQuickAttackCoinFlip(t *testing.T) {
	sneasel, err := toEngineCard(rawCard{
		ID: "B3a-038", Name: "Sneasel", CardType: "Darkness", Stage: "Basic", HP: 60,
		Attacks: []rawAttack{{Cost: "D", Name: "Quick Attack", Damage: "10+", Effect: "Flip a coin. If heads, this attack does 20 more damage."}},
	})
	if err != nil {
		t.Fatalf("toEngineCard: %v", err)
	}
	deck0 := deckOfCard(sneasel, 20)
	deck1 := deckOfCard(dummyOpponent(t), 20)

	g := engine.NewGame(1, deck0, deck1)
	doSetup(t, g)
	waitForAttackerReady(t, g, 0)
	beforeNarration := len(g.Narration)
	if err := g.Submit(engine.UseAttack{Player: 0, Index: 0}); err != nil {
		t.Fatalf("attack: %v", err)
	}

	dealt := g.S.Players[1].Active.Damage
	if dealt != 10 && dealt != 30 {
		t.Errorf("damage dealt = %d, want 10 (tails) or 30 (heads)", dealt)
	}
	if !containsSubstring(g.Narration[beforeNarration:], "flipped") {
		t.Errorf("narration = %v, want a line naming the coin-flip outcome", g.Narration[beforeNarration:])
	}
}

func containsSubstring(lines []string, substr string) bool {
	for _, l := range lines {
		if strings.Contains(l, substr) {
			return true
		}
	}
	return false
}

// Ponyta's real printing: Singe leaves the opponent's Active Burned.
func TestEndToEndPonytaSingeBurns(t *testing.T) {
	ponyta, err := toEngineCard(rawCard{
		ID: "B2b-010", Name: "Ponyta", CardType: "Fire", Stage: "Basic", HP: 60,
		Attacks: []rawAttack{{Cost: "R", Name: "Singe", Damage: "", Effect: "Your opponent's Active Pokémon is now Burned."}},
	})
	if err != nil {
		t.Fatalf("toEngineCard: %v", err)
	}
	deck0 := deckOfCard(ponyta, 20)
	deck1 := deckOfCard(dummyOpponent(t), 20)

	g := engine.NewGame(1, deck0, deck1)
	doSetup(t, g)
	waitForAttackerReady(t, g, 0)
	beforeNarration := len(g.Narration)
	beforeDamage := g.S.Players[1].Active.Damage
	if err := g.Submit(engine.UseAttack{Player: 0, Index: 0}); err != nil {
		t.Fatalf("attack: %v", err)
	}

	// Attacking ends the turn, so concludeTurn's Pokémon-Checkup resolves the
	// burn Singe just inflicted in this same call — same as a real Pokémon
	// Checkup right after the attacker's turn. That tick (20 damage) proves
	// Burned was set even though a 50% recovery roll may have since cleared
	// the flag, so check the damage it dealt rather than the flag itself.
	if got, want := g.S.Players[1].Active.Damage-beforeDamage, 20; got != want {
		t.Errorf("burn tick dealt %d damage, want %d", got, want)
	}
	if !containsSubstring(g.Narration[beforeNarration:], "took 20 from burn") {
		t.Errorf("narration = %v, want a burn-tick line", g.Narration[beforeNarration:])
	}
}

// Munchlax's real printing: Hungrily Draw (free) draws a card.
func TestEndToEndMunchlaxDraws(t *testing.T) {
	munchlax, err := toEngineCard(rawCard{
		ID: "B3b-054", Name: "Munchlax", CardType: "Colorless", Stage: "Basic", HP: 50,
		Attacks: []rawAttack{{Cost: "0", Name: "Hungrily Draw", Damage: "10", Effect: "Draw a card."}},
	})
	if err != nil {
		t.Fatalf("toEngineCard: %v", err)
	}
	deck0 := deckOfCard(munchlax, 20)
	deck1 := deckOfCard(dummyOpponent(t), 20)

	g := engine.NewGame(1, deck0, deck1)
	doSetup(t, g)
	waitForAttackerReady(t, g, 0)

	before := len(g.S.Players[0].Hand)
	if err := g.Submit(engine.UseAttack{Player: 0, Index: 0}); err != nil {
		t.Fatalf("attack: %v", err)
	}

	if got, want := len(g.S.Players[0].Hand), before+1; got != want {
		t.Errorf("hand size after attacking = %d, want %d (drew a card)", got, want)
	}
}
