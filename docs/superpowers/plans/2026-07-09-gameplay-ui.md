# Gameplay UI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Play a full simulated Pokémon TCG Pocket game — real curated decks, real card art — inside the React app, driven by the existing Go simulator server, with responsive selection/target feedback and visible HP/energy changes.

**Architecture:** React is a pure renderer + dispatcher. It renders the Go server's `gameView` JSON and POSTs moves to `/api/*`; the Go engine validates and applies all rules. Every successful move replaces the client state wholesale with the server's returned state. No rules are reimplemented in TS in this slice (ADR-0006 client prediction deferred). The board lives at `?play` as its own screen; the draft app is untouched.

**Tech Stack:** React 18 + TypeScript + Vite + Zustand (all already in the project). Go simulator server (`server/`, `localhost:8080`) is unchanged — no backend work. No new dependencies.

## Global Constraints

- **No new npm dependencies.** Verification is `npm run lint` (tsc typecheck) + manual browser check against the running Go server. There is no frontend test runner and this plan does not add one.
- **No rules in TS.** The client never decides legality on its own authority; it may *mirror* the server's affordability for disabling buttons, but the server is the source of truth and its `ok:false` reason is always shown.
- **ADR-0004 — show state, never theater.** No animation the player waits through. Feedback = instant highlight of a state fact (selection ring, a number that changed). No attack cinematics, card flourishes, or spinning coins.
- **Follow existing conventions.** Zustand store shaped like `src/stores/draftStore.ts` (State + Actions split, `create`). Inline `style={{}}` objects as the rest of `src/` uses; no CSS framework.
- **New code lives in `src/game/`.** The only file outside it that changes is `src/App.tsx` (the `?play` branch).
- **Server contract is fixed** (see `server/view.go`, `server/main.go`, `server/bot.go`): endpoints `POST /api/new[?seed=N]`, `GET /api/state`, `POST /api/move`, `POST /api/bot`. `winner` is `-1` while the game is ongoing, else `0` or `1`. Bench holds up to 3. `AttachEnergy`/`UseAttack` act on the Active; `Target:0` = Active. During `phase:"setup"` each side submits one `SetupPlace`; on a knockout the engine sets `pending.kind:"new_active"` and waits for a `ChooseNewActive`.

---

## File Structure

- `vite.config.ts` — **modify**: add a dev proxy so `/api` → `localhost:8080`.
- `src/game/types.ts` — **create**: TypeScript mirror of the server's `gameView` and the move payloads.
- `src/game/api.ts` — **create**: thin fetch client (`newGame`, `getState`, `sendMove`, `botMove`).
- `src/game/store.ts` — **create**: Zustand game store — current view, previous view, status, error, client-side selection, and actions.
- `src/game/diff.ts` — **create**: pure helper that reports which Active/bench HP and energy values changed between two views (drives change-highlights).
- `src/game/GameCard.tsx` — **create**: renders one `CardView` (art + name); selectable/highlight/disabled props.
- `src/game/GameView.tsx` — **create**: `?play` entry; owns the store lifecycle and lays out the board.
- `src/game/PlayerSide.tsx` — **create**: one player's active slot + bench + counts (points/deck/discard/hand-count).
- `src/game/Hand.tsx` — **create**: the human's hand row (clickable cards).
- `src/game/ActionBar.tsx` — **create**: context actions (attach / attack / retreat / evolve / end turn / bot / new game) + pending-prompt controls.
- `src/game/Log.tsx` — **create**: renders `gameView.narration`.
- `src/App.tsx` — **modify**: render `<GameView />` when `?play` is present, before the draft flow.

---

## Task 1: Dev proxy + view/move types

**Files:**
- Modify: `vite.config.ts`
- Create: `src/game/types.ts`

**Interfaces:**
- Produces: `GameView`, `PlayerView`, `InPlayView`, `CardView`, `AttackView`, `PromptView`, `MoveResp`, `BotResp`, and the `Move` union — consumed by every later task.

- [ ] **Step 1: Add the dev proxy.** Replace `vite.config.ts` with:

```ts
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

// The Go simulator serves the game API on :8080. In dev the Vite server (:5173)
// proxies /api there so client code can use same-origin relative URLs.
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
    },
  },
});
```

- [ ] **Step 2: Create the view + move types.** Create `src/game/types.ts` — a faithful mirror of `server/view.go` and the `moveReq` fields in `server/main.go`:

