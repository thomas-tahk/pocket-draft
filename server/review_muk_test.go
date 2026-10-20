package main

import (
	"testing"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// Alolan Muk turns off Druddigon's Rough Skin.
func TestMukStopsRoughSkin(t *testing.T) {
	g, p := abGame(t)
	atk := put(g, p, 0, realCard(t, "B1a-006"))
	atk.Energy[engine.Grass] = 1
	put(g, 1-p, 0, realCard(t, "A1a-056"))
	put(g, p, 1, realCard(t, "B2-097")) // Alolan Muk

	mustSubmit(t, g, engine.UseAttack{Player: p, Index: 0})

	if atk.Damage != 0 {
		t.Errorf("attacker damage = %d, want 0 (Abilities off)", atk.Damage)
	}
}
