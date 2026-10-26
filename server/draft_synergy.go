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
//  3. a candidate Pokémon's type is named by the text of one of your picks;
//  4. the other direction of rule 1 (#126): a candidate Pokémon is named in
//     the text of a Trainer you picked, or a Trainer and a Pokémon share a
//     group in draft_niche_support.go and you picked the other one.
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

	trainerTexts    []string     // text of picked Trainers, for the Pokémon they name (rule 4)
	nicheOfTrainers map[int]bool // niche groups of picked Trainers (rule 4)
	nicheOfPokemon  map[int]bool // niche groups of picked Pokémon (rule 4)
}

// pokemonBaseName is a Pokémon's name without a trailing " ex", the form a Trainer's
// text uses: "Snorlax ex" is named by text saying "Snorlax". The match is on
// the whole name, so "Snorlax" is not named by text saying only "Snorlax ex".
func pokemonBaseName(name string) string { return strings.TrimSuffix(name, " ex") }

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
	s := synergy{
		pickedTypes: map[string]bool{}, textTypes: map[string]bool{},
		nicheOfTrainers: map[int]bool{}, nicheOfPokemon: map[int]bool{},
	}
	for _, p := range picks {
		for _, typ := range namedTypes(p) {
			s.textTypes[typ] = true
		}
		if p.CardType == "Trainer" {
			s.trainerTexts = append(s.trainerTexts, p.TrainerText)
			for _, g := range nicheGroupsOf(p) {
				s.nicheOfTrainers[g] = true
			}
			continue
		}
		for _, g := range nicheGroupsOf(p) {
			s.nicheOfPokemon[g] = true
		}
		if p.CardType != "Colorless" {
			s.pickedTypes[p.CardType] = true
		}
		s.pickedNames = append(s.pickedNames, pokemonBaseName(p.Name))
	}
	return s
}

// matches reports whether c works with the picks.
func (s synergy) matches(c draftCard) bool {
	if c.CardType == "Trainer" {
		// A Trainer naming a picked type is type support, see weightOf.
		return s.namesPickedIn(c.TrainerText) || anyIn(s.nicheOfPokemon, nicheGroupsOf(c))
	}
	if s.textTypes[c.CardType] || s.namedByPickedTrainer(c) || anyIn(s.nicheOfTrainers, nicheGroupsOf(c)) {
		return true
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

// anyIn reports whether any of groups is in set.
func anyIn(set map[int]bool, groups []int) bool {
	for _, g := range groups {
		if set[g] {
			return true
		}
	}
	return false
}

// namedByPickedTrainer reports whether the text of a Trainer you picked names
// the Pokémon c as a whole word. A card named "X ex" is named by text saying
// "X" (as rule 1 does for picked names) or "X ex"; a card named "X" is not
// named by text saying "X ex", which is a different card.
func (s synergy) namedByPickedTrainer(c draftCard) bool {
	name := pokemonBaseName(c.Name)
	notFollowedBy := ""
	if name == c.Name {
		notFollowedBy = " ex"
	}
	for _, text := range s.trainerTexts {
		if containsWordNot(text, name, notFollowedBy) {
			return true
		}
	}
	return false
}

func containsWord(text, word string) bool { return containsWordNot(text, word, "") }

// containsWordNot is containsWord that skips an occurrence directly followed by
// notFollowedBy (when not empty), so "Snorlax" is not found in "Snorlax ex".
func containsWordNot(text, word, notFollowedBy string) bool {
	for from := 0; from <= len(text); {
		i := strings.Index(text[from:], word)
		if i < 0 {
			return false
		}
		start := from + i
		end := start + len(word)
		before, _ := utf8.DecodeLastRuneInString(text[:start])
		after, _ := utf8.DecodeRuneInString(text[end:])
		if (start == 0 || !unicode.IsLetter(before)) && (end == len(text) || !unicode.IsLetter(after)) &&
			(notFollowedBy == "" || !strings.HasPrefix(text[end:], notFollowedBy)) {
			return true
		}
		from = start + 1
	}
	return false
}
