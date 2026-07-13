# Play Your Drafted Deck — Implementation Plan (Phases 0 + 1)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax.

**Goal:** From the "Deck finalized" review screen, click **Play deck** to start a game that uses your finalized 20-card deck against a bot; exit returns to review. Deck construction is enforced server-side (exactly 20 cards, ≤2 copies by name, ≥1 Basic).

**Architecture:** The Go engine's `NewGame(seed, deck0, deck1)` already takes two decks; the whole change is (a) let the server *receive* a deck of card IDs, validate it, and build player-0 from it (opponent = the existing curated preset for now), and (b) un-silo the React shell so the game is an in-app view launched from Review with the finalized deck, not a dead-end `?play` route.

**Tech Stack:** Go stdlib HTTP server (`server/`), Go engine (`engine/`), React 18 + TS + Vite + zustand (`src/`).

## Global Constraints

- Deck format enforced at the server: **exactly 20 cards**, **≤2 copies of any card by name**, **≥1 Basic Pokémon**. Malformed decks are rejected; game state unchanged.
- Card *access* stays open — any card ID is allowed; only the format is enforced.
- No new dependencies. Go: stdlib only. Frontend: no JS test runner exists — frontend tasks verify via `npx tsc --noEmit` + manual browser check (project convention).
- Opponent deck for this slice = the existing curated Fire preset (`fireDeck`), which is guaranteed playable. "Bot drafts its own deck from the pool" is deferred to a later phase.
- Commit after each task. End every commit message with `Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>`.

---

## Task 1: `deckFromIDs` — build + validate a deck from card IDs (server)

**Files:**
- Modify: `server/carddata.go` (add function near `buildDeck`)
- Test: `server/carddata_test.go` (add cases)

**Interfaces:**
- Produces: `func deckFromIDs(ids []string) ([]engine.Card, error)` — expands card IDs to engine cards via `toEngineCard`, enforcing 20 / ≤2-by-name / ≥1-Basic.

- [ ] **Step 1: Write failing tests** in `server/carddata_test.go`. Populate `rawByID` with a small fixture set, then assert:

```go
func TestDeckFromIDs(t *testing.T) {
	// Arrange: a fixture pool — 1 Basic + 1 Stage2, both Fire.
	rawByID = map[string]rawCard{
		"BASIC": {ID: "BASIC", Name: "Basicmon", CardType: "Fire", Stage: "Basic", HP: 60,
			Attacks: []rawAttack{{Cost: "R", Name: "Hit", Damage: "20"}}},
		"S2": {ID: "S2", Name: "Finalmon", CardType: "Fire", Stage: "Stage 2", HP: 150,
			Attacks: []rawAttack{{Cost: "RR", Name: "Big", Damage: "90"}}},
	}
	rep := func(id string, n int) []string {
		out := make([]string, n)
		for i := range out { out[i] = id }
		return out
	}

	// Valid: 18 basics + 2 finals = 20, ≤2 by name? No — 18 same name breaks ≤2.
	// Build a legal 20 by alternating is impossible with 2 names; instead test the
	// rules directly:

	// exactly-20 rule
	if _, err := deckFromIDs(rep("BASIC", 19)); err == nil {
		t.Error("expected error for 19-card deck")
	}
	// ≤2-by-name rule
	if _, err := deckFromIDs(append(rep("BASIC", 3), rep("S2", 17)...)); err == nil {
		t.Error("expected error for 3 copies of a name")
	}
	// ≥1-Basic rule (20 Stage2s, but that's 20 copies of one name — also illegal;
	// use two distinct all-Stage2 names capped at 2 each is impossible at 20, so
	// this rule is covered structurally: a legal 20 with 0 basics needs ≥10 names).
	// unknown id
	if _, err := deckFromIDs(rep("NOPE", 20)); err == nil {
		t.Error("expected error for unknown card id")
	}
}
```

