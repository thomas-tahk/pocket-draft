# Status

<!-- Every chief PR updates this file. It exists so the next run — and you,
     three weeks from now — do not start cold. Keep it short enough to read. -->

## Where this project is

A Pokémon TCG Pocket **draft** tool plus a **playable simulator**, aimed at the
north star in `docs/VISION.md`: the real battle game — drafting *and* faithful
gameplay including card effects. Effects are core, not a stretch goal.

Three pieces, one repo:

| Piece | What it is |
|---|---|
| The SPA | React/TS/Vite draft UI. Deck construction is enforced: 20 cards, ≤2 copies |
| `engine/` | Go rules engine. Stdlib only, deterministic via a seeded RNG (ADR-0005) |
| `server/` | Go HTTP server + a vanilla-JS board on `localhost:8080`. Hotseat and vs-bot |

## What works end to end

- A full game is playable in a browser on the real engine with real scraped
  cards — the cloud probe drove it start to finish on 2026-09-18.
- Card effects, **slice 1**: five `EffectOp` verbs in `engine/effects.go`
  (`FlipForBonus`, `DamagePerHeads`, `ApplyStatus`, `DamagePerEnergy`,
  `DrawCards`), wired to five real cards in `server/carddata.go`. Effects are
  authored as data, never per-card code — Approach A, see ADR-0002.
- The bot drafts its own 16-round deck rather than using a fixed preset
  (`server/draft_rarity.go`).
- Tests and the production build now run on every pull request (`ci.yml`).

## What is mocked, stubbed, or hardcoded

- **Only 5 cards of the whole set have effects.** Every other card plays at base
  damage. This is deliberate per ADR-0002 (add coverage one card at a time), not
  an oversight — but it means "the game works" is still a narrow claim.
- **Player 1's deck is still a curated three-card stand-in** (`fireDeckList` in
  `server/carddata.go:212`) when no drafted deck is supplied. Only the *bot's*
  deck became a real draft.
- **No multiplayer.** Hotseat and vs-bot only; the WebSocket/two-machine step is
  unstarted.
- **The board's own copy is out of date and contradicts the code.**
  `server/static/index.html:56` still reads "Effects not yet implemented — cards
  play at base damage" and describes both decks as presets. Both stopped being
  true. Player-visible wording is `your-call`, so this needs the owner, not a run.

## Next

- Effect slice 2: a verb that reads game state rather than flipping a coin.
  Issue #5 (Illumise's Ire-Fly, +60 if Volbeat is in the discard) is written up
  with a done-gate already and is the natural next one.
- Give player 1 a real drafted deck, closing the last preset.
