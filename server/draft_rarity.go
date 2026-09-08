package main

// Go port of src/data/rarity.ts, reused server-side to draft the bot's deck
// (issue #9: give the bot a real 16-round draft instead of a fixed preset).
// Keep this in lockstep with the TS source — it is the rarity/tier
// classification the client's draft algorithm is built on.

const (
	rarityDot1 = "◊"
	rarityDot2 = "◊◊"
	rarityDot3 = "◊◊◊"
	rarityDot4 = "◊◊◊◊"
)

// isCosmeticVariant mirrors rarity.ts's isCosmeticVariant: these rarities are
// cosmetic-only reprints, never draftable.
func isCosmeticVariant(rarity string) bool {
	switch rarity {
	case "☆", "☆☆", "☆☆☆", "♕":
		return true
	}
	return false
}

// inferDraftRarity mirrors rarity.ts's inferDraftRarity (Decision 3.A): Promo
// cards have no printed rarity tier, so their draft rarity is inferred from
// exKind. ok is false for a rarity this system doesn't draft.
func inferDraftRarity(rarity, exKind string) (draftRarity string, ok bool) {
	switch rarity {
	case rarityDot1, rarityDot2, rarityDot3, rarityDot4:
		return rarity, true
	case "Promo":
		if exKind == "regular" {
			return rarityDot1, true
		}
		return rarityDot4, true
	}
	return "", false
}