> Note: a strictly-legal 20-card fixture needs ≥10 distinct names (≤2 each). To keep the fixture small, add distinct dummy Basics in the test setup (e.g. a loop creating "B0".."B9" Basics, 2 each = 20) for the happy-path assertion. Include that happy-path case: `deckFromIDs(twentyLegalIDs)` returns 20 cards, no error.

- [ ] **Step 2: Run tests, verify they fail.** `cd server && go test ./... -run TestDeckFromIDs -v` → FAIL (undefined: deckFromIDs).

- [ ] **Step 3: Implement `deckFromIDs`** in `server/carddata.go`:

```go
// deckFromIDs expands a list of card IDs into engine cards, enforcing the deck
// format: exactly 20 cards, at most 2 copies of any card by name, and at least
// one Basic Pokémon. Card access is unrestricted — only the format is checked.
func deckFromIDs(ids []string) ([]engine.Card, error) {
	if len(ids) != 20 {
		return nil, fmt.Errorf("deck must be exactly 20 cards, got %d", len(ids))
	}
	deck := make([]engine.Card, 0, 20)
	countByName := map[string]int{}
	basics := 0
	for _, id := range ids {
		rc, ok := rawByID[id]
		if !ok {
			return nil, fmt.Errorf("unknown card id %q", id)
		}
		card, err := toEngineCard(rc)
		if err != nil {
			return nil, err
		}
		countByName[card.Name]++
		if countByName[card.Name] > 2 {
			return nil, fmt.Errorf("at most 2 copies of %q allowed", card.Name)
		}
		if card.Stage == engine.Basic {
			basics++
		}
		deck = append(deck, card)
	}
	if basics == 0 {
		return nil, fmt.Errorf("deck needs at least 1 Basic Pokémon")
	}
	return deck, nil
}
```

- [ ] **Step 4: Run tests, verify pass.** `cd server && go test ./... -v` → PASS.

- [ ] **Step 5: Commit.** `git add server/carddata.go server/carddata_test.go && git commit` (message: `feat(server): deckFromIDs builds and format-validates a deck`).

---

## Task 2: `/api/new` accepts a deck (server)

**Files:**
- Modify: `server/main.go` (`handleNew`, add `newReq`)

**Interfaces:**
- Consumes: `deckFromIDs` (Task 1), existing `fireDeck`/`waterDeck` globals.
- Produces: `POST /api/new` accepts optional JSON body `{ "you": ["id",...], "seed": N }`. With `you`: player-0 = your deck, player-1 = `fireDeck`; invalid deck → HTTP 400 + `{"error":...}`, game unchanged. Without body: current curated-vs-curated behavior (backward compatible). `?seed=` query still honored.

- [ ] **Step 1: Add `newReq` and rewrite `handleNew`** in `server/main.go`:

```go
type newReq struct {
	You  []string `json:"you"`
	Seed *uint64  `json:"seed"`
}

// POST /api/new — start a fresh game. Optional JSON body {you:[cardId...], seed}
// plays your deck (player 0) against the curated bot deck (player 1). With no
// body it falls back to the curated demo matchup. ?seed=N is also honored.
func handleNew(w http.ResponseWriter, r *http.Request) {
	var req newReq
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req) // empty/invalid body -> no deck
	}

	seed := uint64(1)
	if req.Seed != nil {
		seed = *req.Seed
	} else if s := r.URL.Query().Get("seed"); s != "" {
		if n, err := strconv.ParseUint(s, 10, 64); err == nil {
			seed = n
		}
	}

	p0, p1 := fireDeck, waterDeck
	if len(req.You) > 0 {
		d, err := deckFromIDs(req.You)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			writeJSON(w, map[string]string{"error": err.Error()})
			return
		}
		p0, p1 = d, fireDeck
	}

	mu.Lock()
	defer mu.Unlock()
	game = engine.NewGame(seed, p0, p1)
	writeJSON(w, toGameView(game))
}
```

Also change `newGame(seed)` call sites: the startup `newGame(1)` in `main()` stays, but `newGame` helper is now only used at startup — keep it. (handleNew no longer calls it.)

- [ ] **Step 2: Verify build + existing tests.** `cd server && go build ./... && go test ./...` → PASS.

