package main

import (
	"regexp"
	"strings"
)

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

// synergyMult mirrors weights.ts's SYNERGY_MULT: a candidate that works with
// the picks (evolution line, or a Trainer naming a picked Pokémon) has its
// weight multiplied by this. One multiplication, never stacked. Energy match is
// not a synergy rule; type weighting already covers it.
const synergyMult = 3.0

// synergy is what the picks make worth boosting. The zero value boosts nothing.
type synergy struct {
	names   map[string]bool // picked Pokémon names
	parents map[string]bool // names the picked Pokémon evolve from
	text    *regexp.Regexp  // a picked Pokémon's base name inside Trainer text
}

// newSynergy mirrors weights.ts's buildSynergy.
func newSynergy(picks []draftCard) synergy {
	s := synergy{names: map[string]bool{}, parents: map[string]bool{}}
	needles := []string{}
	for _, p := range picks {
		if p.CardType == "Trainer" {
			continue
		}
		s.names[p.Name] = true
		if p.EvolvesFrom != "" {
			s.parents[p.EvolvesFrom] = true
		}
		needles = append(needles, regexp.QuoteMeta(strings.TrimSuffix(p.Name, " ex")))
	}
	if len(needles) > 0 {
		s.text = regexp.MustCompile(`(^|[^\p{L}])(` + strings.Join(needles, "|") + `)($|[^\p{L}])`)
	}
	return s
}

// matches reports whether c works with the picks.
func (s synergy) matches(c draftCard) bool {
	if c.CardType == "Trainer" {
		return s.text != nil && s.text.MatchString(c.TrainerText)
	}
	return (c.EvolvesFrom != "" && s.names[c.EvolvesFrom]) || s.parents[c.Name]
}

// weightOf mirrors weights.ts's weightOf: Trainers and Colorless Pokémon are
// flat 1 (always equally eligible); other types get the type-nudge term; then
// a matching synergy multiplies the result once.
func weightOf(c draftCard, tally typeTally, syn synergy) float64 {
	w := 1.0
	if c.CardType != "Trainer" && c.CardType != "Colorless" {
		w = 1 + typeWeightCoef*float64(tally[c.CardType])
	}
	if syn.matches(c) {
		w *= synergyMult
	}
	return w
}
