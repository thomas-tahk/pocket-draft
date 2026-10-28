# Draft depth: evidence for the next-milestone issues (a)–(e)

2026-10-28. Evidence for the owner's "Draft design: next milestone" list in
`docs/decisions/2026-10-28-m4-review-answers.md`. Measured with the real card
data (`public/data/cards.json`, `public/data/draft-pool.json`) on branch
`m4-draft-runs` at `4a5e88f`. **This is the code before the type-lean cap and the
text-decided type-support change that another worker is making now**; numbers in
(a) and (e) will move once that lands. Re-run on any commit with:

```sh
cd server && PD_METRICS=1 go test -run TestDraftDepthMetrics -v
```

The measurement is `server/draft_depth_metrics_test.go` (skipped without
`PD_METRICS=1`; about 1 minute). Every run is 500 drafts with fixed seeds
1…500, so the same code prints the same numbers.

## Which code drafts

- **Draft bot** (the opponent): `draftBotIDs(seed)` in `server/draft_bot.go`
  draws 20 offers of 5 with `generateOffer`, picks with `botPick`, and builds the
  deck with `botDeckIDs` (`server/draft_bot_deck.go`). No shop.
- **Auto-draft (test)** on Home: `HomeView` → `App.tsx` → `src/game/api.ts`
  `POST /api/autodeck` → `autoDeckFor` (`server/autodeck.go`, also wired in
  `server/wasm_main.go` for the live site) → the **same `draftBotIDs`** with the
  seed XOR `yourDraftSalt`. The store's `autoDraft` then jumps straight to the
  finished deck (`shopFinalized: true`), so Auto-draft skips the shop too. There
  is no separate Auto-draft algorithm in `src/`.
- Where a "human" view helps, a **random picker** stands in: same offers, a
  uniformly random card from each one (it ignores types, so it gets little lean).

---

## (a) How many of the 20 picks end up in the deck

