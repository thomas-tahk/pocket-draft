# Draft Mode — Full Flow (Design)

_2026-07-12. Milestone design. The "why" and destination live in `docs/VISION.md`;
this is the concrete architecture and phasing to ship the **full draft-mode loop**._

## Purpose & framing

One app, one card pool, one game engine, one board. Draft / constructed / replay
are **callers of a shared match core**, not separate silos. This milestone ships
**draft mode's complete loop, start to finish**:

> draft picks → finalize deck → play *that deck* in a **run** of games → run ends
> with a persisted win/loss record → (later) step through any game as a replay.

Card **effects/abilities** and a **legible, good-looking board** are core parts of
this loop, not add-ons. Constructed mode and replay come after and reuse the same
core.

**The silo, concretely:** today `App.tsx` does `if (isPlay) return <GameView/>` —
the game is a dead-end URL short-circuit that boots with the Go server's
**hardcoded preset decks**, disconnected from your draft. The fix is to make "play"
always *take a deck as input* and wrap games in a run. That one change un-silos the
whole app.

## Locked decisions

- **Construction enforced:** a deck must be exactly **20 cards** with **≤2 copies of
  any card by name** to enter a game. Equal card *access* is preserved — you can
  proxy any card freely; construction format is a rule of fair play, not an access
  gate. _(This reverses the earlier "no gating at all" note in VISION/memory, which
  conflated card access with deck format.)_
- **Effects via a shared primitive toolbox (Approach A):** ~30-50 reusable effect
  "verbs" that each card composes, plus an imperative escape hatch for genuine
  one-offs. Not per-card bespoke handlers.
- **In-game rules always enforced by the engine, and card text overrides base
  rules** (abilities that accelerate energy, conditional bonus damage, status
  modifiers, on-board stat changes, etc.). This enforcement is never removed.
- **Draft-run opponent:** a bot playing a legal 20-card deck drafted from the same
  shared pool (fair mirror). Multiplayer is out of scope for this milestone.
- **Run end condition:** first to **3 losses** or **9 wins** ends the run.
- **Out of scope this milestone:** constructed mode, multiplayer/networking,
  animations/visual flair, collectibility/economy/rarities, trade.

## Architecture — shared core + mode callers

```
                 ┌─────────────────── shared core ───────────────────┐
 card pool ──►   │  engine (Go, event-sourced)  ──►  effect toolbox   │
 (data/, done)   │        ▲                                           │
                 │        │  builds game from given decks             │
                 │   deck bridge  ◄──────────────────────────────┐    │
                 │        │                                       │    │
                 │   match host (Go server: /api/new|state|move) │    │
                 │        │                                       │    │
                 │   board (React, renders view + dispatches)     │    │
                 └────────┼───────────────────────────────────────────┘
                          │
   mode callers:   draft-run wrapper   │ (later) constructed queue │ (later) replay
```

- **Card pool** — `src/data/`, scraped `public/data/cards.json`. Done.
- **Engine** — Go, event-sourced/deterministic/seeded (ADR-0005), three-mode
  interaction + prompt seam (ADR-0007). Turn loop, energy, KO/points, setup,
  retreat, evolve, poison/burn exist. **No effects yet.**
- **Effect toolbox** — new. Primitive verbs + per-card effect specs (below).
- **Deck bridge** — new. Turns a chosen deck (card-id list) into an engine game,
  replacing the hardcoded preset. Validates construction (20/≤2) first.
- **Match host** — the existing Go server holding one in-memory game. Sequential
  games in a run reuse it as-is (no rooms needed until multiplayer).
- **Board** — React at `?play` today; redesigned into the real game surface.
- **Draft-run wrapper** — new. Sequences games with your finalized deck, tracks
  W/L, ends the run, persists the record.

## Components & interfaces

### 1. Deck bridge
- `POST /api/new` accepts decks as input — `{ you: [cardId...], opponent: [cardId...], seed? }`
  — and builds the engine game from them. The hardcoded `curatedDecks()` preset is
  removed from the default path (kept only as a fallback/bot deck source if useful).
- **Construction validation** runs before a game starts: exactly 20 cards, ≤2 per
  name. Malformed decks are rejected with a clear reason; the game state is
  unchanged. Draft decks are already well-formed, so this mainly guards constructed
  mode and hand-edited decks.
- **Opponent deck:** generated as a legal 20-card deck drafted from the draftable
  pool (reuse the existing offer/draft logic, or a simpler random-legal-20 to
  start). Default = random legal 20 from the draftable pool; revisit at Phase 1.
- The board's `newGame()` sends **your finalized `draftStore.deck`** (expanded
  `cardId→count` into a card-id list) instead of calling the preset endpoint.

