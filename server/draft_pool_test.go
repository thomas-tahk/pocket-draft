package main

import (
	"reflect"
	"testing"
)

func TestIsDraftCandidate(t *testing.T) {
	cases := []struct {
		name string
		rc   rawCard
		want bool
	}{
		{"ordinary Basic", rawCard{SetID: "A1", Rarity: "◊"}, true},
		{"excluded set A4b", rawCard{SetID: "A4b", Rarity: "◊"}, false},
		{"cosmetic rarity", rawCard{SetID: "A1", Rarity: "☆"}, false},
		{"empty rarity", rawCard{SetID: "A1", Rarity: ""}, false},
		{"free universal Trainer", rawCard{SetID: "A1", Rarity: "◊", CardType: "Trainer", Name: "Poké Ball"}, false},
		{"non-universal Trainer", rawCard{SetID: "A1", Rarity: "◊", CardType: "Trainer", Name: "Potion"}, true},
	}
	for _, c := range cases {
		if got := isDraftCandidate(c.rc); got != c.want {
			t.Errorf("%s: isDraftCandidate() = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestFilterToFinalEvolutions(t *testing.T) {
	// Charmander -> Charmeleon -> Charizard: only Charizard (nothing evolves
	// from it) should survive. A standalone Basic with no evolution line and a
	// Trainer both pass through untouched.
	cards := []draftCard{
		{rawCard: rawCard{ID: "1", Name: "Charmander"}},
		{rawCard: rawCard{ID: "2", Name: "Charmeleon", EvolvesFrom: "Charmander"}},
		{rawCard: rawCard{ID: "3", Name: "Charizard", EvolvesFrom: "Charmeleon"}},
		{rawCard: rawCard{ID: "4", Name: "Standalone Basic"}},
		{rawCard: rawCard{ID: "5", Name: "Potion", CardType: "Trainer"}},
	}
	got := filterToFinalEvolutions(cards)
	var gotIDs []string
	for _, c := range got {
		gotIDs = append(gotIDs, c.ID)
	}
	want := []string{"3", "4", "5"}
	if !reflect.DeepEqual(gotIDs, want) {
		t.Errorf("filterToFinalEvolutions() ids = %v, want %v", gotIDs, want)
	}
}

func TestFunctionalDedupe(t *testing.T) {
	// Two reprints of "Eevee" with the identical signature collapse to the
	// lowest id; a mechanically distinct third Eevee survives as its own entry.
	sig := rawAttack{Cost: "C", Name: "Tackle", Damage: "10"}
	cards := []draftCard{
		{rawCard: rawCard{ID: "B2", Name: "Eevee", Attacks: []rawAttack{sig}}},
		{rawCard: rawCard{ID: "A1", Name: "Eevee", Attacks: []rawAttack{sig}}},
		{rawCard: rawCard{ID: "C3", Name: "Eevee", Attacks: []rawAttack{{Cost: "C", Name: "Growl", Damage: "0"}}}},
	}
	got := functionalDedupe(cards)
	if len(got) != 2 {
		t.Fatalf("functionalDedupe() = %d entries, want 2: %+v", len(got), got)
	}
	ids := map[string]bool{}
	for _, c := range got {
		ids[c.ID] = true
	}
	if !ids["A1"] || !ids["C3"] {
		t.Errorf("functionalDedupe() ids = %v, want A1 (lowest of the Tackle dupes) and C3", ids)
	}
}

func TestDedupePromoRayquaza(t *testing.T) {
	cards := []draftCard{
		{rawCard: rawCard{ID: "P-A-065", SetID: "P-A", Name: "Rayquaza ex", CardType: "Dragon", HP: 180, ExKind: "ex"}},
		{rawCard: rawCard{ID: "P-A-064", SetID: "P-A", Name: "Rayquaza ex", CardType: "Dragon", HP: 180, ExKind: "ex"}},
		{rawCard: rawCard{ID: "A1-100", SetID: "A1", Name: "Some Other Card"}},
	}
	got := dedupePromoRayquaza(cards)
	if len(got) != 2 {
		t.Fatalf("dedupePromoRayquaza() = %d entries, want 2: %+v", len(got), got)
	}
	for _, c := range got {
		if c.Name == "Rayquaza ex" && c.ID != "P-A-064" {
			t.Errorf("dedupePromoRayquaza() kept %s, want lowest id P-A-064", c.ID)
		}
	}
}

func TestToDraftablePool(t *testing.T) {
	saved := rawByID
	defer func() { rawByID = saved }()

	rawByID = map[string]rawCard{
		"keep1":     {ID: "keep1", SetID: "A1", Name: "Keeper", Rarity: "◊", CardType: "Fire", Stage: "Basic"},
		"excl-set":  {ID: "excl-set", SetID: "A4b", Name: "Excluded Set", Rarity: "◊", CardType: "Fire"},
		"cosmetic":  {ID: "cosmetic", SetID: "A1", Name: "Shiny", Rarity: "☆", CardType: "Fire"},
		"under-evo": {ID: "under-evo", SetID: "A1", Name: "Baby", Rarity: "◊", CardType: "Fire", Stage: "Basic"},
		"evolved":   {ID: "evolved", SetID: "A1", Name: "Grown Up", Rarity: "◊◊", CardType: "Fire", Stage: "Stage 1", EvolvesFrom: "Baby"},
		"universal": {ID: "universal", SetID: "A1", Name: "Poké Ball", Rarity: "◊", CardType: "Trainer"},
	}

	pool := toDraftablePool()
	ids := map[string]bool{}
	for _, c := range pool {
		ids[c.ID] = true
	}

	for _, want := range []string{"keep1", "evolved"} {
		if !ids[want] {
			t.Errorf("toDraftablePool() missing %s, got %+v", want, ids)
		}
	}
	for _, unwanted := range []string{"excl-set", "cosmetic", "under-evo", "universal"} {
		if ids[unwanted] {
			t.Errorf("toDraftablePool() unexpectedly kept %s", unwanted)
		}
	}
}