```ts
// Mirrors server/view.go. The client renders these; it does not own them.
export type AttackView = { name: string; cost: string[]; damage: number };

export type CardView = {
  id: string;
  name: string;
  image?: string;
  stage: string;
  hp: number;
  type: string;
  isEx: boolean;
  evolvesFrom?: string;
  retreatCost: number;
  weakness?: string;
  attacks: AttackView[];
};

export type InPlayView = {
  card: CardView;
  damage: number;
  remainingHp: number;
  energy: Record<string, number>;
  totalEnergy: number;
  poisoned: boolean;
  burned: boolean;
};

export type PlayerView = {
  hand: CardView[];
  active: InPlayView | null;
  bench: InPlayView[];
  points: number;
  deckCount: number;
  discardCount: number;
  energyZone: string;
  energyUsed: boolean;
  retreated: boolean;
};

export type PromptView = { player: number; kind: 'setup' | 'new_active' };

export type GameView = {
  players: [PlayerView, PlayerView];
  active: number;
  turn: number;
  phase: 'setup' | 'main' | 'over';
  pending: PromptView | null;
  winner: number; // -1 while ongoing, else 0 or 1
  narration: string[];
};

// Mirrors server toEvent(): one shape per move type. `player` is always the
// actor (gameView.active, or pending.player during a prompt).
export type Move =
  | { type: 'SetupPlace'; player: number; activeCardId: string; benchCardIds: string[] }
  | { type: 'PlayBasic'; player: number; cardId: string }
  | { type: 'AttachEnergy'; player: number; target: number }
  | { type: 'Evolve'; player: number; handCardId: string; target: number }
  | { type: 'Retreat'; player: number; benchIndex: number }
  | { type: 'UseAttack'; player: number; index: number }
  | { type: 'EndTurn'; player: number }
  | { type: 'ChooseNewActive'; player: number; benchIndex: number };

export type MoveResp = { ok: boolean; error?: string; state: GameView };
export type BotResp = { acted: boolean; state: GameView };
```

- [ ] **Step 3: Typecheck.** Run: `npm run lint`. Expected: PASS (no errors). `src/game/types.ts` compiles; `vite.config.ts` is valid.

- [ ] **Step 4: Commit.**

```bash
git add vite.config.ts src/game/types.ts
git commit -m "feat(game): dev proxy and TS mirror of the server game view"
```

---

## Task 2: Sim API client

**Files:**
- Create: `src/game/api.ts`

**Interfaces:**
- Consumes: `GameView`, `Move`, `MoveResp`, `BotResp` from Task 1.
- Produces: `newGame(seed?: number): Promise<GameView>`, `getState(): Promise<GameView>`, `sendMove(move: Move): Promise<MoveResp>`, `botMove(): Promise<BotResp>`.

- [ ] **Step 1: Write the client.** Create `src/game/api.ts`:

```ts
import type { GameView, Move, MoveResp, BotResp } from './types';

// Thin wrappers over the Go server's /api endpoints. Relative URLs resolve via
// the Vite proxy in dev (see vite.config.ts).

async function asJson<T>(res: Response): Promise<T> {
  if (!res.ok) throw new Error(`${res.status} ${res.statusText}`);
  return (await res.json()) as T;
}

export function newGame(seed?: number): Promise<GameView> {
  const q = seed === undefined ? '' : `?seed=${seed}`;
  return fetch(`/api/new${q}`, { method: 'POST' }).then((r) => asJson<GameView>(r));
}

export function getState(): Promise<GameView> {
  return fetch('/api/state').then((r) => asJson<GameView>(r));
}

export function sendMove(move: Move): Promise<MoveResp> {
  return fetch('/api/move', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(move),
  }).then((r) => asJson<MoveResp>(r));
}

export function botMove(): Promise<BotResp> {
  return fetch('/api/bot', { method: 'POST' }).then((r) => asJson<BotResp>(r));
}
```

- [ ] **Step 2: Typecheck.** Run: `npm run lint`. Expected: PASS.

- [ ] **Step 3: Commit.**

```bash
git add src/game/api.ts
git commit -m "feat(game): sim API client"
```

---

## Task 3: Game store

**Files:**
- Create: `src/game/store.ts`

**Interfaces:**
- Consumes: `GameView`, `Move` (Task 1); `newGame`, `getState`, `sendMove`, `botMove` (Task 2).
- Produces the Zustand hook `useGameStore` exposing:
  - state: `view: GameView | null`, `prev: GameView | null`, `status: 'idle' | 'loading' | 'ready'`, `error: string | null`, `selection: Selection`
  - actions: `init(): Promise<void>`, `restart(seed?: number): Promise<void>`, `dispatch(move: Move): Promise<void>`, `runBot(): Promise<void>`, `select(sel: Selection): void`, `clearSelection(): void`
  - `Selection = { kind: 'hand'; cardId: string } | { kind: 'bench'; index: number } | { kind: 'setupActive'; cardId: string } | null`

