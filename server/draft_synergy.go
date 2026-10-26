package main

import (
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

// Go port of the synergy half of src/draft/weights.ts.

// synergyMult mirrors weights.ts's SYNERGY_MULT: a candidate that works with
// the picks has its weight multiplied by this. One multiplication, never
// stacked, however many rules match. Rules:
//  1. a Trainer's text names a picked Pokémon;
//  2. a candidate Pokémon's effect text names a type you have picked Pokémon
//     of (a Trainer doing so is type support instead: weightOf's lean);
//  3. a candidate Pokémon's type is named by the text of one of your picks.
//
// A card's text naming its own type never counts (the type weight covers it).
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
	pickedNames []string        // base names (no " ex") of picked Pokémon, matched in Trainer text
	eeveelution bool            // an Eeveelution is picked (draft_eevee.go)
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

// namedTypesCache memoises namedTypes by card id; a card's text never changes.
var namedTypesCache sync.Map

// namedTypes is the types c's effect text names, minus c's own type: a card
// naming its own type is already covered by the type weight, so it never
// counts for rule 2 (as a candidate) or rule 3 (as a pick). Trainers have no
// type, so they are unaffected. Cached per card id.
func namedTypes(c draftCard) []string {
	if cached, ok := namedTypesCache.Load(c.ID); ok {
		return cached.([]string)
	}
	out := []string{}
	for typ := range typesNamedIn(effectText(c)) {
		if typ != c.CardType {
			out = append(out, typ)
		}
	}
	namedTypesCache.Store(c.ID, out)
	return out
}

// newSynergy mirrors weights.ts's buildSynergy.
func newSynergy(picks []draftCard) synergy {
	s := synergy{pickedTypes: map[string]bool{}, textTypes: map[string]bool{}}
	for _, p := range picks {
		s.eeveelution = s.eeveelution || isEeveelution(p)
		for _, typ := range namedTypes(p) {
			s.textTypes[typ] = true
		}
		if p.CardType == "Trainer" {
			continue
		}
		if p.CardType != "Colorless" {
			s.pickedTypes[p.CardType] = true
		}
		s.pickedNames = append(s.pickedNames, strings.TrimSuffix(p.Name, " ex"))
	}
	return s
}

// matches reports whether c works with the picks.
func (s synergy) matches(c draftCard) bool {
	if c.CardType == "Trainer" && s.namesPickedIn(c.TrainerText) {
		return true
	}
	if c.CardType != "Trainer" && s.textTypes[c.CardType] {
		return true
	}
	if c.CardType == "Trainer" {
		return false // a Trainer naming a picked type is type support, see weightOf
	}
	for _, typ := range namedTypes(c) {
		if s.pickedTypes[typ] {
			return true
		}
	}
	return false
}

// namesPickedIn reports whether text contains a picked name as a whole word:
// not preceded or followed by a letter. Every occurrence is checked.
func (s synergy) namesPickedIn(text string) bool {
	for _, name := range s.pickedNames {
		if containsWord(text, name) {
			return true
		}
	}
	return false
}

func containsWord(text, word string) bool {
	for from := 0; from <= len(text); {
		i := strings.Index(text[from:], word)
		if i < 0 {
			return false
		}
		start := from + i
		end := start + len(word)
		before, _ := utf8.DecodeLastRuneInString(text[:start])
		after, _ := utf8.DecodeRuneInString(text[end:])
		if (start == 0 || !unicode.IsLetter(before)) && (end == len(text) || !unicode.IsLetter(after)) {
			return true
		}
		from = start + 1
	}
	return false
}
