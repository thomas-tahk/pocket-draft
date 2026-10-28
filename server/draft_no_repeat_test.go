package main

import "testing"

// A pick is secured for the whole draft (owner, 2026-10-25): once picked, no
// later offer holds a card of that name, whatever its printing. Mirrors the
// e2e in e2e/draft-offer.spec.ts for the TS side.
func TestBotDraftNeverReoffersAPickedName(t *testing.T) {
	useRealCards(t)
	pool := toDraftablePool()
	byTier := indexByTier(pool)
	byName := nameIndex()
	for seed := uint64(1); seed <= 20; seed++ {
		rng := newDraftRng(seed)
		var drafted []draftCard
		picked := map[string]bool{}
		for round := 0; round < totalDraftPacks; round++ {
			offer, err := generateOffer(byTier, drafted, rng.Float64)
			if err != nil {
				t.Fatalf("seed %d round %d: %v", seed, round, err)
			}
			for _, c := range offer {
				if picked[c.Name] {
					t.Fatalf("seed %d round %d: offered %q (%s), already picked", seed, round, c.Name, c.ID)
				}
			}
			c := botPick(offer, drafted, byName)
			picked[c.Name] = true
			drafted = append(drafted, c)
		}
	}
}
