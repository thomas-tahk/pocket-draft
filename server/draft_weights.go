package main

// Go port of src/draft/weights.ts.

// typeWeightCoef mirrors weights.ts:14's TYPE_WEIGHT_COEF. 0 means true random
// draft — no type bias; this is currently a dead no-op term client-side too, so
// it's ported as-is rather than spent time on (see issue #9's implementation
// notes).
const typeWeightCoef = 0.0

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
// flat 1 (always equally eligible); other types get the (currently inert)
// type-nudge term.
func weightOf(c draftCard, tally typeTally) float64 {
	if c.CardType == "Trainer" || c.CardType == "Colorless" {
		return 1
	}
	return 1 + typeWeightCoef*float64(tally[c.CardType])
}