- [ ] **Step 3: Manual curl check.** Start server (`cd server && go run .`), grab 20 valid card IDs from `public/data/cards.json` (any 10 distinct Basic Fire IDs ×2). POST them and confirm player-0's hand/active are your cards, and a 19-card body returns HTTP 400 with an error. (Command in the verification section below.)

- [ ] **Step 4: Commit.** `git add server/main.go && git commit` (message: `feat(server): /api/new plays your deck vs the bot preset`).

---

## Task 3: `expandDeck` + deck-aware game API (frontend)

**Files:**
- Create: `src/game/deck.ts`
- Modify: `src/game/api.ts` (`newGame`)

**Interfaces:**
- Produces: `expandDeck(deck: Record<string, number>): string[]` — flattens `cardId→count` to a flat id list.
- Produces (changed): `newGame(opts?: { deck?: string[]; seed?: number }): Promise<GameView>`.

- [ ] **Step 1:** Create `src/game/deck.ts`:

```ts
// Flattens a finalized deck (cardId → count) into a flat list of card ids,
// the shape the sim's /api/new expects.
export function expandDeck(deck: Record<string, number>): string[] {
  const ids: string[] = [];
  for (const [id, count] of Object.entries(deck)) {
    for (let i = 0; i < count; i++) ids.push(id);
  }
  return ids;
}
```

- [ ] **Step 2:** Change `newGame` in `src/game/api.ts`:

```ts
export function newGame(opts: { deck?: string[]; seed?: number } = {}): Promise<GameView> {
  const { deck, seed } = opts;
  const body = deck ? JSON.stringify({ you: deck, seed }) : seed !== undefined ? JSON.stringify({ seed }) : undefined;
  return fetch('/api/new', {
    method: 'POST',
    headers: body ? { 'Content-Type': 'application/json' } : undefined,
    body,
  }).then((r) => asJson<GameView>(r));
}
```

- [ ] **Step 3: Typecheck.** `npx tsc --noEmit` → no errors from these files. (Callers updated in Task 4.)

- [ ] **Step 4: Commit.** `git add src/game/deck.ts src/game/api.ts && git commit` (message: `feat(game): expandDeck helper and deck-aware newGame`).

---

## Task 4: Game store starts from a deck (frontend)

**Files:**
- Modify: `src/game/store.ts`

**Interfaces:**
- Consumes: `newGame({deck, seed})` (Task 3).
- Produces: store gains `deck: string[] | null` and `startGame(deck: string[], seed?: number): Promise<void>`; `restart(seed?)` replays the stored deck.

- [ ] **Step 1:** In `src/game/store.ts`, add `deck: string[] | null` to `State` (init `null`), add `startGame` to `Actions`, and update `restart`:

```ts
// in State:  deck: string[] | null;
// in Actions: startGame: (deck: string[], seed?: number) => Promise<void>;

startGame: async (deck, seed) => {
  set({ status: 'loading', error: null, selection: null, deck });
  try {
    const view = await newGame({ deck, seed });
    set({ view, prev: null, status: 'ready' });
  } catch (e) {
    set({ status: 'idle', error: e instanceof Error ? e.message : String(e) });
  }
},

restart: async (seed) => {
  set({ status: 'loading', error: null, selection: null });
  try {
    const view = await newGame({ deck: get().deck ?? undefined, seed });
    set({ view, prev: null, status: 'ready' });
  } catch (e) {
    set({ status: 'idle', error: e instanceof Error ? e.message : String(e) });
  }
},
```

Add `deck: null,` to the store's initial-state object.

- [ ] **Step 2: Typecheck.** `npx tsc --noEmit` → clean.

- [ ] **Step 3: Commit.** `git add src/game/store.ts && git commit` (message: `feat(game): store starts and restarts from a chosen deck`).

---

## Task 5: GameView takes a deck + exit (frontend)

**Files:**
- Modify: `src/game/GameView.tsx`

