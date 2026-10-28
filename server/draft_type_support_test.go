package main

import (
	"os"
	"regexp"
	"slices"
	"sort"
	"strings"
	"testing"
)

// Type-support Trainers: the card text decides (owner 2026-10-28, decision 2).
// A Trainer supports the types its text names plus the card types of the
// Pokémon its text names; the hand table is only a fallback for text that
// gives neither.

func sortedTypes(in []string) []string {
	out := append([]string{}, in...)
	sort.Strings(out)
	return out
}

func TestSupportTypesComeFromTheText(t *testing.T) {
	useRealCards(t)
	pool := toDraftablePool()
	for _, tc := range []struct {
		name string
		want []string
	}{
		// "Garchomp or Togekiss": Garchomp is Dragon, Garchomp ex Fighting, Togekiss Psychic.
		{"Cynthia", []string{"Dragon", "Fighting", "Psychic"}},
		// "Decidueye ex, Incineroar ex, or Primarina ex".
		{"Hau", []string{"Fire", "Grass", "Water"}},
		// "Snorlax, Heracross, and Staraptor": Colorless names add nothing.
		{"Barry", []string{"Grass"}},
		// "[R] Energy" plus "Alolan Marowak or Turtonator" (a Fire and a Dragon printing).
		{"Kiawe", []string{"Dragon", "Fire"}},
		{"Marlon", []string{"Psychic", "Water"}},
		// Names only Colorless Pokémon: no type.
		{"Whitney", nil},
	} {
		got := sortedTypes(supportTypes(poolCard(t, pool, tc.name)))
		if !slices.Equal(got, tc.want) && !(len(got) == 0 && len(tc.want) == 0) {
			t.Errorf("%s supports %v, want %v", tc.name, got, tc.want)
		}
	}

	// Text that names neither a type nor a Pokémon falls back to the table.
	drayden := trainerText("Drayden", "During this turn, 1 of your opponent's Pokémon is chosen 1 more time for the Draco Meteor attack used by your Pokémon.")
	if got := supportTypes(drayden); !slices.Equal(got, []string{"Dragon"}) {
		t.Errorf("Drayden supports %v, want [Dragon] (the table)", got)
	}
	// Squirt Bottle names [R] but works against it.
	if got := supportTypes(trainerText("Squirt Bottle", "Discard a [R] Energy from your opponent's Active Pokémon.")); len(got) != 0 {
		t.Errorf("Squirt Bottle supports %v, want none", got)
	}
}

// The table holds only what the text cannot say: every entry is a Trainer
// whose text names no type and no typed Pokémon.
func TestTypeSupportTableIsOnlyTheFallback(t *testing.T) {
	useRealCards(t)
	for _, rc := range rawByID {
		if rc.CardType != "Trainer" || typeSupportTable[rc.Name] == nil {
			continue
		}
		if got := supportTypesFromText(draftCard{rawCard: rc}); len(got) > 0 {
			t.Errorf("%s (%s): the text already gives %v; drop it from typeSupportTable", rc.Name, rc.ID, got)
		}
	}
}

// One name, one answer: the pool file keys support types by Trainer name.
func TestSupportTypesAreTheSameForEveryPrintingOfAName(t *testing.T) {
	useRealCards(t)
	seen := map[string][]string{}
	for _, c := range toDraftablePool() {
		if c.CardType != "Trainer" {
			continue
		}
		got := supportTypes(c)
		if prev, ok := seen[c.Name]; ok && !slices.Equal(prev, got) {
			t.Errorf("%s: %v and %v", c.Name, prev, got)
		}
		seen[c.Name] = got
	}
}

// src/draft/typeSupport.ts keeps the same fallback table and exclusions.
func TestTypeSupportTSTableMatchesGo(t *testing.T) {
	src, err := os.ReadFile("../src/draft/typeSupport.ts")
	if err != nil {
		t.Fatal(err)
	}
	body := string(src)
	table := body[strings.Index(body, "TYPE_SUPPORT_TABLE"):]
	table = table[:strings.Index(table, "};")]
	got := map[string][]string{}
	for _, m := range regexp.MustCompile(`(?m)^\s*(?:'([^']+)'|"([^"]+)"|(\w+)): \[([^\]]*)\]`).FindAllStringSubmatch(table, -1) {
		name := m[1] + m[2] + m[3]
		var types []string
		for _, q := range regexp.MustCompile(`'([^']+)'`).FindAllStringSubmatch(m[4], -1) {
			types = append(types, q[1])
		}
		got[name] = types
	}
	if len(got) != len(typeSupportTable) {
		t.Errorf("TS table has %d entries %v, Go %d %v", len(got), got, len(typeSupportTable), typeSupportTable)
	}
	for name, types := range typeSupportTable {
		if !slices.Equal(got[name], types) {
			t.Errorf("%s: TS %v, Go %v", name, got[name], types)
		}
	}
	excl := body[strings.Index(body, "NOT_TYPE_SUPPORT"):]
	for name := range notTypeSupport {
		if !strings.Contains(excl, "'"+name+"'") {
			t.Errorf("NOT_TYPE_SUPPORT in TS lacks %s", name)
		}
	}
}
