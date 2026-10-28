package main

// Go port of src/draft/weights.ts.

// typeWeightCoef mirrors weights.ts's TYPE_WEIGHT_COEF: each drafted on-type
// Pokémon adds this much to that type's weight (weight = 1 + coef × count,
// count capped at typeLeanCap). Raised from 1 to 2 so the lean shows after 2-3
// picks (owner, 2026-10-25, point 3); see the measured odds in
// draft_type_lean_test.go.
const typeWeightCoef = 2.0

// typeSupportCoef mirrors weights.ts's TYPE_SUPPORT_COEF: the same lean for a
// Trainer/Item that supports a type (weight = 1 + coef × picks of that type,
// capped at typeLeanCap).
const typeSupportCoef = 2.0

// typeLeanCap mirrors weights.ts's TYPE_LEAN_CAP: picks of one type past this
// many add no more lean, so off-type cards never become near-impossible over a
// mono-type draft (owner 2026-10-28, decision 1). The lean at 2-3 picks is
// unchanged; the weight tops out at 1 + 2 × 4 = 9. Measured on the real pool
// over a mono-Water draft, other-typed Pokémon are 28.5% of offer slots after
// 3 picks; uncapped they fall to 15.2% after 12 and 12.8% after 19, capped at
// 4 they stay at 28.0% and 30.8% (TestTypeLeanIsCapped).
const typeLeanCap = 4

// leanCount is the number of picks of a type that counts for its lean.
func leanCount(n int) float64 { return float64(min(n, typeLeanCap)) }

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

// weightOf mirrors weights.ts's weightOf: Colorless Pokémon are flat 1; other
// types get 1 + coef × picks of that type (capped); a Trainer gets the same
// lean from the most-picked type it supports (flat 1 if it supports none);
// then a matching synergy multiplies the result once.
func weightOf(c draftCard, tally typeTally, syn synergy) float64 {
	w := 1.0
	switch c.CardType {
	case "Colorless":
	case "Trainer":
		most := 0
		for _, typ := range supportTypes(c) {
			if tally[typ] > most {
				most = tally[typ]
			}
		}
		w = 1 + typeSupportCoef*leanCount(most)
	default:
		w = 1 + typeWeightCoef*leanCount(tally[c.CardType])
	}
	matched := syn.matches(c)
	if matched {
		w *= synergyMult
	}
	return w * eeveeBoost(c, syn, matched)
}
