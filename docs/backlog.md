# Backlog

Ideas captured to revisit later — not active work.

## Seasonal rotation for the free-universal set

The deckbuild rules currently grant two cards as "free universals" (Pokeball, Professor's Research). Idea raised 2026-05-24: rotate which Trainers fill that slot on a fixed cadence (monthly? matched to new-set drops?) so the meta shifts without being chased through individual card balance.

**Why now? It isn't.** Needs a calendar/persistence layer the backend doesn't have yet, and the right free-universal set should be informed by what playtest decks routinely brick on — premature before the deckbuild loop is even shippable.

**Tracked from:** rules.md "Whether additional free cards (X Speed, Potion, Red Card) are included is TBD through playtesting."

---

## Pai-Gow-style Pocket variant

After the core draft format ships, consider a Pocket adaptation of Magic's Pai-Gow format: each player drafts a pool and must form two distinct sub-decks with rules about how they relate (in MTG: the "back hand" must beat the "front hand" in evaluation). Becomes a deck-building puzzle on top of the draft.

**When:** after Phase 1 (draft + shop + deckbuild + run) is playable end-to-end and the trade-token system is tuned.

**Open questions:** what does "stronger / weaker" mean for a 20-card Pocket deck? Win-rate proxy via the simulator? Card-quality heuristic? Whether to do Conquest (already on the roadmap) first since it's simpler.

---

## Publish a `pocket-draft-data` dataset

After Phase 1 ships, extract the draft-specific data transformations into a public GitHub repo for community use.

**What it would expose** (none of which the existing datasets do, because they're general-purpose):
- Normalized 5-tier rarity scale, mapped from Pocket's 9 raw rarity codes
- Functionally-deduped card pool (one entry per unique card, cosmetic variants stripped)
- Per-booster pack composition tables (e.g., A1-Charizard, A1-Mewtwo, A1-Pikachu) with tier-stratified card lists
- Optional: pull-rate distributions sourced from community testing

**Why now? It isn't.** All three transformations are required for Phase 1 anyway. After they exist as Go code + JSON output, extracting them into a separate repo is a few hours of cleanup. Make the publish/no-publish decision then, with the actual code in hand and a sense of whether anyone else is asking for it.

**Source data:** would consume `chase-manning/pokemon-tcg-pocket-cards` (or equivalent) as upstream raw cards; this repo's contribution is the derived layer on top.

**License if published:** MIT, matching upstream.
