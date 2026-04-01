# Draft Format Rules

## Overview

An Arena-style draft format for Pokemon TCG Pocket. Each session is self-contained — no pre-existing card collection required. Any card in the game is draftable on equal footing. Basic universals are freely available without spending picks.

Inspired by the **SWUCUBE** cube draft format (see `swucube-rules.md`) — originally published on YouTube by SWUCUBE — and Hearthstone/MTG Arena draft modes. The 1-of-5 pick mechanic mirrors the Wonder Pick feature in Pokemon TCG Pocket.

---

## Session Structure

```
1. Draft Phase       (20 rounds of pick 1-of-5)
2. Shop Phase        (spend flat tickets on staples + undrafted picks)
3. Deckbuilding      (finalize 20-card deck, slot in free universals, cut)
4. Run               (play games until win/loss threshold)
```

---

## 1. Draft Phase

### Pick structure
- Each round, the player is offered **5 cards** and selects **1**.
- This repeats **20 times**.
- Mirrors the 5-card pack structure and Wonder Pick mechanic from the actual game.

### Pack 1 — soft anchor
- Pack 1 is guaranteed to contain at least **1 EX or Mega EX Pokemon** among the 5 options.
- The player may pick it (establishing a build-around and energy direction) or pass on it (knowingly opting into a less directed draft).
- This is a soft anchor — nothing is forced.

### Type weighting (packs 2–20)
- After each pick, the player's drafted energy type(s) are tracked.
- Subsequent packs are **modestly weighted** toward the player's detected primary type(s). This is a nudge, not a lock — off-type cards still appear.
- **Colorless Pokemon are always eligible** regardless of detected direction and are weighted as compatible with any type profile. This ensures archetypes like Dragonite (dual energy) and colorless-featured attackers remain naturally draftable.

### Card pool
- All functionally distinct cards in Pokemon TCG Pocket are eligible.
- Cosmetic variants (full art, special illustration rare, immersive, shiny, promo) are excluded — only one functional version of each card exists in the pool.
- Rarity tiers (used for draft weighting, not pull rates):

| Tier | Contents |
|---|---|
| 1 (rarest) | EX / Mega EX Pokemon |
| 2 | ◆◆◆◆ non-EX, high-impact Trainers |
| 3 | ◆◆◆ |
| 4 | ◆◆ |
| 5 (most common) | ◆ |

---

## 2. Shop Phase

After the draft, players spend a flat allotment of **3 tickets** (starting value — adjust through playtesting) on the shop.

Each card costs 1 ticket. Tickets cannot be saved or carried over.

### Staple shelf
- A fixed list of ~10–12 essential Trainer cards always available (e.g. Cyrus, Sabrina, Repel, Copycat).
- Ensures no player is bricked without key Supporters due to bad draft luck.
- Same contents for every player every session.

### Undrafted shelf
- 1 card selected from each of the 20 packs (the 4 unchosen cards per pack), scored and filtered by:
  - Synergy with the player's drafted cards
  - Rarity weighting
- Results in up to **20 candidates** from the undrafted pool.
- Total shop pool: ~30–32 cards.

### Shop purpose
The shop is for **patching holes and minor upgrades only** — not redesigning the deck. The ticket cap enforces this. Target outcomes: ensuring access to at least 1–2 Supporters, filling a missing item slot, or adding a Pokemon that completes a line.

---

## 3. Deckbuilding

After the shop, players build their final **20-card deck** from:
- Their 20 drafted cards
- Any cards acquired from the shop

### Free universals
The following cards are freely available in any quantity during deckbuilding and do not require a draft pick or ticket:
- **Pokeball**
- **Professor's Research**

These count toward the 20-card deck limit and replace drafted/shopped cards as the player sees fit.

> Note: Whether additional free cards (X Speed, Potion, Red Card) are included is TBD through playtesting.

### Cut
If the player has more than 20 cards after shop purchases, they cut down to exactly 20 during this phase.

---

## 4. Run Structure

Players play games with their drafted deck until a win/loss threshold is reached.

- **Starting threshold (to be tested):** 7 wins or 3 losses
- Specific numbers subject to adjustment based on playtesting.

---

## Matchmaking

- Open matchmaking for v1.
- MMR-based skill matching is a future consideration.

---

## Future / Deferred

- **Conquest format:** Players build 3 distinct decks and must win at least 1 game with each (see `swucube-rules.md`). Deferred — potential tournament mode after Arena is implemented.
- **Additional free universals:** X Speed, Potion, Red Card — whether these are free TBD.
- **Ticket count:** Starting at 3, to be tuned through playtesting.
- **Win/loss threshold:** Starting at 7W/3L, to be tuned through playtesting.
