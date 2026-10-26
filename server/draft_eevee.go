package main

// Go port of src/draft/eeveeBoost.ts.
//
// Eevee ex ("Veevee ’volve") can evolve into any Eeveelution, so once one is
// picked, its odds rise by the same ×3 a synergy match gets (owner 2026-10-25,
// "Eevee ex becomes draftable"). Raised odds, never a guarantee.

const eeveeEx = "Eevee ex"

// eeveeMult is the same value as synergyMult.
const eeveeMult = 3.0

// isEeveelution reports whether c is a Pokémon that evolves from Eevee.
func isEeveelution(c draftCard) bool {
	return c.CardType != "Trainer" && c.EvolvesFrom == "Eevee"
}

// eeveeBoost is the multiplier on one candidate: eeveeMult for Eevee ex once an
// Eeveelution is among the picks, else 1. alreadyBoosted (the candidate already
// got its synergy multiplier) keeps the boost from stacking.
func eeveeBoost(c draftCard, eeveelutionPicked, alreadyBoosted bool) float64 {
	if eeveelutionPicked && !alreadyBoosted && c.Name == eeveeEx {
		return eeveeMult
	}
	return 1
}
