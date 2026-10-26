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
	pickedNames []string        // names of picked Pokémon, matched in Trainer text by textNames

	trainerTexts    []string     // text of picked Trainers, for the Pokémon they name (rule 4)
	nicheOfTrainers map[int]bool // niche groups of picked Trainers (rule 4)
	nicheOfPokemon  map[int]bool // niche groups of picked Pokémon (rule 4)
	eeveelution     bool         // an Eeveelution is picked (draft_eevee.go)
}

// pokemonBaseName is a Pokémon's name without a trailing " ex", the form a
// Trainer's text uses; see textNames for how it is matched.
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
		s.eeveelution = s.eeveelution || isEeveelution(p)
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
		s.pickedNames = append(s.pickedNames, p.Name)
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

// namesPickedIn reports whether text names a picked Pokémon (see textNames).
func (s synergy) namesPickedIn(text string) bool {
	for _, name := range s.pickedNames {
		if textNames(text, name) {
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
// the Pokémon c (see textNames).
func (s synergy) namedByPickedTrainer(c draftCard) bool {
	for _, text := range s.trainerTexts {
		if textNames(text, c.Name) {
			return true
		}
	}
	return false
}

// longerNamePrefixes precede a Pokémon name to make a different, longer card
// name: text saying "Alolan Marowak" does not name "Marowak".
var longerNamePrefixes = []string{"Alolan ", "Galarian ", "Hisuian ", "Paldean ", "Team Rocket's "}

// textNames reports whether a Trainer's text names the Pokémon card called
// cardName, as a whole word, used the same way in both directions (rule 1 and
// rule 4). A card named "X ex" is named by text saying "X" or "X ex"; a card
// named "X" is not named by text saying "X ex", a different card. An
// occurrence does not count when it is part of a longer card name (preceded by
// a regional or Team Rocket's prefix) or when it only says what something
// evolves from ("evolves from Eevee" names no Eevee ex).
func textNames(text, cardName string) bool {
	word := pokemonBaseName(cardName)
	isEx := word != cardName
	for from := 0; from <= len(text); {
		i := strings.Index(text[from:], word)
		if i < 0 {
			return false
		}
		start := from + i
		end := start + len(word)
		from = start + 1
		before, _ := utf8.DecodeLastRuneInString(text[:start])
		after, _ := utf8.DecodeRuneInString(text[end:])
		if (start > 0 && unicode.IsLetter(before)) || (end < len(text) && unicode.IsLetter(after)) {
			continue
		}
		if !isEx && strings.HasPrefix(text[end:], " ex") {
			continue
		}
		head := text[:start]
		if strings.HasSuffix(head, "evolve from ") || strings.HasSuffix(head, "evolves from ") {
			continue
		}
		longer := false
		for _, p := range longerNamePrefixes {
			if strings.HasSuffix(head, p) && !strings.HasPrefix(word, p) {
				longer = true
			}
		}
		if !longer {
			return true
		}
	}
	return false
}
