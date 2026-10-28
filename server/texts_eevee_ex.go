package main

import "github.com/thomas-tahk/pocket-draft/engine"

// Eevee ex, Veevee ’volve (issue #128). It is evolved by hand from the player's
// hand, so only the evolve rule needs to know it takes any Eevee evolution; the
// parenthetical (no evolving on your first turn or the turn it is played) is the
// ordinary rule, so nothing else is needed.

func init() {
	abilityTexts["This Pokémon can evolve into any Pokémon that evolves from Eevee if you play it from your hand onto this Pokémon. (This Pokémon can't evolve during your first turn or the turn you play it.)"] = engine.Ability{
		Name:      "Veevee ’volve",
		EvolvesAs: "Eevee",
	}
}
