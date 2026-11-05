package main

import "fmt"

// Go port of src/draft/offer.ts.

const slotsPerPack = 5

// pickWeighted mirrors offer.ts's pickWeighted. rng must return a value in
// [0, 1), same contract as the TS side's Rng (default Math.random).
func pickWeighted[T any](items []T, weights []float64, rng func() float64) T {
	total := 0.0
	for _, w := range weights {
		total += w
	}
	r := rng() * total
	for i, w := range weights {
		r -= w
		if r <= 0 {
			return items[i]
		}
	}
	return items[len(items)-1]
}

// rollTier mirrors offer.ts's rollTier.
func rollTier(rng func() float64) tier {
	weights := make([]float64, len(tierOrder))
	for i, t := range tierOrder {
		weights[i] = slotTierWeights[t]
	}
	return pickWeighted(tierOrder, weights, rng)
}

// pickFromTier mirrors offer.ts's pickFromTier. ok is false when every card in
// the tier is already excluded (used up in this offer, or its exact name already
// picked: a pick is secured for the whole draft, so other printings of it are
// out too, while "Snorlax" and "Snorlax ex" stay distinct names).
func pickFromTier(pool []draftCard, tally typeTally, syn synergy, exclude, pickedNames map[string]bool, rng func() float64) (draftCard, bool) {
	eligible := make([]draftCard, 0, len(pool))
	for _, c := range pool {
		if !exclude[c.ID] && !pickedNames[c.Name] {
			eligible = append(eligible, c)
		}
	}
	if len(eligible) == 0 {
		return draftCard{}, false
	}
	weights := make([]float64, len(eligible))
	for i, c := range eligible {
		weights[i] = weightOf(c, tally, syn)
	}
	return pickWeighted(eligible, weights, rng), true
}

// generateOffer mirrors offer.ts's generateOffer: one 5-card offer. Within an
// offer, no duplicate ids; and no card whose name is already in picks (a pick is
// secured for the whole draft; a tier emptied by that falls through to the
// other tiers like any empty tier). No picks yet
// means pack 1, whose last slot is always tier 1 (ex / mega-ex); if tier 1 has
// nothing eligible the usual fall-through applies.
func generateOffer(byTier map[tier][]draftCard, picks []draftCard, rng func() float64) ([]draftCard, error) {
	tally, syn := tallyTypes(picks), newSynergy(picks)
	isPack1 := len(picks) == 0
	offer := make([]draftCard, 0, slotsPerPack)
	used := map[string]bool{}
	pickedNames := map[string]bool{}
	for _, p := range picks {
		pickedNames[p.Name] = true
	}

	for slot := 0; slot < slotsPerPack; slot++ {
		targetTier := rollTier(rng)
		if isPack1 && slot == slotsPerPack-1 {
			targetTier = tier1 // pack 1 always offers an ex
		}

		card, ok := pickFromTier(byTier[targetTier], tally, syn, used, pickedNames, rng)
		if !ok {
			// Tier ran out (e.g. tier 1 has no eligible cards left). Step through
			// other tiers.
			for _, t := range tierOrder {
				if t == targetTier {
					continue
				}
				card, ok = pickFromTier(byTier[t], tally, syn, used, pickedNames, rng)
				if ok {
					break
				}
			}
		}

		if !ok {
			return nil, fmt.Errorf("card pool exhausted while generating offer")
		}
		offer = append(offer, card)
		used[card.ID] = true
	}

	return offer, nil
}