- [ ] **Step 1: Write the store.** Create `src/game/store.ts`. Mirrors the `draftStore` shape (State + Actions, `create`). No `persist` — the server owns the game; the store is a live cache. `prev` retains the view from before the last successful move so Task 8 can diff. `error` holds the server's last `ok:false` reason and clears on the next successful move.

```ts
import { create } from 'zustand';
import type { GameView, Move } from './types';
import { newGame, getState, sendMove, botMove } from './api';

export type Selection =
  | { kind: 'hand'; cardId: string }
  | { kind: 'bench'; index: number }
  | { kind: 'setupActive'; cardId: string }
  | null;

type State = {
  view: GameView | null;
  prev: GameView | null;
  status: 'idle' | 'loading' | 'ready';
  error: string | null;
  selection: Selection;
};

type Actions = {
  init: () => Promise<void>;
  restart: (seed?: number) => Promise<void>;
  dispatch: (move: Move) => Promise<void>;
  runBot: () => Promise<void>;
  select: (sel: Selection) => void;
  clearSelection: () => void;
};

export const useGameStore = create<State & Actions>((set, get) => ({
  view: null,
  prev: null,
  status: 'idle',
  error: null,
  selection: null,

  init: async () => {
    set({ status: 'loading', error: null });
    const view = await getState();
    set({ view, prev: null, status: 'ready' });
  },

  restart: async (seed) => {
    set({ status: 'loading', error: null, selection: null });
    const view = await newGame(seed);
    set({ view, prev: null, status: 'ready' });
  },

  dispatch: async (move) => {
    const before = get().view;
    const res = await sendMove(move);
    if (!res.ok) {
      // Illegal move: server state is unchanged; surface the reason, keep selection.
      set({ error: res.error ?? 'illegal move', view: res.state });
      return;
    }
    set({ prev: before, view: res.state, error: null, selection: null });
  },

  runBot: async () => {
    const before = get().view;
    const res = await botMove();
    set({ prev: before, view: res.state, error: null });
  },

  select: (sel) => set({ selection: sel }),
  clearSelection: () => set({ selection: null }),
}));
```

- [ ] **Step 2: Typecheck.** Run: `npm run lint`. Expected: PASS.

- [ ] **Step 3: Commit.**

```bash
git add src/game/store.ts
git commit -m "feat(game): zustand game store over the sim API"
```

---

## Task 4: GameCard component

**Files:**
- Create: `src/game/GameCard.tsx`

**Interfaces:**
- Consumes: `CardView` (Task 1).
- Produces: `GameCard` component. Props: `{ card: CardView; size?: 'lg' | 'sm'; selected?: boolean; highlight?: boolean; disabled?: boolean; onClick?: () => void; title?: string }`.

Note: this deliberately does NOT reuse the draft app's `CardTile`, which is bound to the draft `Card` type (`imageThumb`, `draftRarity`). `GameCard` renders the server's `CardView` (`image`, `name`, `id`) directly — a small, honest divergence from the spec's "reuse where practical," made because the type mismatch would be lossy.

- [ ] **Step 1: Write the component.** Create `src/game/GameCard.tsx`:

```tsx
import type { CardView } from './types';

type Props = {
  card: CardView;
  size?: 'lg' | 'sm';
  selected?: boolean;
  highlight?: boolean; // legal target for the current selection
  disabled?: boolean;
  onClick?: () => void;
  title?: string;
};

export function GameCard({ card, size = 'lg', selected, highlight, disabled, onClick, title }: Props) {
  const width = size === 'lg' ? 140 : 84;
  const Tag = onClick ? 'button' : 'div';
  const ring = selected ? '#2b5cff' : highlight ? '#ffce54' : 'transparent';

  return (
    <Tag
      onClick={onClick}
      disabled={disabled}
      title={title ?? `${card.name} · ${card.id}`}
      style={{
        all: 'unset',
        cursor: onClick && !disabled ? 'pointer' : 'default',
        display: 'block',
        width,
        opacity: disabled ? 0.45 : 1,
        outline: `3px solid ${ring}`,
        outlineOffset: 2,
        borderRadius: 8,
      }}
    >
      {card.image ? (
        <img src={card.image} alt={card.name} loading="lazy" style={{ width: '100%', borderRadius: 8, display: 'block' }} />
      ) : (
        <div style={{ width: '100%', aspectRatio: '5 / 7', borderRadius: 8, background: '#33363d', color: '#e6e6e8', display: 'grid', placeItems: 'center', fontSize: 12, textAlign: 'center', padding: 4 }}>
          {card.name}
        </div>
      )}
    </Tag>
  );
}
```

- [ ] **Step 2: Typecheck.** Run: `npm run lint`. Expected: PASS.

- [ ] **Step 3: Commit.**

```bash
git add src/game/GameCard.tsx
git commit -m "feat(game): GameCard renders a server CardView"
```