**Interfaces:**
- Consumes: `startGame`, `init` (store).
- Produces: `GameView({ deck?, onExit? }: { deck?: string[]; onExit?: () => void })`. With `deck` → `startGame(deck)`; without → `init()` (legacy `?play` server game). Renders an Exit button when `onExit` is given.

- [ ] **Step 1:** Update `src/game/GameView.tsx`:

```tsx
export function GameView({ deck, onExit }: { deck?: string[]; onExit?: () => void }) {
  const { view, prev, status, error, init, startGame } = useGameStore();

  useEffect(() => {
    if (deck) void startGame(deck);
    else void init();
  }, [deck, init, startGame]);
  // ...unchanged loading/error block...
  // In the returned <main>, add at the top:
  //   {onExit && <button onClick={onExit} style={{ justifySelf: 'start', fontSize: 12 }}>← Back to deck</button>}
}
```

Keep the rest of the component identical.

- [ ] **Step 2: Typecheck.** `npx tsc --noEmit` → clean.

- [ ] **Step 3: Commit.** `git add src/game/GameView.tsx && git commit` (message: `feat(game): GameView plays a passed deck with an exit back`).

---

## Task 6: Un-silo the shell + "Play deck" from Review (frontend)

**Files:**
- Modify: `src/App.tsx`
- Modify: `src/components/ReviewView.tsx`

**Interfaces:**
- Consumes: `expandDeck` (Task 3), `GameView({deck,onExit})` (Task 5).
- Produces: an in-app `playing` view state; `ReviewView` gains `onPlay: () => void` and a **Play deck** button.

- [ ] **Step 1:** In `src/App.tsx`, add `import { expandDeck } from './game/deck';` and a state hook `const [playing, setPlaying] = useState(false);`. Replace the legacy short-circuit:

```tsx
// legacy quick-test entry: ?play plays the server's current curated game
const isPlayParam = new URLSearchParams(window.location.search).has('play');
if (isPlayParam) return <GameView />;

// in-app deck play, launched from Review
if (playing) return <GameView deck={expandDeck(deck)} onExit={() => setPlaying(false)} />;
```

Pass `onPlay={() => setPlaying(true)}` into the `<ReviewView ... />` at the bottom of `App`.

- [ ] **Step 2:** In `src/components/ReviewView.tsx`, add `onPlay: () => void;` to `Props`, destructure it, and add a prominent button as the first item in the header's button row:

```tsx
<button onClick={onPlay} style={{ fontWeight: 600 }}>▶ Play deck</button>
```

- [ ] **Step 3: Typecheck.** `npx tsc --noEmit` → clean.

- [ ] **Step 4: Manual browser verification** (the real goal): `npm run dev` + `cd server && go run .` in parallel. Finalize a deck (or reuse an existing finalized one) → on "Deck finalized", click **▶ Play deck** → a game starts whose player-0 cards are from *your* deck (not the Entei/Torkoal/Heatran preset) → play a move, then **← Back to deck** returns to Review.

- [ ] **Step 5: Commit.** `git add src/App.tsx src/components/ReviewView.tsx && git commit` (message: `feat(game): play your finalized deck from the review screen`).

---

## Verification (end-to-end)

- **Backend unit:** `cd server && go test ./...` green (Task 1 cases).
- **Backend integration (curl):** with the server running,
  ```bash
  # 19 cards -> rejected
  curl -s -X POST localhost:8080/api/new -H 'Content-Type: application/json' \
    -d '{"you":["A1-001"]}' -o /dev/null -w '%{http_code}\n'   # expect 400
  ```
  and a valid 20-id body returns 200 with player-0 holding your cards.
- **Frontend:** `npx tsc --noEmit` clean; manual browser click-through per Task 6 Step 4.
- **Goal check:** clicking **Play deck** on Review starts a game with your drafted cards vs the bot, and Back returns to Review. That is the Phase 0 + 1 deliverable.

## Deferred (next phases, not this plan)

- Bot drafts its own deck from the pool (currently the curated Fire preset).
- Effect engine (Phase 2), run wrapper + records (Phase 3), board redesign, replay, constructed.
