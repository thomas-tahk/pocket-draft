# Draft Format Rules

## Overview

An Arena-style draft format for Pokemon TCG Pocket. Each session is self-contained — no pre-existing card collection required. Any card in the game is draftable on equal footing. There are no free Trainers: every card comes from a pick or the shop (lower evolutions of drafted Pokémon excepted).

Inspired by the **SWUCUBE** cube draft format (see `swucube-rules.md`) — originally published on YouTube by SWUCUBE — and Hearthstone/MTG Arena draft modes. The 1-of-5 pick mechanic mirrors the Wonder Pick feature in Pokemon TCG Pocket.

---

## Session Structure

```
1. Draft Phase       (20 rounds of pick 1-of-5)
2. Shop Phase        (spend flat tickets on staples + same-type ex cards)
3. Deckbuilding      (finalize 20-card deck, cut)
4. Run               (play games until win/loss threshold)
```

---

## 1. Draft Phase

### Pick structure
- Each round, the player is offered **5 cards** and selects **1**.
- This repeats **20 times**.
- Mirrors the 5-card pack structure and Wonder Pick mechanic from the actual game.
- The draft is **solitary and asynchronous** — each player drafts independently with no timer and no dependency on any other player. Players may leave and return mid-draft at any point. Matchmaking happens after both players have completed their drafts.

### Pack 1 — soft anchor
- Pack 1 is guaranteed to contain at least **1 EX or Mega EX Pokemon** among the 5 options.
- The player may pick it (establishing a build-around and energy direction) or pass on it (knowingly opting into a less directed draft).
- This is a soft anchor — nothing is forced.

### Type weighting (packs 2–20)
- After each pick, the player's drafted energy type(s) are tracked.
- Subsequent packs lean toward the player's types, visibly after 2–3 picks of one type (owner, 2026-10-25). A type's Pokémon and the Trainers/Items that support that type both get the lean. Raised odds, never a lock: off-type cards still appear.
- **Colorless Pokemon are always eligible** regardless of detected direction and are weighted as compatible with any type profile. This ensures archetypes like Dragonite (dual energy) and colorless-featured attackers remain naturally draftable.

### Offers that fit (owner, 2026-10-25)
- A niche Trainer and the Pokémon it supports raise each other's odds, whichever is picked first. Archetype support cards rise with their archetype. Raised odds, never guarantees.
- Eevee ex is draftable, and picking any Eeveelution raises its odds.

### A pick is secured
- Once a card is picked, it never appears in a later offer (matched by name). One pick already allows up to 2 copies at deckbuild.

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
- A fixed list of 12 essential Trainer cards (`src/shop/staples.ts`: Cyrus, Sabrina, Repel, Copycat, Giovanni, Pokémon Communication, Erika, Mars, Iono, Cynthia, Pokémon Center Lady, Lillie).
- Ensures no player is bricked without key Supporters due to bad draft luck.
- Same contents for every player every session, minus any you already drafted.

### Same-type ex shelf
- Every ex / Mega ex Pokémon of a type you drafted Pokémon of (Colorless doesn't count), one printing per name.
- Names you already drafted are left out: you already hold up to 2 copies of them.
- Pokémon-specific Trainers (e.g. Blaine for Ninetales / Rapidash / Magmar) are not on the shelf yet.

Source of truth: `src/shop/inventory.ts`. (This replaces the earlier "undrafted shelf" idea of one card from each pack's unchosen four.)

### Shop purpose
The shop is for **patching holes and minor upgrades only** — not redesigning the deck. The ticket cap enforces this. Target outcomes: ensuring access to at least 1–2 Supporters, filling a missing item slot, or adding a Pokemon that completes a line.

---

## 3. Deckbuilding

After the shop, players build their final **20-card deck** from:
- Their 20 drafted cards
- Any cards acquired from the shop

### Copy limits
Pokemon TCG Pocket's standard deck-construction rule applies: **maximum 2 copies of any card**, matched by card name. There is no separate restriction on EX or Mega-EX cards — they follow the same 2-copy cap as everything else. Cosmetic art variants of the same card do not stack beyond 2 (they're filtered out of the draft pool anyway).

### Energy Zone (owner, 2026-10-25)
The player chooses the deck's Energy Zone types (1 to 3) at deckbuild. The default is the energy the deck's attacks need (most-needed first), the same rule the engine uses when none is chosen. The choice holds for every battle of the run.

### No free Trainers
There are no free universal Trainers. Trainers (Pokeball and Professor's Research included) come only from draft picks and the shop. Lower evolutions of a drafted Pokemon stay free at deckbuild. A ruleset with a few open cards may come later (`docs/backlog.md`).

### Cut
If the player has more than 20 cards after shop purchases, they cut down to exactly 20 during this phase.

---

## 4. Run Structure

Players play games with their drafted deck until a win/loss threshold is reached.

- **Threshold:** 5 wins or 2 losses (owner, 2026-10-24)
- Specific numbers subject to adjustment based on playtesting.

---

## Matchmaking

- Open matchmaking for v1: queue against a real player; after 2 minutes with no match, fall back to a bot playing another player's drafted deck, clearly marked as a bot (owner, 2026-10-24).
- MMR-based skill matching is a future consideration.

---

## Future / Deferred

- **Conquest format:** Players build 3 distinct decks and must win at least 1 game with each (see `swucube-rules.md`). Deferred — potential tournament mode after Arena is implemented.
- **Ticket count:** Starting at 4, to be tuned through playtesting.
- **Win/loss threshold:** 5W/2L, to be tuned through playtesting.