---

## Task 5: Read-only board — CHECKPOINT 1

**Files:**
- Create: `src/game/GameView.tsx`, `src/game/PlayerSide.tsx`, `src/game/Log.tsx`
- Modify: `src/App.tsx`

**Interfaces:**
- Consumes: `useGameStore` (Task 3), `GameCard` (Task 4), view types (Task 1).
- Produces: `GameView` (default-exported piece used by `App`), `PlayerSide` (`{ player: PlayerView; role: 'you' | 'opponent' }`), `Log` (`{ lines: string[] }`).

This task renders state only — no actions yet. `PlayerSide` shows: the active slot (a `GameCard size="lg"` + an HP line `remainingHp/hp` + an energy summary from `Object.entries(active.energy)` + `totalEnergy`), the bench as up to 3 `GameCard size="sm"`, and a counts line (`points` ● out of 3, `deckCount`, `discardCount`, and for the opponent a hand *count* rather than face-up cards). The human's own hand is rendered later by `Hand` (Task 6); for this checkpoint render `you`'s hand as face-up `GameCard`s too so there is something to see. `BoardCenter` content (turn/phase/active/winner) is folded into `GameView` for now as a simple header.

- [ ] **Step 1: PlayerSide.** Create `src/game/PlayerSide.tsx`:

```tsx
import type { PlayerView, InPlayView } from './types';
import { GameCard } from './GameCard';

function EnergyLine({ mon }: { mon: InPlayView }) {
  const parts = Object.entries(mon.energy).map(([t, n]) => `${t}×${n}`);
  return <div style={{ fontSize: 12, opacity: 0.8 }}>{parts.length ? parts.join(' ') : 'no energy'} · {mon.totalEnergy} total</div>;
}

function ActiveSlot({ mon }: { mon: InPlayView | null }) {
  if (!mon) return <div style={{ fontSize: 13, opacity: 0.6 }}>— no active —</div>;
  return (
    <div>
      <GameCard card={mon.card} size="lg" />
      <div style={{ fontSize: 13 }}>HP {mon.remainingHp}/{mon.card.hp}{mon.poisoned ? ' ☠' : ''}{mon.burned ? ' 🔥' : ''}</div>
      <EnergyLine mon={mon} />
    </div>
  );
}

export function PlayerSide({ player, role }: { player: PlayerView; role: 'you' | 'opponent' }) {
  return (
    <section style={{ padding: 12, border: '1px solid #33363d', borderRadius: 10 }}>
      <div style={{ fontSize: 12, opacity: 0.7, marginBottom: 6 }}>
        {role === 'you' ? 'YOU' : 'OPPONENT'} · pts {player.points}/3 · deck {player.deckCount} · discard {player.discardCount}
        {role === 'opponent' ? ` · hand ${player.hand.length}` : ''}
      </div>
      <div style={{ display: 'flex', gap: 16, alignItems: 'flex-start' }}>
        <ActiveSlot mon={player.active} />
        <div style={{ display: 'flex', gap: 8 }}>
          {player.bench.map((mon, i) => (
            <div key={i}>
              <GameCard card={mon.card} size="sm" />
              <div style={{ fontSize: 11 }}>{mon.remainingHp}/{mon.card.hp} · {mon.totalEnergy}⚡</div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
```

- [ ] **Step 2: Log.** Create `src/game/Log.tsx`:

```tsx
export function Log({ lines }: { lines: string[] }) {
  return (
    <div style={{ maxHeight: 160, overflowY: 'auto', fontSize: 12, lineHeight: 1.5, border: '1px solid #33363d', borderRadius: 8, padding: 8 }}>
      {lines.length === 0 ? <div style={{ opacity: 0.6 }}>No events yet.</div> : lines.map((l, i) => <div key={i}>{l}</div>)}
    </div>
  );
}
```

- [ ] **Step 3: GameView.** Create `src/game/GameView.tsx`. On mount it calls `init()` (loads current state; the server always has a game running). Renders opponent side (index 1), a center header, your side (index 0), your hand face-up, and the log.

