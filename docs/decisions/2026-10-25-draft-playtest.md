# Draft play-test + PR #112 decisions (owner, 2026-10-25)

Played on `main` @ 6e17d22 (not the PR #112 preview), then on PR #112.

## PR #112's decisions
1. **Copies per pick: keep 2.** Strict HS Arena would be 1, but 20 Pocket cards at 1 copy each would be "non-cohesive, forced together non-sense".
2. **Offers that fit: go further.**
   - A niche Trainer and the Pokémon it supports raise each other's odds, in both directions, whichever arrives first.
   - Type and archetype support cards (each type has some) become more common once you lean that type.
   - These are raised odds, never guarantees.
3. **Type lean: much earlier, and wider.**
   - The lean should show after 2–3 Pokémon of one type, not 8.
   - It boosts that type's support Trainers and Items, not only Pokémon of the type.
4. **Poké Ball and Professor's Research: keep them draftable for now.** Arguments go both ways: they raise the floor and the speed, and may reduce diversity. Revisit once real players have played; if it's bad for the mode, treat them like any other card.
5. **Merge:** not answered yet.

## New rules
- **A pick is secured for the whole draft.** Once picked, a card never shows in later offers; one pick already allows up to 2 copies at deckbuild. Today a card you already hold can be offered and picked again.
- **Energy selection is missing.** The player must be able to choose the deck's Energy Zone types. The engine accepts chosen types and falls back to a default; no screen offers the choice.
- **Eevee ex becomes draftable** (main and the rework). It is a Basic, so the final-evolutions-only pool drops it today. Picking an Eeveelution should raise its odds. It enables a full playset of 2 different Eeveelutions (or 4 different ones at 1 each).

## Parked
- A community-voted, curated card pool. Not now.

## Answered from the code
- **Shop ex cards:** the rework's shop still offers every ex / Mega ex of your drafted Pokémon's types, minus names you already drafted (`src/shop/inventory.ts`). `docs/rules.md`'s shop section still describes the older "undrafted shelf" and needs updating.
- **Sylveon ex's Happy Ribbon:** prompting yes/no at the moment of evolving is correct. The card text reads "when you play this Pokémon from your hand to evolve 1 of your Pokémon, you may draw 2 cards".
