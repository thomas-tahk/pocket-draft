package main

import (
	"testing"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// Giant Cape's extra HP counts when an Ability damages a Benched Pokémon.
func TestBenchedGiantCapeSurvivesAbilityDamage(t *testing.T) {
	g, p, _ := abilityGame45(t, "A1-089")
	bench := onBoard(t, g, 1-p, 1, torkoal) // 100 HP
	tool(t, bench, "A2-147")                // +20
	bench.Damage = 90

	useAbility(t, g, p, 0, engine.Spot{Player: 1 - p, Pos: 1}) // 110 damage

	if g.S.Players[1-p].Bench[0] != bench || g.S.Players[p].Points != 0 {
		t.Error("Benched Torkoal with Giant Cape was knocked out at 110 of 120")
	}
}