```tsx
import { useEffect } from 'react';
import { useGameStore } from './store';
import { PlayerSide } from './PlayerSide';
import { GameCard } from './GameCard';
import { Log } from './Log';

export function GameView() {
  const { view, status, error, init } = useGameStore();

  useEffect(() => {
    void init();
  }, [init]);

  if (status !== 'ready' || !view) return <main style={{ padding: 24 }}>Loading game…</main>;

  const [you, opp] = view.players;
  const turnLabel = view.phase === 'over'
    ? `Game over — winner: P${view.winner + 1}`
    : `Turn ${view.turn} · phase ${view.phase} · active P${view.active + 1}` + (view.pending ? ` · waiting on P${view.pending.player + 1} (${view.pending.kind})` : '');

  return (
    <main style={{ padding: 16, maxWidth: 900, margin: '0 auto', display: 'grid', gap: 12 }}>
      <PlayerSide player={opp} role="opponent" />
      <div style={{ textAlign: 'center', fontSize: 13, fontWeight: 600 }}>{turnLabel}</div>
      {error && <div style={{ color: '#ff6b6b', fontSize: 13, textAlign: 'center' }}>{error}</div>}
      <PlayerSide player={you} role="you" />
      <div>
        <div style={{ fontSize: 12, opacity: 0.7, marginBottom: 4 }}>Your hand</div>
        <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
          {you.hand.map((c) => <GameCard key={c.id} card={c} size="sm" />)}
        </div>
      </div>
      <Log lines={view.narration} />
    </main>
  );
}
```

- [ ] **Step 4: Mount at `?play`.** In `src/App.tsx`, add the import and an early branch. Insert near the top of `App()`'s render, before the `sharedDeck` check:

```tsx
// add with the other imports
import { GameView } from './game/GameView';
```

```tsx
// add as the first statement inside App(), before any hooks?  No — after hooks,
// place it as the first render branch so it wins over the draft flow:
const isPlay = new URLSearchParams(window.location.search).has('play');
if (isPlay) return <GameView />;
```

Place the `isPlay` branch immediately after the existing `if (error) …` / `if (!pool) …` guards are declared but it must not skip hooks — put it right after the `const phase = derivePhase(...)` line and before `if (error) return …`. (All hooks above it still run; it only short-circuits rendering.)

- [ ] **Step 5: Typecheck.** Run: `npm run lint`. Expected: PASS.

- [ ] **Step 6: Browser verify.** In one terminal: `cd server && go run .` (serves API on :8080). In another: `npm run dev` (Vite on :5173). Open `http://localhost:5173/?play`. Expected: both player sides render with real card art, a turn/phase header, your hand face-up, and the event log. No console errors.

- [ ] **Step 7: Commit.**

```bash
git add src/game/GameView.tsx src/game/PlayerSide.tsx src/game/Log.tsx src/App.tsx
git commit -m "feat(game): read-only React board at ?play (checkpoint 1)"
```

---

## Task 6: Actions — playable game — CHECKPOINT 2

**Files:**
- Create: `src/game/ActionBar.tsx`, `src/game/Hand.tsx`
- Modify: `src/game/GameView.tsx`

**Interfaces:**
- Consumes: `useGameStore` (`dispatch`, `runBot`, `restart`, `select`, `clearSelection`, `selection`), view types.
- Produces: `ActionBar` (`{ view: GameView }`) and `Hand` (`{ player: PlayerView }`).

The actor for every move is `view.pending ? view.pending.player : view.active`. The UI drives whichever side the engine is waiting on; the "Bot move" button advances that same actor one greedy step (handy for playing the opponent or skipping). Interaction model, per engine:

- **Setup (`phase:"setup"`, `pending.kind:"setup"`):** the actor picks one Basic from hand as Active, then up to 3 Basics as bench, then confirms → `SetupPlace{ player, activeCardId, benchCardIds }`. Implement as: click a hand card to set it as pending Active (selection `{kind:'setupActive'}`), click more hand cards to toggle them into a bench list, a "Place" button submits. Keep the bench list in `ActionBar` local `useState`.
- **New active (`pending.kind:"new_active"`):** clicking a bench card submits `ChooseNewActive{ player, benchIndex }`.
- **Main phase:** buttons — `Attach energy` → `AttachEnergy{player, target:0}`; one `Attack: <name> (<dmg>)` button per `active.card.attacks` index → `UseAttack{player, index}`; `Retreat` → after selecting a bench card, `Retreat{player, benchIndex}`; `Evolve` → select a hand card then a target (0=active) → `Evolve{player, handCardId, target}`; `Play basic` → clicking a Basic in hand → `PlayBasic{player, cardId}`; `End turn` → `EndTurn{player}`. Always-present: `Bot move` → `runBot()`; `New game` → `restart()`.

- [ ] **Step 1: Hand.** Create `src/game/Hand.tsx`. Renders the actor's hand; clicking a card sets a `hand` selection (used by Play basic / Evolve) — the actual dispatch is decided in `ActionBar` based on phase.

