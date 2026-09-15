package main

import "testing"

func fakeCard(id string, t tier) draftCard {
	dr := rarityDot1
	isEx := false
	switch t {
	case tier1:
		dr, isEx = rarityDot4, true
	case tier2:
		dr = rarityDot4
	case tier3:
		dr = rarityDot3
	case tier4:
		dr = rarityDot2
	}
	return draftCard{rawCard: rawCard{ID: id, Name: id, CardType: "Fire"}, DraftRarity: dr, IsEx: isEx}
}

func TestGenerateOfferNoDuplicatesWithinOffer(t *testing.T) {
	// A tiny pool: if generateOffer didn't exclude already-used ids within one
	// offer, 5 slots drawn from 3 cards would collide.
	pool := []draftCard{fakeCard("a", tier5), fakeCard("b", tier5), fakeCard("c", tier5)}
	byTier := indexByTier(pool)
	rng := newDraftRng(1)

	offer, err := generateOffer(byTier, nil, rng.Float64)
	if err == nil {
		t.Fatalf("expected an error: only 3 cards can't fill 5 unique slots, got offer %+v", offer)
	}
}

func TestGenerateOfferFillsFromAnotherTierWhenOneIsEmpty(t *testing.T) {
	// Only tier-5 cards exist; every roll of tier 1-4 must fall through to tier 5.
	pool := make([]draftCard, 0, 5)
	for i := 0; i < 5; i++ {
		pool = append(pool, fakeCard(string(rune('a'+i)), tier5))
	}
	byTier := indexByTier(pool)
	rng := newDraftRng(42)

	offer, err := generateOffer(byTier, nil, rng.Float64)
	if err != nil {
		t.Fatalf("generateOffer() error = %v", err)
	}
	if len(offer) != slotsPerPack {
		t.Fatalf("len(offer) = %d, want %d", len(offer), slotsPerPack)
	}
	seen := map[string]bool{}
	for _, c := range offer {
		if seen[c.ID] {
			t.Errorf("duplicate id %s within one offer", c.ID)
		}
		seen[c.ID] = true
	}
}

func TestGenerateOfferDeterministicForSameRng(t *testing.T) {
	pool := make([]draftCard, 0, 30)
	for i := 0; i < 30; i++ {
		pool = append(pool, fakeCard(string(rune('a'+i)), tier5))
	}
	byTier := indexByTier(pool)

	offer1, err := generateOffer(byTier, nil, newDraftRng(7).Float64)
	if err != nil {
		t.Fatalf("generateOffer() error = %v", err)
	}
	offer2, err := generateOffer(byTier, nil, newDraftRng(7).Float64)
	if err != nil {
		t.Fatalf("generateOffer() error = %v", err)
	}
	for i := range offer1 {
		if offer1[i].ID != offer2[i].ID {
			t.Fatalf("same seed produced different offers: %+v vs %+v", offer1, offer2)
		}
	}
}
