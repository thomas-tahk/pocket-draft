package main

import "sync"

// Go port of src/draft/weights.ts.

// typeWeightCoef mirrors weights.ts's TYPE_WEIGHT_COEF: each drafted on-type
// Pokémon adds this much to that type's weight (weight = 1 + coef × count).
// Raised from 1 to 2 so the lean shows after 2-3 picks (owner, 2026-10-25,
// point 3); see the measured odds in draft_type_lean_test.go.
const typeWeightCoef = 2.0

// typeSupportCoef mirrors weights.ts's TYPE_SUPPORT_COEF: the same lean for a
// Trainer/Item that supports a type (weight = 1 + coef × picks of that type).
const typeSupportCoef = 2.0

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

// supportTypesCache memoises supportTypes by card id.
var supportTypesCache sync.Map

// supportTypes mirrors weights.ts's supportedTypes: the types a Trainer/Item
// supports. Its text naming a type counts (a Trainer has no type of its own,
// so the own-type trap cannot apply); so does typeSupportTable, minus
// notTypeSupport. Pokémon support nothing.
func supportTypes(c draftCard) []string {
	if c.CardType != "Trainer" || notTypeSupport[c.Name] {
		return nil
	}
	if cached, ok := supportTypesCache.Load(c.ID); ok {
		return cached.([]string)
	}
	seen := map[string]bool{}
	out := []string{}
	for _, typ := range append(append([]string{}, namedTypes(c)...), typeSupportTable[c.Name]...) {
		if !seen[typ] {
			seen[typ] = true
			out = append(out, typ)
		}
	}
	supportTypesCache.Store(c.ID, out)
	return out
}

// weightOf mirrors weights.ts's weightOf: Colorless Pokémon are flat 1; other
// types get 1 + coef × picks of that type; a Trainer gets the same lean from
// the most-picked type it supports (flat 1 if it supports none); then a
// matching synergy multiplies the result once.
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
		w = 1 + typeSupportCoef*float64(most)
	default:
		w = 1 + typeWeightCoef*float64(tally[c.CardType])
	}
	if syn.matches(c) {
		w *= synergyMult
	}
	return w
}