```tsx
import type { PlayerView } from './types';
import { GameCard } from './GameCard';
import { useGameStore } from './store';

export function Hand({ player }: { player: PlayerView }) {
  const { selection, select } = useGameStore();
  return (
    <div>
      <div style={{ fontSize: 12, opacity: 0.7, marginBottom: 4 }}>Your hand</div>
      <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
        {player.hand.map((c) => {
          const sel = selection?.kind === 'hand' && selection.cardId === c.id;
          const selSetup = selection?.kind === 'setupActive' && selection.cardId === c.id;
          return (
            <GameCard
              key={c.id}
              card={c}
              size="sm"
              selected={sel || selSetup}
              title={`${c.name} · ${c.stage} · ${c.id}`}
              onClick={() => select({ kind: 'hand', cardId: c.id })}
            />
          );
        })}
      </div>
    </div>
  );
}
```

- [ ] **Step 2: ActionBar.** Create `src/game/ActionBar.tsx`. Computes `actor`, branches on phase/pending, and renders the buttons described above. Setup bench selection uses local `useState<string[]>`. Every button calls `dispatch(...)` or `runBot()`/`restart()` from the store.

```tsx
import { useState } from 'react';
import type { GameView } from './types';
import { useGameStore } from './store';

export function ActionBar({ view }: { view: GameView }) {
  const { dispatch, runBot, restart, selection, clearSelection } = useGameStore();
  const [setupBench, setSetupBench] = useState<string[]>([]);

  const actor = view.pending ? view.pending.player : view.active;
  const me = view.players[actor];
  const btn = { fontSize: 13, padding: '6px 12px', marginRight: 6, marginBottom: 6 } as const;

  if (view.phase === 'over') {
    return <div><button style={btn} onClick={() => restart()}>New game</button></div>;
  }

  // Setup: choose Active + up to 3 bench Basics from hand.
  if (view.pending?.kind === 'setup') {
    const activeId = selection?.kind === 'setupActive' ? selection.cardId : '';
    return (
      <div>
        <div style={{ fontSize: 12, marginBottom: 6 }}>
          Setup P{actor + 1}: pick an Active Basic (click a hand card, then “Set Active”), toggle up to 3 bench, then Place.
        </div>
        <button style={btn} onClick={() => {
          if (selection?.kind === 'hand') useGameStore.getState().select({ kind: 'setupActive', cardId: selection.cardId });
        }}>Set Active</button>
        <button style={btn} onClick={() => {
          if (selection?.kind === 'hand') {
            const id = selection.cardId;
            setSetupBench((b) => (b.includes(id) ? b.filter((x) => x !== id) : b.length < 3 ? [...b, id] : b));
          }
        }}>Toggle bench ({setupBench.length}/3)</button>
        <button style={btn} disabled={!activeId} onClick={async () => {
          await dispatch({ type: 'SetupPlace', player: actor, activeCardId: activeId, benchCardIds: setupBench });
          setSetupBench([]);
        }}>Place</button>
      </div>
    );
  }

  if (view.pending?.kind === 'new_active') {
    return (
      <div style={{ fontSize: 12 }}>
        P{actor + 1} was knocked out — click a benched Pokémon to promote it.
        <div><button style={btn} onClick={() => runBot()}>Bot choose</button></div>
      </div>
    );
  }

  // Main phase.
  const attacks = me.active?.card.attacks ?? [];
  return (
    <div>
      <button style={btn} onClick={() => dispatch({ type: 'AttachEnergy', player: actor, target: 0 })}>Attach energy</button>
      {attacks.map((a, i) => (
        <button key={i} style={btn} onClick={() => dispatch({ type: 'UseAttack', player: actor, index: i })}>
          Attack: {a.name} ({a.damage})
        </button>
      ))}
      <button style={btn} onClick={() => {
        if (selection?.kind === 'hand') dispatch({ type: 'PlayBasic', player: actor, cardId: selection.cardId });
      }}>Play basic (selected)</button>
      <button style={btn} onClick={() => {
        if (selection?.kind === 'hand') dispatch({ type: 'Evolve', player: actor, handCardId: selection.cardId, target: 0 });
      }}>Evolve active (selected)</button>
      <button style={btn} onClick={() => {
        if (selection?.kind === 'bench') dispatch({ type: 'Retreat', player: actor, benchIndex: selection.index });
      }}>Retreat → selected bench</button>
      <button style={btn} onClick={() => { dispatch({ type: 'EndTurn', player: actor }); clearSelection(); }}>End turn</button>
      <button style={btn} onClick={() => runBot()}>Bot move</button>
      <button style={btn} onClick={() => restart()}>New game</button>
    </div>
  );
}
```

- [ ] **Step 3: Make benches clickable for selection.** In `src/game/PlayerSide.tsx`, for the `you` role only, wrap each bench `GameCard` with `onClick={() => select({ kind: 'bench', index: i })}` and pass `selected={selection?.kind==='bench' && selection.index===i}`. Import `useGameStore`. (Opponent bench stays non-interactive.) Add a `role`-guarded `onClick` so opponent cards remain display-only.