### 2. Effect engine (the meaty part)
- **Vocabulary (~30-50 verbs)**, by category:
  - _Damage modifiers:_ bonus-if-condition, per-count (per energy / per tool / per
    benched), coin-scaled (N per heads), self-damage, spread to bench.
  - _Healing_ and _damage removal._
  - _Status:_ apply / extend / modify poison, burn, sleep, paralysis, confusion.
  - _Energy:_ attach, discard, accelerate, move between Pokémon.
  - _Card flow:_ draw, discard, search, reveal.
  - _Targeting:_ self / opponent-active / a chosen bench / all.
  - _On-board modifiers:_ weakness, retreat cost, damage taken/dealt during a turn.
- **EffectSpec** — a card's attack/ability effect is an ordered list of primitive
  invocations with params, conditions, and targets. Data, not code, for the common
  case.
- **Mapping** — authored per **distinct effect string** (512 attack + 143 ability
  ≈ 655; parameterization collapses this further). Keyed so shared texts are done
  once. Coverage prioritized to the **draftable pool** first, then widened.
- **Escape hatch** — genuinely unique effects get a bespoke handler that implements
  the same primitive interface, so the board/engine treat them uniformly.
- **Choices** (coin flips, target selection, discard-from-hand) route through the
  engine's existing **prompt system** (ADR-0007), keeping the engine deterministic.
- **Fidelity** — breadth-first at basic fidelity across the pool, then refine each
  card toward real-game accuracy by adjusting its spec or fixing the shared verb.
  Fixing one verb cascades to every card using it. Bugs post-release are expected
  and fixed one by one.

### 3. Run wrapper + records
- **Run state:** `{ deck, wins, losses, games: GameResult[] }`; ends at 3 losses or
  9 wins.
- **Persistence:** extend the existing `DraftHistoryEntry` (it already carries
  `wins`/`losses`) to hold the run outcome and a per-game **move log** (for replay).
  Reuses the existing zustand `persist` (localStorage).
- **Run screen:** shows deck, run progress (W/L toward the thresholds), launches
  each game via the board, and records the result on return.

### 4. Board redesign
- All zones legible: **active, bench (0-3), hand, energy zone + next energy,
  discard, points (0-3), deck count**, plus **status badges** and damage/modifier
  indicators. Full card detail on demand (tap/hover).
- Its own aesthetic — clear and good-looking, **not** a Pocket clone. Built with the
  frontend-design skill.
- Co-evolves with effects, since effects introduce new things to show (statuses,
  coin-flip outcomes, stat modifiers).

### 5. Shell un-silo
- Replace the `?play` short-circuit with an in-app view state
  (`draft | shop | deckbuild | review | run | game | replay`) so finishing a deck →
  starting a run → playing games → returning is continuous, no dead-ends.

## Data & persistence

- **Finalized deck** — `draftStore.deck` (`cardId→count`) is the handoff artifact
  between draft and play. Already exists.
- **Run record** — extended `DraftHistoryEntry` with run outcome + per-game move
  logs. localStorage via existing persist.
- **Event log per game** — the engine is event-sourced, so a game's move log is
  captured and replay is "re-apply the log," not new game logic.

## Phasing

Each phase gets its own implementation plan (writing-plans) and verification.

- **Phase 0 — Un-silo the shell.** View state replaces the `?play` dead-end. Small,
  structural, unblocks the rest.
- **Phase 1 — Deck bridge.** Play your finalized drafted deck vs a bot; construction
  enforced. → _You play your actual deck._
- **Phase 2 — Effect engine.** Primitives + map the pool's effects at basic
  fidelity. → _Cards do what their text says._
- **Phase 3 — Run wrapper + records.** → _Full loop: draft → finalize → run →
  persisted record._
- **Board** designed properly across Phases 1-3, so you never see a debug harness.
- **Later:** replay (cheap — event log exists) → constructed mode (reuses the bridge
  + run wrapper).

## Testing & verification

- **Engine/primitives:** deterministic seeded unit tests per verb.
- **Deck bridge:** builds a game from given decks; rejects malformed (19 cards, 3
  copies) with state unchanged.
- **Effect specs:** table-driven tests sampling representative cards per verb.
- **End-to-end:** draft → finalize → run → game to a result → record persisted →
  (later) replay steps through the log.
- **UI:** manual browser check (project convention; no JS test runner).

## Open / deferred (non-blocking)

- Bot opponent deck source — default random-legal-20 from the draftable pool;
  confirm at Phase 1.
- Roguelike run flourishes (rewards/meta beyond W/L) — out for now.
- PR #1 (`feat/gameplay-ui`) — merge first vs build-on-top. Project-management
  decision surfaced separately; this branch builds on its HEAD either way.
