# Gameplay UI Spec — React board for the simulator

**Status:** proposed · **Date:** 2026-07-09
**Relates to:** design.md Phase 2, ADR-0001 (Go rules-engine), ADR-0004 (show state, not theater), ADR-0006 (client-side prediction)

## Goal

Play a full simulated game — real curated decks, real card art — in the **React app** (the draft-tool stack that deploys to Vercel), driven by the existing Go simulator server. Replace the throwaway `server/static` vanilla page as the place gameplay is visualized.

Success = from a browser I can start a game, take every legal action for both sides through to a winner, and read the board clearly at every step, with responsive selection/target feedback and visible HP/energy changes.

## What already exists (not rebuilt)

The Go server (`server/`) is the source of truth and already exposes everything the UI needs:

- `POST /api/new[?seed=N]` — start a fresh game, returns full state.
- `GET  /api/state` — current state.
- `POST /api/move` — apply one move `{type, player, ...}`; returns `{ok, error?, state}`. Illegal moves change nothing and come back `ok:false` with a reason.
- `POST /api/bot` — advance the active side one greedy move (existing bot policy).

The returned `gameView` (see `server/view.go`) already carries: both players' hands (with card `image`), active + bench (each with `damage`, `remainingHp`, `energy`, `totalEnergy`, status flags), `points`, `deckCount`, `discardCount`, `energyZone`, whose turn (`active`), `turn`, `phase`, `pending` prompt, `winner`, and a `narration` log. **No backend work is required for this slice.**

Move types the UI dispatches (from `toEvent` in `server/main.go`): `SetupPlace`, `PlayBasic`, `AttachEnergy`, `Evolve`, `Retreat`, `UseAttack`, `EndTurn`, `ChooseNewActive`.

## Architecture

React is a **pure renderer + dispatcher**. It renders `gameView` and POSTs moves; the Go engine validates and applies all rules. No rules are reimplemented in TS in this slice. (ADR-0006 foresees TS rules eventually for client-side prediction; explicitly deferred here — the honest minimal is server-authoritative with full redraw per move.)

```
click → POST /api/move → Go engine validates & applies → new gameView → redraw
```

### Where it lives

A separate **mode**, not folded into the draft state machine. `App.tsx` checks the URL: if `?play` is present, render `<GameView>` instead of the draft flow. The draft app (draft → shop → deckbuild → review) is untouched; the game is its own screen. Zero risk to existing features.

### Dev wiring

- Vite dev server (`:5173`) proxies `/api` → `localhost:8080` (Go server) via `vite.config.ts`. No CORS handling, code matches a same-origin prod deploy.
- Production deploy of the realtime path is out of scope for this slice (separate Go server on Fly/Railway per prior decisions); this slice targets local dev against `localhost:8080`.

### State

A dedicated store (`src/game/` — Zustand, matching `draftStore` conventions) holds the current `gameView`, in-flight status, last error, and a client-side selection (which card/target the player has picked, pre-dispatch). Every successful move replaces the state wholesale with the server's returned `state`.

## Components

Mirror the server's view model; reuse existing card-rendering where practical.

```
GameView                     ?play entry; owns the game store; redraws per move
 ├─ PlayerSide (opponent)    active slot, bench row, hand-count, points/deck/discard
 ├─ BoardCenter              turn #, phase, active-player indicator, winner banner
 ├─ PlayerSide (you)         active slot, bench row, points/deck/discard
 ├─ Hand                     your cards, clickable to select/play
 ├─ ActionBar                context actions: Attach energy · Attack (per index) ·
 │                           Retreat · End turn · Bot move · New game
 └─ Log                      the server's `narration`, newest last
```

- **CardArt** reuses the draft app's existing card image rendering so the board is visually consistent with the rest of the product.
- **Pending prompts** (`gameView.pending`, e.g. `setup`, `new_active`) drive what's actionable: during setup the UI collects a `SetupPlace`; on a knockout it collects `ChooseNewActive`.
- Attacks render one button per attack index with name/cost/damage; disabled (with reason on hover) when unaffordable, matching the server's own validation so the client refuses obviously-illegal actions instantly (ADR-0004 "refuse, don't block").

## Micro-feedback (the polish, in v1)

Within ADR-0004 — show state, add no theater, never make the player wait:

- **Selection & targets:** selected card and legal targets get a highlight ring; hover states on every clickable element.
- **HP / energy changes:** when a value changes between the previous and new `gameView`, the number updates with a brief highlight (no animation the player waits through). This requires diffing prev vs. next state in the store — small but real; if it threatens the first playable build it drops to a fast-follow, but the target is to include it.
- **Illegal action:** the returned `ok:false` reason shows inline instantly; nothing on the board changes.

Explicitly **not** built: attack cinematics, card-play flourishes, spinning coins, any blocking animation. (ADR-0004.)

## Build order (each step is a commit/push checkpoint)

Delivered "everything at once" — the user tests the finished board, not intermediate versions — but built and committed in this internal order so history stays legible and a wall mid-way is surfaced, not bulldozed:

1. **Read-only render.** Vite proxy + sim client + game store; `<GameView>` fetches and draws state (no actions). → board reflects a real game's state.
2. **Actions.** Wire every move type through the ActionBar / clicks; play a full game start-to-finish for both sides in React. → a game reaches a winner via the UI.
3. **Selection & targeting.** Highlight rings, hover, legal-target affordances, inline illegal-move reasons. → the board feels responsive.
4. **Change feedback.** Prev/next diff for HP & energy with brief highlight. → the polish pass.

## Out of scope

- Reimplementing any rules in TS / client-side prediction (ADR-0006, later).
- Card effects beyond base damage (ADR-0002 incremental coverage — separate track).
- Networked multiplayer / rooms; production deploy of the sim server.
- Replay (design.md Phase 3).
- Animation systems of any kind (ADR-0004).

## Open questions

- None blocking. Deck selection is fixed to the two curated decks for now (Fire vs. Water); a deck picker is a later nicety, not part of this slice.
