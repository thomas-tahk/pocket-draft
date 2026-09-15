package main

import (
	"fmt"
	"testing"
)

func TestTopUpToDeckSizeRespectsCapAndReachesSize(t *testing.T) {
	// 16 "picks": "Dupe" was drafted 3 times (across different packs, which the
	// real draft allows), everything else once. The cap must still hold even
	// though a raw pick count already exceeds it.
	picks := []draftCard{
		{rawCard: rawCard{ID: "dupe", Name: "Dupe"}, DraftRarity: rarityDot1},
		{rawCard: rawCard{ID: "dupe", Name: "Dupe"}, DraftRarity: rarityDot1},
		{rawCard: rawCard{ID: "dupe", Name: "Dupe"}, DraftRarity: rarityDot1},
	}
	for i := 0; i < 13; i++ {
		name := fmt.Sprintf("solo-%d", i)
		picks = append(picks, draftCard{rawCard: rawCard{ID: name, Name: name}, DraftRarity: rarityDot1})
	}

	ids := topUpToDeckSize(picks)
	if len(ids) != draftDeckSize {
		t.Fatalf("len(ids) = %d, want %d", len(ids), draftDeckSize)
	}
	count := map[string]int{}
	for _, id := range ids {
		count[id]++
	}
	for id, n := range count {
		if n > 2 {
			t.Errorf("id %s appears %d times, want at most 2", id, n)
		}
	}
	if count["dupe"] != 2 {
		t.Errorf("dupe count = %d, want 2 (capped from 3 raw picks)", count["dupe"])
	}
}

func TestTopUpToDeckSizePrefersRarestTierFirst(t *testing.T) {
	// One tier-1 pick and one tier-5 pick, nothing else: topping up to 20 has to
	// give both a 2nd copy either way, so use a size that only needs one more
	// slot to show tier order — tier 1 must be topped up before tier 5.
	picks := []draftCard{
		{rawCard: rawCard{ID: "rare", Name: "Rare"}, DraftRarity: rarityDot4, IsEx: true},
		{rawCard: rawCard{ID: "common", Name: "Common"}, DraftRarity: rarityDot1},
	}
	ids := topUpToDeckSize(picks)
	// Both names can only reach 2 copies each (cap), giving 4 total — nowhere
	// near 20 — but the first topped-up slot (index 2) must still be the rarer
	// pick, matching the issue's "in tier order" bridge.
	if len(ids) < 3 || ids[2] != "rare" {
		t.Fatalf("ids = %v, want index 2 to be the tier-1 top-up (rare)", ids)
	}
}

// syntheticDraftPool installs a small but varied rawByID so draftBotDeck can
// run a full 16-round draft without depending on the scraped public/data/cards.json.
func syntheticDraftPool(t *testing.T) {
	t.Helper()
	saved := rawByID
	t.Cleanup(func() { rawByID = saved })

	types := []string{"Fire", "Water", "Grass", "Lightning", "Psychic"}
	rarities := []struct {
		rarity, exKind string
	}{
		{"◊", "regular"}, {"◊◊", "regular"}, {"◊◊◊", "regular"}, {"◊◊◊◊", "ex"},
	}

	pool := map[string]rawCard{}
	for i := 0; i < 60; i++ {
		typ := types[i%len(types)]
		rar := rarities[i%len(rarities)]
		id := fmt.Sprintf("SYN-%03d", i)
		pool[id] = rawCard{
			ID: id, SetID: "A1", Name: fmt.Sprintf("Synth%03d", i),
			Rarity: rar.rarity, ExKind: rar.exKind, CardType: typ, Stage: "Basic", HP: 60,
			Attacks: []rawAttack{{Cost: typ[:1], Name: "Hit", Damage: "10"}},
		}
	}
	rawByID = pool
}

func TestDraftBotDeckReproducibleAndVariesBySeed(t *testing.T) {
	syntheticDraftPool(t)

	deckNames := func(seed uint64) []string {
		deck, err := draftBotDeck(seed)
		if err != nil {
			t.Fatalf("draftBotDeck(%d): %v", seed, err)
		}
		if len(deck) != draftDeckSize {
			t.Fatalf("draftBotDeck(%d): got %d cards, want %d", seed, len(deck), draftDeckSize)
		}
		names := make([]string, len(deck))
		for i, c := range deck {
			names[i] = c.Name
		}
		return names
	}

	seed1a := deckNames(1)
	seed1b := deckNames(1)
	seed2 := deckNames(2)

	if fmt.Sprint(seed1a) != fmt.Sprint(seed1b) {
		t.Errorf("same seed produced different decks:\n%v\n%v", seed1a, seed1b)
	}
	if fmt.Sprint(seed1a) == fmt.Sprint(seed2) {
		t.Errorf("different seeds produced the identical deck: %v", seed1a)
	}
}
