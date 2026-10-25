package engine

import "testing"

// HealTarget targets only your own damaged Pokémon and heals at most N.
func TestHealTarget(t *testing.T) {
	potion := Card{ID: "P1", Name: "Potion", Kind: KindItem, Trainer: []EffectOp{HealTarget{N: 20}}}
	g, p := mainPhaseWith(t, potion)
	at := Spot{p, 0}
	if isLegal(g, p, PlayTrainer{Player: p, CardID: "P1", Target: at}) {
		t.Fatal("Potion legal on an undamaged Pokémon")
	}
	g.S.Players[p].Active.Damage = 30
	g.S.Players[1-p].Active.Damage = 30
	if isLegal(g, p, PlayTrainer{Player: p, CardID: "P1", Target: Spot{1 - p, 0}}) {
		t.Error("Potion legal on the opponent's Pokémon")
	}

	mustSubmit(t, g, PlayTrainer{Player: p, CardID: "P1", Target: at})

	if d := g.S.Players[p].Active.Damage; d != 10 {
		t.Errorf("damage after Potion = %d, want 10", d)
	}
}
