package main

import "testing"

func TestIsCosmeticVariant(t *testing.T) {
	cosmetic := []string{"☆", "☆☆", "☆☆☆", "♕"}
	for _, r := range cosmetic {
		if !isCosmeticVariant(r) {
			t.Errorf("isCosmeticVariant(%q) = false, want true", r)
		}
	}
	notCosmetic := []string{"◊", "◊◊", "◊◊◊", "◊◊◊◊", "Promo", ""}
	for _, r := range notCosmetic {
		if isCosmeticVariant(r) {
			t.Errorf("isCosmeticVariant(%q) = true, want false", r)
		}
	}
}

func TestInferDraftRarity(t *testing.T) {
	cases := []struct {
		rarity, exKind string
		want           string
		ok             bool
	}{
		{"◊", "regular", "◊", true},
		{"◊◊◊◊", "ex", "◊◊◊◊", true},
		// Promo has no printed rarity tier — inferred from exKind (Decision 3.A).
		{"Promo", "regular", "◊", true},
		{"Promo", "ex", "◊◊◊◊", true},
		{"Promo", "mega-ex", "◊◊◊◊", true},
		{"☆", "regular", "", false},
		{"", "regular", "", false},
	}
	for _, c := range cases {
		got, ok := inferDraftRarity(c.rarity, c.exKind)
		if got != c.want || ok != c.ok {
			t.Errorf("inferDraftRarity(%q, %q) = (%q, %v), want (%q, %v)", c.rarity, c.exKind, got, ok, c.want, c.ok)
		}
	}
}
