package engine

import "testing"

// An Ability op that takes a target is offered once per target, and its gate
// is asked (HealType aims at a damaged Grass Pokémon; DrawCards refuses at 10).
func TestAbilityTargetsAndGates(t *testing.T) {
	// Arrange
	g := abilityGame(AnySpot, true)
	pl := &g.S.Players[0]
	pl.Active.Card.Ability.Use = []EffectOp{HealType{Amount: 30, Type: Grass}}
	pl.Bench[BenchLeft].Card.Ability = nil
	pl.Bench[BenchLeft].Damage = 20
	pl.Bench[BenchLeft].Card.Type = Grass

	// Act
	got := abilityMoves(g, 0)
	err := g.Submit(UseAbility{0, 0, Spot{0, 1 + BenchLeft}})

	// Assert
	assertMoves(t, got, []Event{UseAbility{0, 0, Spot{0, 1 + BenchLeft}}})
	if err != nil || pl.Bench[BenchLeft].Damage != 0 {
		t.Errorf("heal: err %v, damage %d; want healed", err, pl.Bench[BenchLeft].Damage)
	}
	pl.Active.Card.Ability.Use = []EffectOp{DrawCards{N: 1}}
	pl.Active.AbilityUsedTurn = 0
	for len(pl.Hand) < handMax {
		pl.Hand = append(pl.Hand, legalBasic("H", "H", 1))
	}
	assertMoves(t, abilityMoves(g, 0), nil)
}
