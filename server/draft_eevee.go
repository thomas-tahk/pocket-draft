package main

// Go port of src/draft/eeveeBoost.ts.
//
// Eevee ex ("Veevee ’volve") can evolve into any Eeveelution, so the two raise
// each other's odds, both directions (owner 2026-10-25, "Eevee ex becomes
// draftable"; 2026-10-28, decision 4): once an Eeveelution is picked, Eevee ex
// gets the same ×3 a synergy match gets, and once Eevee ex is picked, every
// Eeveelution does. Raised odds, never a guarantee, and never stacked with
// another synergy match.

const eeveeEx = "Eevee ex"

// eeveeMult is the same value as synergyMult.
const eeveeMult = 3.0

// isEeveelution reports whether c is a Pokémon that evolves from Eevee.
func isEeveelution(c draftCard) bool {
	return c.CardType != "Trainer" && c.EvolvesFrom == "Eevee"
}

// eeveeBoost is the multiplier on one candidate: eeveeMult for Eevee ex once an
// Eeveelution is among the picks, or for an Eeveelution once Eevee ex is; else
// 1. alreadyBoosted (the candidate already got its synergy multiplier) keeps
// the boost from stacking.
func eeveeBoost(c draftCard, syn synergy, alreadyBoosted bool) float64 {
	if alreadyBoosted {
		return 1
	}
	if (syn.eeveelution && c.Name == eeveeEx) || (syn.eeveeEx && isEeveelution(c)) {
		return eeveeMult
	}
	return 1
}
