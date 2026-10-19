package main

import "github.com/thomas-tahk/pocket-draft/engine"

// Ability texts that override a rule, batch 4.2 (issue #75).

func init() {
	for text, a := range texts42 {
		abilityTexts[text] = a
	}
}

var texts42 = map[string]engine.Ability{
	// Eevee, Boosted Evolution
	"As long as this Pokémon is in the Active Spot, it can evolve during your first turn or the turn you play it.": {
		Passive: []engine.Passive{{Hook: engine.HookEvolveEarly, When: func(c engine.PassiveCheck) bool {
			return c.Self() && c.Target.Pos == 0
		}}},
	},
	// Alolan Muk, Power of Alchemy
	"Basic Pokémon in play (both yours and your opponent's) have no Abilities.": {
		Passive: []engine.Passive{{Hook: engine.HookNoAbilities, When: func(c engine.PassiveCheck) bool {
			return c.Target.Card.Stage == engine.Basic
		}}},
	},
}