A pick counts as "in the deck" when at least one copy of its name is in the
final 20. Copies per pick today: **2** (one pick allows 2 copies at deckbuild;
`docs/rules.md` §1, `COPY_CAP_BY_NAME = 2` in `src/draft/expansions.ts`, the
`2` targets in `botDeckBuilder.fillFrom`). A picked name is never offered again
(#124), so the 20 picks are 20 different names.

| drafter (500 drafts) | mean | median | min | max |
|---|---|---|---|---|
| draft bot | 51.8% | 50.0% | 40% | 60% |
| Auto-draft | 51.8% | 50.0% | 40% | 60% |

Distribution (picks in the deck, of 20):

| picks in deck | bot drafts | Auto-draft drafts |
|---|---|---|
| 8 | 2 | 3 |
| 9 | 58 | 54 |
| 10 | 223 | 231 |
| 11 | 193 | 188 |
| 12 | 24 | 24 |

What happened to each pick (bot; Auto-draft within 0.5 points):

| pick fate | share of all picks |
|---|---|
| in the deck at 2 copies | 31.8% |
| in the deck at 1 copy | 20.0% |
| Trainer left out | 0.7% |
| Pokémon off the deck's 2 attack types | 0.9% |
| Pokémon on the deck's types, but no room | **46.6%** |

Where the 20 deck slots come from:

| slot source | share of slots |
|---|---|
| a picked card | 83.6% |
| a free under-evolution | 16.4% |
| shop | 0% (bot and Auto-draft never shop) |

Bot decks holding any Stage 2 card: 15 of 500 (3.0%); Auto-draft 11 of 500 (2.2%).

**Reading.** The owner's sense that 35–60% of picks are not contenders matches:
40–60% of picks never reach the bot's deck (mean 48%). Almost all of the loss is
structural, not bad picks: with 2 copies per pick plus free under-evolutions, 20
slots hold only about 10 names, so the deck is full after roughly half the picks.

What-if, **1 copy per pick** (same 500 bot drafts, each pick added once with its
under-evolutions at 1 copy, in pick order; a measurement, not a rule change):
picks in the deck mean **80.2%**, median 80%, min 65%, max 95%; 481 of 500
drafts reach 20 cards from picks + free under-evolutions alone (the other 19
would need a shop card or a second copy).

## (b) Build-around opening pick: ex / high tier by type

Nothing to simulate: a pick that ignores type lean draws from the whole tier.
The pool today (434 cards, `onlyWorkingText=true`):

| type | pool cards | ex (tier 1) | Mega ex | ◊◊◊ (tier 3) | ◊◊ | ◊ | Stage 2 (of which ex) |
|---|---|---|---|---|---|---|---|
| Grass | 35 | 1 | 0 | 7 | 15 | 12 | 6 (0) |
| Fire | 14 | 0 | 0 | 3 | 7 | 4 | 3 (0) |
| Water | 47 | 4 | 0 | 11 | 20 | 12 | 7 (1) |
| Lightning | 21 | 0 | 0 | 5 | 11 | 5 | 2 (0) |
| Psychic | 39 | 3 | 0 | 15 | 14 | 7 | 3 (0) |
| Fighting | 35 | 2 | 0 | 5 | 19 | 9 | 4 (1) |
| Darkness | 24 | 3 | 0 | 8 | 8 | 5 | 2 (0) |
| Metal | 19 | 0 | 0 | 6 | 10 | 3 | 3 (0) |
| Dragon | 1 | 0 | 0 | 0 | 1 | 0 | 0 (0) |
| Colorless | 40 | 1 | 0 | 9 | 12 | 18 | 4 (0) |
| Trainer | 159 | 0 | 0 | 0 | 141 | 18 | 0 (0) |
| **all** | 434 | **14** | 0 | 69 | 258 | 93 | 34 (2) |

Only 14 ex are in the pool, and Fire, Lightning, Metal and Dragon have none. For
comparison, the ex in all of `cards.json` (non-cosmetic printings, distinct names,
any text state):

| type | ex in pool | ex in cards.json | Mega ex in cards.json |
|---|---|---|---|
| Grass | 1 | 15 | 4 |
| Fire | 0 | 13 | 5 |
| Water | 4 | 21 | 5 |
| Lightning | 0 | 16 | 2 |
| Psychic | 3 | 14 | 3 |
| Fighting | 2 | 15 | 4 |
| Darkness | 3 | 11 | 3 |
| Metal | 0 | 9 | 4 |
| Dragon | 0 | 3 | 2 |
| Colorless | 1 | 15 | 3 |

A "god pack" of ex today would repeat the same 14 cards; its variety depends on
card-text work (the working-text filter) more than on the offer rule.

## (c) Mega ex

- `public/data/draft-pool.json`: `"onlyWorkingText": true`; the switch is
  `onlyWorkingTextInDraft = true` in `server/draft_text_filter.go`, applied in
  `toDraftablePool` (`keepWorkingText`).
- `cards.json`: **111** Mega ex ids (35 names; 37 non-cosmetic printings).
  Working text: **0**. In the draft pool: **0 of 111**. The shop's same-type ex
  shelf uses only working cards (`App.tsx` `shopPool`), so the shop has none either.
- By name: 10 Basic, 12 Stage 1, 13 Stage 2.
- Why: **all 35** have an attack whose text has no authored effect; 3 also have
  no working earlier stage. None is blocked by an Ability. 35 distinct missing
  attack texts, 7 of which are also printed on a non-Mega card. Samples:
  - Mega Ampharos ex: "1 of your opponent's Benched Pokémon is chosen at random 3 times. For each time a Pokémon was chosen, also do 20 damage to it."
  - Mega Camerupt ex: "1 of your opponent's Pokémon is chosen at random. Do 160 damage to it."
  - Mega Swampert ex: "Discard 2 random Energy from among the Energy attached to all Pokémon (both yours and your opponent's)."
  - Mega Blaziken ex: "Discard a [R] Energy from this Pokémon. Your opponent's Active Pokémon is now Burned."
  - Mega Latios ex: "Discard all Energy from this Pokémon."
  - Mega Rayquaza ex: "Discard all [R] and [L] Energy from this Pokémon, and this attack does 50 damage for each Energy you discarded in this way."
  - Mega Gyarados ex: "Discard the top 3 cards of your opponent's deck."
- Mega rules already in the engine: a KO scores 3 (`koPoints`, `engine/engine.go`,
  spec E1; `IsMegaEX` set in `server/carddata.go`); Serena searches a Mega
  (`server/texts_3_11.go`); a Stadium leaves Mega ex out of doubled Weakness
  (`server/texts_3_13.go`); `DrawPerMegaInPlay` (`engine/texts_3_9.go`); Calem
  needs a Mega in play.
- **No "one Mega per deck" rule** exists in the engine or in deck checks, and
  `docs/rules.md` §3 "Copy limits" says Mega ex follow the plain 2-copy cap. Whether
  Pocket has such a rule is a rule question for the owner; this doc does not answer it.

So (c) is card-text work (35 attack texts) plus 3 earlier stages, not a missing
Mega mechanic.

## (d) Shop: Trainers and Items

**What the shop can offer** (`src/shop/inventory.ts`, `src/shop/staples.ts`;
there is no Go twin, and the bot never shops):

1. Staple shelf, 12 fixed Trainers, minus names already drafted: Cyrus, Sabrina,
   Repel (Item), Copycat, Giovanni, Pokémon Communication (Item), Erika, Mars,
   Iono, Cynthia, Pokémon Center Lady, Lillie. All 12 have working text, so all 12
   show. 10 are Supporters, 2 are Items.
2. Same-type ex shelf: Pokémon only. It never offers a Trainer.

**Reach.** Trainer names: 163 in `cards.json`, 159 with working text, 159 in the
draft pool. All 12 staples are also in the draft pool, so the shop adds no Trainer
the draft could not already offer. It only lets you get one for sure. Reachable
(draft ∪ shop): 159.

**Rare Candy and search / tutor cards.** These are cards whose text matches
"from your deck into your hand", "search your deck", "look at the top", or
"from your deck onto your Bench", plus the names asked about:

| card | kind | in cards.json | text works | in draft pool (tier) | on shop shelf |
|---|---|---|---|---|---|
| Rare Candy | Item | 4 printings | yes | yes (◊◊) | no |
| Poké Ball | Item | 2 | yes | yes (◊) | no |
| Pokémon Communication | Item | 3 | yes | yes (◊◊) | **yes** |
| Clemont | Supporter | 2 | yes | yes (◊◊) | no |
| Team Galactic Grunt | Supporter | 2 | yes | yes (◊◊) | no |
| Juliana (random Stage 2 from deck) | Supporter | 2 | yes | yes (◊◊) | no |
| Sightseer (Stage 1s from top 4) | Supporter | 2 | yes | yes (◊◊) | no |
| Arven (Item or Tool) | Supporter | 3 | yes | yes (◊◊) | no |
| Lisia (2 Basics ≤50 HP) | Supporter | 2 | yes | yes (◊◊) | no |
| Serena (a Mega ex) | Supporter | 2 | yes | yes (◊◊) | no |
| Gladion, Puppy-Loving Girl, Team Rocket's Researcher (named Pokémon) | Supporter | 2 each | yes | yes (◊◊) | no |
| Cabbie (Stadium), Traveling Merchant (Tools), Order Pad (Item) | Supporter, Supporter, Item | 2, 2, 1 | yes | yes (◊◊) | no |
| Mythical Slab, Pokédex, Rotom Dex (look at the top) | Item | 1, 2, 1 | yes | yes (◊◊ / ◊ / ◊◊) | no |
| May (2 random Pokémon) | Supporter | 2 | **no** | no | no |

So Rare Candy and the tutors exist and work. The problem is how rarely a draft
shows them: each is one card in a 258-card ◊◊ tier that gets 25% of slots.

| card | drafts that offer it at least once: random picker | draft bot |
|---|---|---|
| Rare Candy | 4.2% | 3.8% |
| Poké Ball | 26.4% | 29.0% |
| Pokémon Communication | 2.6% | 4.2% |
| Clemont | 10.2% | 7.2% |
| Team Galactic Grunt | 9.0% | 9.6% |
| Pokédex | 24.6% | 25.6% |
| Arven | 4.2% | 4.8% |

**Stage 2 lines in the pool:** 34 of 275 Pokémon (2 ex). By type: Water 7,
Grass 6, Fighting 4, Colorless 4, Fire 3, Metal 3, Psychic 3, Darkness 2,
Lightning 2. Stage 2 picks per draft: random picker 1.92, bot 0.20 (the bot prefers
short lines on purpose: `botPickScore` = 12 − line length).

**Proposal (not built).** Add a third shop shelf, "Evolution kit", of **4
utility Trainers that are always offered**, at the usual 1 ticket a copy:
**Rare Candy, Poké Ball, Juliana, Sightseer**. All four already have working
text and are in the pool, so it needs no new engine text. Pokémon Communication
stays on the staple shelf. With this shelf, a drafter who takes a Stage 2 sees
Rare Candy in 100% of drafts instead of about 4%, and a Basic or Stage 2 tutor
in 100% instead of about 26% / 4%. The shop stays at 4 tickets, so it stays
"patch holes" (rules.md §2): at most 2 Rare Candy + 2 tutors. Possible variant:
show the shelf only when the draft holds a Stage 2 card (the random picker
averages 1.92 Stage 2 picks per draft; the bot averages 0.20). The owner decides; either way it is
a draft-format change.

## (e) How many Trainers are offered and kept

Pool baseline: 159 Trainers of 434 cards (36.6%). Trainers are only in tiers
◊◊ (141 of 258) and ◊ (18 of 93). Weighted by slot tier, a slot with no type
lean is a Trainer **23.3%** of the time. 47 of the 159 Trainers (29.6%) count
as type support today (`supportTypes`).

Trainer share of offer slots, by pick phase (500 drafts each):

| phase | draft bot | Auto-draft | random picker |
|---|---|---|---|
| early (picks 1–7) | 19.4% | 19.5% | 19.0% |
| mid (8–14) | 14.8% | 14.7% | 14.1% |
| late (15–20) | 12.3% | 12.2% | 11.8% |

Type-support share among the Trainers offered, and how many of those support a
type already picked:

| phase | bot: support share | bot: on a picked type | random: support share | random: on a picked type |
|---|---|---|---|---|
| early | 25.3% | 33.0% | 27.3% | 43.6% |
| mid | 38.4% | 70.2% | 41.6% | 80.3% |
| late | 43.6% | 82.9% | 50.2% | 90.8% |

Trainers kept:

| | Trainer share of picks | Trainer share of final-deck slots |
|---|---|---|
| draft bot | 20.6% | 19.9% (3.98 of 20) |
| Auto-draft | 20.8% | 20.0% (4.00 of 20) |
| random picker | 15.4% | (no Go deckbuild for this picker) |

**Reading.** Even before any lean, Trainers fill less of a slot (23%) than of the
pool (37%), because the ◊ tier holds 50% of the slot weight and is mostly Pokémon.
As Pokémon are picked, the type lean raises on-type Pokémon, and only the 30% of
Trainers that support a type keep pace. So the Trainer share of offers falls from
19% to 12% across the draft. The bot's 4 Trainers per deck is its own cap
(`botTrainerPicks = 4`), not a measure of supply. No rule change here; the
type-lean cap work in progress will move these numbers, so re-run after it lands.

## Guesses and limits

- The random picker stands in for a human who ignores types. Real players lean
  into types, so they would see somewhat fewer Trainers than this picker.
- "In the deck" for (a) uses the bot's deckbuilder. A human builds their own deck,
  but the 20-slot limit with 2 copies per pick binds them the same way.
- The 1-copy what-if keeps the bot's 2-type limit and pick order. It is an
  estimate, not a design.
- The tutor regex is broad: it includes "look at the top" cards that only look.
  They are listed separately in the table.

---

## Re-measured after the type-lean cap (m4-draft-runs, 2026-10-28)

Same test, same seeds, after the cap (weight 1 + 2 × min(picks, 4)) and the
text-decided type support merged:

| measure | before | after |
|---|---|---|
| (a) bot: picks with ≥1 copy in the deck, mean (min–max) | 51.8% (40–60%) | 51.8% (40–60%) |
| (a) Auto-draft: same | 51.8% (40–60%) | 51.9% (40–60%) |
| (a) what-if 1 copy per pick: picks in deck, mean | 80.2% | 80.2% |
| (a) bot decks holding a Stage 2 | 3.0% | 2.8% |
| (d) drafts offering Rare Candy at least once (bot) | about 4% | 3.2% |
| (e) bot: Trainer share of offers, early / mid / late | 19.4 / 14.8 / 12.3% | 19.8 / 16.0 / 14.9% |
| (e) type-support share of offered Trainers, early / mid / late | 25 / 38 / 44% | 27 / 39 / 42% |
| (e) Trainer share of final decks (bot) | 19.9% | 20.0% |

The cap and the wider text-decided support moved late-draft Trainer offers up a
little; nothing else changed in a way that matters for (a)–(e).
