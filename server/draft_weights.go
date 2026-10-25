package main

// Go port of src/draft/weights.ts.

// typeWeightCoef mirrors weights.ts's TYPE_WEIGHT_COEF: each drafted on-type
// Pokémon adds this much to that type's weight (weight = 1 + coef × count).
// Back on (1.0) since the 2026-10-15 rework answers: offers lean toward the
// types you've picked.
const typeWeightCoef = 1.0

type typeTally map[string]int

// tallyTypes mirrors weights.ts's tallyTypes.
func tallyTypes(picks []draftCard) typeTally {
	t := typeTally{}
	for _, p := range picks {
		if p.CardType == "Trainer" || p.CardType == "Colorless" {
			continue
		}
		t[p.CardType]++
	}
	return t
}

// weightOf mirrors weights.ts's weightOf: Trainers and Colorless Pokémon are
// flat 1 (always equally eligible); other types get the type-nudge term.
func weightOf(c draftCard, tally typeTally) float64 {
	if c.CardType == "Trainer" || c.CardType == "Colorless" {
		return 1
	}
	return 1 + typeWeightCoef*float64(tally[c.CardType])
}
