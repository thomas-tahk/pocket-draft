package main

import (
	"slices"
	"testing"
)

// onTypeShare is the fraction of non-Trainer, non-Colorless cards in offers
// that match typ, over 20 seeded runs of 10 offers each given picks.
func onTypeShare(byTier map[tier][]draftCard, picks []draftCard, typ string) float64 {
	onType, total := 0, 0
	for seed := uint64(1); seed <= 20; seed++ {
		rng := newDraftRng(seed)
		for i := 0; i < 10; i++ {
			offer, err := generateOffer(byTier, picks, rng.Float64)
			if err != nil {
				panic(err)
			}
			for _, c := range offer {
				if c.CardType == "Trainer" || c.CardType == "Colorless" {
					continue
				}
				total++
				if c.CardType == typ {
					onType++
				}
			}
		}
	}
	return float64(onType) / float64(total)
}

func TestPackOneAlwaysOffersAnEx(t *testing.T) {
	useRealCards(t)
	byTier := indexByTier(toDraftablePool())
	for seed := uint64(1); seed <= 20; seed++ {
		offer, err := generateOffer(byTier, nil, newDraftRng(seed).Float64)
		if err != nil {
			t.Fatalf("seed %d: %v", seed, err)
		}
		if !slices.ContainsFunc(offer, func(c draftCard) bool { return c.IsEx }) {
			t.Errorf("seed %d: pack 1 has no ex: %+v", seed, offer)
		}
	}
}

func TestLaterOffersFavourPickedTypes(t *testing.T) {
	useRealCards(t)
	pool := toDraftablePool()
	byTier := indexByTier(pool)
	var fire []draftCard
	for _, c := range pool {
		if c.CardType == "Fire" && len(fire) < 8 {
			fire = append(fire, c)
		}
	}
	// One Trainer pick: past pack 1, but nothing for the tally to lean on.
	neutral := []draftCard{{rawCard: rawCard{ID: "t", CardType: "Trainer"}}}

	weighted := onTypeShare(byTier, fire, "Fire")
	baseline := onTypeShare(byTier, neutral, "Fire")
	t.Logf("Fire share: weighted %.3f, baseline %.3f", weighted, baseline)
	if weighted < baseline*1.5 {
		t.Errorf("weighted Fire share %.3f is not measurably above baseline %.3f", weighted, baseline)
	}
}

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
