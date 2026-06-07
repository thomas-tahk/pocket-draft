# Draft Format Rules

## Overview

An Arena-style draft format for Pokemon TCG Pocket. Each session is self-contained — no pre-existing card collection required. Any card in the game is draftable on equal footing. Basic universals are freely available without spending picks.

Inspired by the **SWUCUBE** cube draft format (see `swucube-rules.md`) — originally published on YouTube by SWUCUBE — and Hearthstone/MTG Arena draft modes. The 1-of-5 pick mechanic mirrors the Wonder Pick feature in Pokemon TCG Pocket.

---

## Session Structure

```
1. Draft Phase       (16 rounds of pick 1-of-5)
2. Shop Phase        (spend flat tickets on staples + undrafted picks)
3. Deckbuilding      (finalize 20-card deck, slot in free universals, cut)
4. Run               (play games until win/loss threshold)
```

---

## 1. Draft Phase

### Pick structure
- Each round, the player is offered **5 cards** and selects **1**.
- This repeats **16 times** (4×4 — homage to cube-draft sensibility).
- Mirrors the 5-card pack structure and Wonder Pick mechanic from the actual game.
- The draft is **solitary and asynchronous** — each player drafts independently with no timer and no dependency on any other player. Players may leave and return mid-draft at any point. Matchmaking happens after both players have completed their drafts.

### Pack 1 — soft anchor
- Pack 1 is guaranteed to contain at least **1 EX or Mega EX Pokemon** among the 5 options.
- The player may pick it (establishing a build-around and energy direction) or pass on it (knowingly opting into a less directed draft).
- This is a soft anchor — nothing is forced.

### Type weighting (packs 2–16)
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

After the draft, players spend a flat allotment of **4 tickets** (starting value — adjust through playtesting) on the shop.

Each ticket buys **one deck copy** of a chosen card. Spend two tickets on the same card to lock in 2 deck slots, or spread them across up to 4 different cards. Per-card cap is the global 2-copy-by-name limit (max 2 tickets on any one card). Tickets can't be saved or carried over.

### Staple shelf
- A fixed list of ~10–12 essential Trainer cards always available (e.g. Cyrus, Sabrina, Repel, Copycat).
- Ensures no player is bricked without key Supporters due to bad draft luck.
- Same contents for every player every session.

### Undrafted shelf
- 1 card selected from each of the 16 packs (the 4 unchosen cards per pack), scored and filtered by:
  - Synergy with the player's drafted cards
  - Rarity weighting
- Results in up to **16 candidates** from the undrafted pool.
- Total shop pool: ~26–28 cards.

### Shop purpose
The shop is for **patching holes and minor upgrades only** — not redesigning the deck. The ticket cap enforces this. Target outcomes: ensuring access to at least 1–2 Supporters, filling a missing item slot, or adding a Pokemon that completes a line.

---

## 3. Deckbuilding

After the shop, players build their final **20-card deck** from:
- Their 16 drafted cards
- Any cards acquired from the shop

### Copy limits
Pokemon TCG Pocket's standard deck-construction rule applies: **maximum 2 copies of any card**, matched by card name. There is no separate restriction on EX or Mega-EX cards — they follow the same 2-copy cap as everything else. Cosmetic art variants of the same card do not stack beyond 2 (they're filtered out of the draft pool anyway).

### Free universals
The following cards are freely available in any quantity during deckbuilding and do not require a draft pick or ticket (subject to the 2-copy cap above):
- **Pokeball**
- **Professor's Research**

These count toward the 20-card deck limit and replace drafted/shopped cards as the player sees fit.

> Note: Whether additional free cards (X Speed, Potion, Red Card) are included is TBD through playtesting. A seasonal-rotation model for the free-universal set is captured in `docs/backlog.md`.

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
- **Ticket count:** Starting at 4, to be tuned through playtesting.
- **Win/loss threshold:** Starting at 7W/3L, to be tuned through playtesting.
