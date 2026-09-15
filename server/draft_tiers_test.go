package main

import "testing"

func TestTierOf(t *testing.T) {
	cases := []struct {
		name string
		c    draftCard
		want tier
	}{
		{"ex at ◊◊◊◊ is tier 1", draftCard{DraftRarity: rarityDot4, IsEx: true}, tier1},
		{"non-ex at ◊◊◊◊ is tier 2", draftCard{DraftRarity: rarityDot4, IsEx: false}, tier2},
		{"◊◊◊ is tier 3", draftCard{DraftRarity: rarityDot3}, tier3},
		{"◊◊ is tier 4", draftCard{DraftRarity: rarityDot2}, tier4},
		{"◊ is tier 5", draftCard{DraftRarity: rarityDot1}, tier5},
	}
	for _, c := range cases {
		if got := tierOf(c.c); got != c.want {
			t.Errorf("%s: tierOf() = %d, want %d", c.name, got, c.want)
		}
	}
}

func TestIndexByTier(t *testing.T) {
	pool := []draftCard{
		{DraftRarity: rarityDot4, IsEx: true, rawCard: rawCard{ID: "a"}},
		{DraftRarity: rarityDot1, rawCard: rawCard{ID: "b"}},
		{DraftRarity: rarityDot1, rawCard: rawCard{ID: "c"}},
	}
	idx := indexByTier(pool)
	if len(idx[tier1]) != 1 || idx[tier1][0].ID != "a" {
		t.Errorf("tier1 = %+v, want [a]", idx[tier1])
	}
	if len(idx[tier5]) != 2 {
		t.Errorf("tier5 = %+v, want 2 entries", idx[tier5])
	}
	if len(idx[tier2]) != 0 || len(idx[tier3]) != 0 || len(idx[tier4]) != 0 {
		t.Errorf("expected empty tiers 2-4, got %+v", idx)
	}
}