- [ ] **Step 4: Wire into GameView.** In `src/game/GameView.tsx`, replace the inline hand block with `<Hand player={you} />` and add `<ActionBar view={view} />` above the log. Remove the now-unused inline `GameCard` hand import if it is no longer referenced.

- [ ] **Step 5: Typecheck.** Run: `npm run lint`. Expected: PASS.

- [ ] **Step 6: Browser verify — play a full game.** With both servers running, open `/?play`. Do setup for both sides (use “Bot move” to let P2 set up), then take main-phase turns: attach energy, attack, end turn; use “Bot move” to advance the opponent. Play through to a `Game over — winner` banner. Illegal actions (e.g. attacking with no energy) show the server's reason and change nothing. Expected: a game reaches a winner entirely through the UI.

- [ ] **Step 7: Commit.**

```bash
git add src/game/ActionBar.tsx src/game/Hand.tsx src/game/PlayerSide.tsx src/game/GameView.tsx
git commit -m "feat(game): actions — playable game to a winner (checkpoint 2)"
```

---

## Task 7: Selection & targeting feedback — CHECKPOINT 3

**Files:**
- Modify: `src/game/PlayerSide.tsx`, `src/game/Hand.tsx`, `src/game/ActionBar.tsx`, `src/game/GameCard.tsx`

**Interfaces:** no new exports; refines existing props usage.

Goal: the board *feels* responsive — clear selection rings (already via `GameCard selected`), hover affordance, disabled attack buttons with a reason, and legal-target hinting. Stays inside ADR-0004 (instant, state-only).

- [ ] **Step 1: Hover affordance.** In `GameCard.tsx`, add a subtle hover using `onMouseEnter/onMouseLeave` local `useState` to lift `boxShadow` when `onClick && !disabled`. (Instant, no transition the player waits through; a ≤120ms CSS transition on box-shadow is acceptable as it never blocks input.)

- [ ] **Step 2: Disable unaffordable attacks with a reason.** In `ActionBar.tsx`, compute affordability by mirroring the server: an attack is affordable when the Active's `totalEnergy >= a.cost.length` AND, for each non-`Colorless` symbol in `a.cost`, the Active has at least that many of the matching energy type in `active.energy`. Render the button `disabled` when unaffordable with `title` = "needs <cost>". This is a client mirror for affordance only — the server remains authoritative (Global Constraints).

```tsx
// helper inside ActionBar.tsx
function canAfford(energy: Record<string, number>, total: number, cost: string[]): boolean {
  if (total < cost.length) return false;
  const need: Record<string, number> = {};
  for (const sym of cost) if (sym !== 'Colorless') need[sym] = (need[sym] ?? 0) + 1;
  return Object.entries(need).every(([t, n]) => (energy[t] ?? 0) >= n);
}
```

Apply: `const active = me.active;` then for each attack `const ok = !!active && canAfford(active.energy, active.totalEnergy, a.cost);` → `disabled={!ok}` and `title={ok ? '' : 'needs ' + a.cost.join(' ')}`.

- [ ] **Step 3: Legal-target hint.** When `selection.kind === 'bench'`, pass `highlight` to the Active slot card to signal it as the retreat destination; when `selection.kind === 'hand'` on an evolvable card, pass `highlight` to the Active as the evolve target. Keep it to the yellow ring already in `GameCard`.

- [ ] **Step 4: Typecheck.** Run: `npm run lint`. Expected: PASS.

- [ ] **Step 5: Browser verify.** Selecting a hand/bench card shows a blue ring; legal targets show a yellow ring; hovering a clickable card lifts it; attacks you can't afford are greyed with a tooltip reason; an illegal move still surfaces the server message. Expected: interactions read clearly.

- [ ] **Step 6: Commit.**

```bash
git add src/game/GameCard.tsx src/game/PlayerSide.tsx src/game/Hand.tsx src/game/ActionBar.tsx
git commit -m "feat(game): selection, hover, and target feedback (checkpoint 3)"
```

---

## Task 8: HP/energy change feedback — CHECKPOINT 4

**Files:**
- Create: `src/game/diff.ts`
- Modify: `src/game/PlayerSide.tsx`, `src/game/GameView.tsx`

**Interfaces:**
- Consumes: `GameView`, `InPlayView` (Task 1).
- Produces: `changedSlots(prev: GameView | null, next: GameView): ChangeSet` where `ChangeSet = { hp: Set<string>; energy: Set<string> }` keyed by `"<playerIndex>:<slot>"` (slot = `"active"` or bench index as string). A slot appears in `hp`/`energy` when that value differs from `prev` (or when `prev` is null → no changes).

This is the one piece of pure logic in the slice. It is verified by eyeball (Option 1: no test runner) — keep it small and obvious.

