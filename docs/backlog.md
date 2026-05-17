# Backlog

Ideas captured to revisit later — not active work.

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
