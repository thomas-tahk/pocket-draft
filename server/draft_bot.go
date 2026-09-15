package main

import (
	"fmt"
	"math/rand/v2"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// totalDraftPacks mirrors draftStore.ts's TOTAL_PACKS; draftDeckSize mirrors
// its DECK_SIZE (also the fixed 20 deckFromIDs requires).
const (
	totalDraftPacks = 16
	draftDeckSize   = 20
)

// newDraftRng builds a seeded, reproducible source for the bot's draft — the
// same PCG construction engine/rng.go uses (ADR-0005), so a bot deck is
// reproducible per game seed. engine.rng is unexported, so this is a
// same-shaped sibling rather than a shared type.
func newDraftRng(seed uint64) *rand.Rand {
	return rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15))
}

// draftBotDeck runs the actual pick-1-of-5×16 draft over the same draftable
// pool and tier-weighted offer algorithm the client's draft uses
// (src/stores/draftStore.ts, src/draft/offer.ts), auto-picking the first card
// of each offer in place of a human tap. It then tops the 16 picks up to a
// legal 20-card deck (the Shop Phase that normally does this is deferred past
// this slice — issue #9) and hands the result to deckFromIDs for the same
// legality/engine-modeling a human's finalized deck gets.
func draftBotDeck(seed uint64) ([]engine.Card, error) {
	pool := toDraftablePool()
	if len(pool) == 0 {
		return nil, fmt.Errorf("draft bot deck: empty draftable pool")
	}
	byTier := indexByTier(pool)
	rng := newDraftRng(seed)

	picks := make([]draftCard, 0, totalDraftPacks)
	for round := 0; round < totalDraftPacks; round++ {
		offer, err := generateOffer(byTier, picks, rng.Float64)
		if err != nil {
			return nil, fmt.Errorf("draft bot deck: round %d: %w", round, err)
		}
		picks = append(picks, offer[0])
	}

	return deckFromIDs(topUpToDeckSize(picks))
}

// topUpToDeckSize expands 16 draft picks into a legal 20-card id list by
// upping drafted picks to their 2-copy-by-name cap, rarest tier (tier 1) first,
// until the deck reaches size. Never exceeds the cap, even if the draft
// happened to pick the same card more than twice.
func topUpToDeckSize(picks []draftCard) []string {
	ids := make([]string, 0, draftDeckSize)
	countByName := map[string]int{}
	add := func(c draftCard) bool {
		if countByName[c.Name] >= 2 {
			return false
		}
		ids = append(ids, c.ID)
		countByName[c.Name]++
		return true
	}

	for _, c := range picks {
		add(c)
	}

	for len(ids) < draftDeckSize {
		addedThisPass := false
		for _, t := range tierOrder {
			for _, c := range picks {
				if tierOf(c) != t {
					continue
				}
				if add(c) {
					addedThisPass = true
				}
				if len(ids) == draftDeckSize {
					return ids
				}
			}
		}
		if !addedThisPass {
			break // every drafted name is already at the 2-copy cap
		}
	}
	return ids
}
