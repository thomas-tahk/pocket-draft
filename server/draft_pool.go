package main

import (
	"fmt"
	"sort"
	"strings"
)

// Go port of the draftable-pool construction in src/data/loader.ts (plus the
// dedup helpers in src/data/signature.ts). The bot's draft (issue #9) must draw
// from the *same* pool a human drafter sees — PROJECTS.md's locked decision for
// this project — so this is a faithful port, not a simplified stand-in.

// draftCard is the enriched card the draft algorithm operates on — the Go
// equivalent of the client's `Card` type (RawCard + draftRarity/isEx/isMegaEx).
type draftCard struct {
	rawCard
	DraftRarity string
	IsEx        bool
	IsMegaEx    bool
}

// excludedDraftSets mirrors loader.ts's EXCLUDED_SETS: A4b is reprint-only with
// glossy treatments; every base-art card has a functional duplicate elsewhere.
var excludedDraftSets = map[string]bool{"A4b": true}

// freeUniversalNames mirrors loader.ts's FREE_UNIVERSAL_NAMES: always available
// at deckbuild, never offered in the draft itself. Deckbuild/shop are deferred
// past this slice, so these never reach the bot's deck via draftBotDeck either.
var freeUniversalNames = map[string]bool{
	"Poké Ball":            true,
	"Professor's Research": true,
}

// isDraftCandidate mirrors loader.ts's isDraftCandidate.
func isDraftCandidate(rc rawCard) bool {
	if excludedDraftSets[rc.SetID] {
		return false
	}
	if isCosmeticVariant(rc.Rarity) {
		return false
	}
	if rc.Rarity == "" {
		return false
	}
	if rc.CardType == "Trainer" && freeUniversalNames[rc.Name] {
		return false
	}
	return true
}

// enrichForDraft mirrors rarity.ts's enrich.
func enrichForDraft(rc rawCard) (draftCard, bool) {
	dr, ok := inferDraftRarity(rc.Rarity, rc.ExKind)
	if !ok {
		return draftCard{}, false
	}
	return draftCard{
		rawCard:     rc,
		DraftRarity: dr,
		IsEx:        rc.ExKind == "ex" || rc.ExKind == "mega-ex",
		IsMegaEx:    rc.ExKind == "mega-ex",
	}, true
}

// functionalSignature mirrors signature.ts's functionalSignature: everything
// that matters for play, so cosmetic reprints across sets collapse together.
func functionalSignature(c draftCard) string {
	attacks := make([]string, len(c.Attacks))
	for i, a := range c.Attacks {
		attacks[i] = fmt.Sprintf("%s|%s|%s|%s", a.Cost, a.Name, a.Damage, a.Effect)
	}
	ability := ""
	if c.Ability != nil {
		ability = c.Ability.Name + "|" + c.Ability.Effect
	}
	return strings.Join([]string{
		c.CardType,
		fmt.Sprint(c.HP),
		c.Stage,
		c.EvolvesFrom,
		c.Weakness,
		fmt.Sprint(c.Retreat),
		strings.Join(attacks, "||"),
		ability,
		c.TrainerKind,
		c.TrainerText,
	}, "::")
}

// functionalDedupe mirrors signature.ts's functionalDedupe: collapse
// name+signature duplicates, keeping the lowest-id representative.
func functionalDedupe(cards []draftCard) []draftCard {
	byKey := map[string]draftCard{}
	order := make([]string, 0, len(cards))
	for _, c := range cards {
		key := c.Name + "::" + functionalSignature(c)
		if existing, ok := byKey[key]; !ok {
			byKey[key] = c
			order = append(order, key)
		} else if c.ID < existing.ID {
			byKey[key] = c
		}
	}
	out := make([]draftCard, 0, len(order))
	for _, k := range order {
		out = append(out, byKey[k])
	}
	return out
}

// dedupePromoRayquaza mirrors loader.ts's dedupePromoRayquaza (Decision 4.A):
// promo reprints of the same card (e.g. Rayquaza ex) collapse to the lowest id.
func dedupePromoRayquaza(cards []draftCard) []draftCard {
	byKey := map[string]draftCard{}
	order := make([]string, 0, len(cards))
	for _, c := range cards {
		if !strings.HasPrefix(c.SetID, "P-") {
			byKey[c.ID] = c
			order = append(order, c.ID)
			continue
		}
		key := fmt.Sprintf("%s|%s|%d|%s", c.Name, c.CardType, c.HP, c.ExKind)
		if existing, ok := byKey[key]; !ok {
			byKey[key] = c
			order = append(order, key)
		} else if c.ID < existing.ID {
			byKey[key] = c
		}
	}
	out := make([]draftCard, 0, len(order))
	for _, k := range order {
		out = append(out, byKey[k])
	}
	return out
}

// filterToFinalEvolutions mirrors loader.ts's filterToFinalEvolutions: the
// pool-then-pick model only drafts "final" Pokémon (nothing evolves further);
// under-evolutions are granted free at deckbuild, which this slice defers.
func filterToFinalEvolutions(cards []draftCard) []draftCard {
	hasSuccessor := map[string]bool{}
	for _, c := range cards {
		if c.EvolvesFrom != "" {
			hasSuccessor[c.EvolvesFrom] = true
		}
	}
	out := make([]draftCard, 0, len(cards))
	for _, c := range cards {
		if c.CardType == "Trainer" || !hasSuccessor[c.Name] {
			out = append(out, c)
		}
	}
	return out
}

// toDraftablePool mirrors loader.ts's toDraftablePool over the loaded card
// data: the same filter → enrich → dedupe → final-evolutions-only pipeline the
// client's draft pool goes through, sorted by id for a deterministic order
// (rawByID is a map; iteration order must not leak into a seeded draft).
func toDraftablePool() []draftCard {
	enriched := make([]draftCard, 0, len(rawByID))
	for _, rc := range rawByID {
		if !isDraftCandidate(rc) {
			continue
		}
		if dc, ok := enrichForDraft(rc); ok {
			enriched = append(enriched, dc)
		}
	}
	sort.Slice(enriched, func(i, j int) bool { return enriched[i].ID < enriched[j].ID })

	deduped := dedupePromoRayquaza(enriched)
	finals := filterToFinalEvolutions(deduped)
	pool := functionalDedupe(finals)
	sort.Slice(pool, func(i, j int) bool { return pool[i].ID < pool[j].ID })
	return pool
}
