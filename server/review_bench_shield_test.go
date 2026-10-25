package main

import (
	"testing"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// A Benched Pokémon with Protective Poncho takes no damage from an Ability.
func TestWaterShurikenSkipsPonchoBench(t *testing.T) {
	g, p, _ := abilityGame45(t, "A1-089") // Greninja
	bench := onBoard(t, g, 1-p, 1, "A1-001")
	tool(t, bench, "B2-147")

	_ = g.Submit(engine.UseAbility{Player: p, Pos: 0, Target: engine.Spot{Player: 1 - p, Pos: 1}})

	if bench.Damage != 0 {
		t.Errorf("shielded bench took %d damage", bench.Damage)
	}
}
