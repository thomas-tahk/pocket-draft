package main

// Go port of src/draft/tiers.ts.

type tier int

const (
	tier1 tier = 1 // ex / mega-ex Pokémon (the pack-1 anchor target)
	tier2 tier = 2 // ◊◊◊◊ non-ex Pokémon and Trainers
	tier3 tier = 3
	tier4 tier = 4
	tier5 tier = 5
)

var tierOrder = []tier{tier1, tier2, tier3, tier4, tier5}

// slotTierWeights mirrors tiers.ts's SLOT_TIER_WEIGHTS. Tier 2 is intentionally
// 0 — Pocket reserves ◊◊◊◊ exclusively for EX/Mega-EX, so there are no ◊◊◊◊
// non-ex Pokémon in the actual card pool.
var slotTierWeights = map[tier]float64{
	tier5: 0.5,
	tier4: 0.25,
	tier3: 0.17,
	tier2: 0,
	tier1: 0.08,
}

// tierOf mirrors tiers.ts's tierOf.
func tierOf(c draftCard) tier {
	switch c.DraftRarity {
	case rarityDot4:
		if c.IsEx {
			return tier1
		}
		return tier2
	case rarityDot3:
		return tier3
	case rarityDot2:
		return tier4
	default:
		return tier5
	}
}

// indexByTier mirrors tiers.ts's indexByTier.
func indexByTier(pool []draftCard) map[tier][]draftCard {
	idx := map[tier][]draftCard{tier1: {}, tier2: {}, tier3: {}, tier4: {}, tier5: {}}
	for _, c := range pool {
		t := tierOf(c)
		idx[t] = append(idx[t], c)
	}
	return idx
}