- [ ] **Step 1: Write the diff helper.** Create `src/game/diff.ts`:

```ts
import type { GameView, InPlayView } from './types';

export type ChangeSet = { hp: Set<string>; energy: Set<string> };

function slotKey(player: number, slot: string): string {
  return `${player}:${slot}`;
}

// Which on-board slots changed HP or total energy between two views. A null prev
// (fresh load / new game) reports no changes so nothing flashes on first render.
export function changedSlots(prev: GameView | null, next: GameView): ChangeSet {
  const hp = new Set<string>();
  const energy = new Set<string>();
  if (!prev) return { hp, energy };

  const compare = (p: number, slot: string, a: InPlayView | null, b: InPlayView | null) => {
    if (!a || !b) return;
    if (a.remainingHp !== b.remainingHp) hp.add(slotKey(p, slot));
    if (a.totalEnergy !== b.totalEnergy) energy.add(slotKey(p, slot));
  };

  for (let p = 0; p < 2; p++) {
    compare(p, 'active', prev.players[p].active, next.players[p].active);
    const pb = prev.players[p].bench;
    const nb = next.players[p].bench;
    for (let i = 0; i < nb.length; i++) compare(p, String(i), pb[i] ?? null, nb[i]);
  }
  return { hp, energy };
}
```

- [ ] **Step 2: Thread the change set through.** In `GameView.tsx`, read `prev` from the store, compute `const changes = changedSlots(prev, view);`, and pass the relevant keys to each `PlayerSide` (add prop `changes: ChangeSet` and `playerIndex: number`). In `PlayerSide.tsx`, when rendering the HP line or energy line, if its slot key is in `changes.hp`/`changes.energy`, wrap the number in a span with a brief highlight background (e.g. `background:'#ffce54'`, `borderRadius:4`, `padding:'0 3px'`). The highlight reflects the *current* state that just changed — it is the state fact, shown instantly (ADR-0004), not an animation.

- [ ] **Step 3: Typecheck.** Run: `npm run lint`. Expected: PASS.

- [ ] **Step 4: Browser verify.** Attack to knock energy/HP around: the HP number that dropped and the energy count that rose are highlighted on the render right after the move; the next move recomputes against the new `prev`. New game / first load flashes nothing. Expected: changes are visible without any blocking animation.

- [ ] **Step 5: Commit.**

```bash
git add src/game/diff.ts src/game/GameView.tsx src/game/PlayerSide.tsx
git commit -m "feat(game): highlight HP/energy changes (checkpoint 4)"
```

---

## Self-Review

**Spec coverage** (against `docs/gameplay-ui-spec.md`):
- Play a full game in React with real decks/art → Tasks 5–6. ✅
- Pure renderer + dispatcher, server-authoritative, no TS rules → Tasks 2–3, Global Constraints. ✅
- Lives at `?play`, draft app untouched → Task 5 Step 4. ✅
- Vite proxy dev wiring → Task 1. ✅
- Zustand store per `draftStore` conventions → Task 3. ✅
- Components mirror the view model (PlayerSide/BoardCenter/Hand/ActionBar/Log) → Tasks 5–6. BoardCenter folded into GameView's header (noted). ✅
- Card art reuse "where practical" → Task 4 renders `CardView` directly; documented divergence from `CardTile`. ✅
- Micro-feedback: selection/hover/target + HP/energy change highlights, inside ADR-0004 → Tasks 7–8. ✅
- Pending prompts (setup, new_active) → Task 6. ✅
- Attacks disabled when unaffordable with reason; illegal move shows server reason → Tasks 6–7. ✅
- Opponent hand as count → Task 5 PlayerSide. ✅
- Out of scope (TS rules, effects, multiplayer, replay, animation) → not built. ✅
- Deck selection fixed to the two curated decks → nothing added; server default. ✅

**Placeholder scan:** every code step contains complete code; visual styling is concrete inline styles, not "add styling here." No TBD/TODO. The only "eyeball" verification (diff helper) is an explicit consequence of Option 1, not a gap.

**Type consistency:** `Move` union field names (`activeCardId`, `benchCardIds`, `handCardId`, `benchIndex`, `index`, `target`, `cardId`) match `moveReq` JSON tags in `server/main.go`. `Selection` kinds (`hand`/`bench`/`setupActive`) used consistently across store, Hand, ActionBar, PlayerSide. `changedSlots`/`ChangeSet` names match between `diff.ts` and its consumers. Store action names (`init`, `restart`, `dispatch`, `runBot`, `select`, `clearSelection`) consistent across tasks.

**One risk flagged for execution:** the App.tsx insertion point (Task 5 Step 4) must sit after all hooks run — the executor should confirm no hook is skipped by the early `return <GameView />`.
