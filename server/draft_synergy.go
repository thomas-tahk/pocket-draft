package main

import (
	"regexp"
	"strings"
)

// Go port of the synergy half of src/draft/weights.ts.

// synergyMult mirrors weights.ts's SYNERGY_MULT: a candidate that works with
// the picks has its weight multiplied by this. One multiplication, never
// stacked, however many rules match. Rules:
//  1. a Trainer's text names a picked Pokémon;
//  2. a candidate's effect text names a type you have picked Pokémon of;
//  3. a candidate Pokémon's type is named by the text of one of your picks.
//
// A Colorless reference never counts.
const synergyMult = 3.0

// typeSymbols maps the energy symbol letters in card text to card types.
var typeSymbols = map[string]string{
	"G": "Grass", "R": "Fire", "W": "Water", "L": "Lightning", "P": "Psychic",
	"F": "Fighting", "D": "Darkness", "M": "Metal", "N": "Dragon",
}

// synergy is what the picks make worth boosting. The zero value boosts nothing.
type synergy struct {
	pickedTypes map[string]bool // types of picked Pokémon (not Colorless)
	textTypes   map[string]bool // types named in the text of any pick
	nameText    *regexp.Regexp  // a picked Pokémon's base name inside Trainer text
}

// effectText is the text that can name a type: Trainer text, ability and
// attack effects. Attack costs are not text.
func effectText(c draftCard) string {
	parts := []string{c.TrainerText}
	if c.Ability != nil {
		parts = append(parts, c.Ability.Effect)
	}
	for _, a := range c.Attacks {
		parts = append(parts, a.Effect)
	}
	return strings.Join(parts, "\n")
}

// typesNamedIn returns the non-Colorless types the text names: "<Type> Pokémon",
// "<Type> Energy", or the symbol in [W] or {W} form.
func typesNamedIn(text string) map[string]bool {
	out := map[string]bool{}
	for letter, typ := range typeSymbols {
		if strings.Contains(text, "["+letter+"]") || strings.Contains(text, "{"+letter+"}") ||
			strings.Contains(text, typ+" Pokémon") || strings.Contains(text, typ+" Energy") {
			out[typ] = true
		}
	}
	return out
}

// newSynergy mirrors weights.ts's buildSynergy.
func newSynergy(picks []draftCard) synergy {
	s := synergy{pickedTypes: map[string]bool{}, textTypes: map[string]bool{}}
	needles := []string{}
	for _, p := range picks {
		for typ := range typesNamedIn(effectText(p)) {
			s.textTypes[typ] = true
		}
		if p.CardType == "Trainer" {
			continue
		}
		if p.CardType != "Colorless" {
			s.pickedTypes[p.CardType] = true
		}
		needles = append(needles, regexp.QuoteMeta(strings.TrimSuffix(p.Name, " ex")))
	}
	if len(needles) > 0 {
		s.nameText = regexp.MustCompile(`(^|[^\p{L}])(` + strings.Join(needles, "|") + `)($|[^\p{L}])`)
	}
	return s
}

// matches reports whether c works with the picks.
func (s synergy) matches(c draftCard) bool {
	if c.CardType == "Trainer" && s.nameText != nil && s.nameText.MatchString(c.TrainerText) {
		return true
	}
	if c.CardType != "Trainer" && s.textTypes[c.CardType] {
		return true
	}
	for typ := range typesNamedIn(effectText(c)) {
		if s.pickedTypes[typ] {
			return true
		}
	}
	return false
}
