package main

import (
	"testing"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// A Pokémon's Ability loads from abilityTexts by its printed wording, named
// from the card; without a wording there it loads bare and doesn't work.
func TestAbilityLoadsByWording(t *testing.T) {
	useRealCards(t)
	rc := rawByID["A2-034"] // Regice, Crystal Body
	wording := rc.Ability.Effect
	saved, had := abilityTexts[wording]
	defer func() {
		delete(abilityTexts, wording)
		if had {
			abilityTexts[wording] = saved
		}
	}()

	delete(abilityTexts, wording)
	if c, _ := toEngineCard(rc); c.Ability != nil || textWorks(rc) {
		t.Fatalf("unmapped Ability: got %v, works %v; want none, not working", c.Ability, textWorks(rc))
	}

	abilityTexts[wording] = engine.Ability{Use: []engine.EffectOp{engine.DrawCards{N: 1}}}
	c, _ := toEngineCard(rc)
	if c.Ability == nil || c.Ability.Name != "Crystal Body" {
		t.Fatalf("mapped Ability = %+v, want one named Crystal Body", c.Ability)
	}
}
