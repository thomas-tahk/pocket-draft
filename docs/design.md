# Design & Architecture

## Project Overview

An online tool for conducting/simulating a drafting format for Pokemon TCG Pocket (the mobile game), with a stretch goal of full gameplay simulation and replay functionality.

## Core Features

### Phase 1 — Draft Tool
- **Arena-style format:** players draft a single deck by repeatedly picking 1 card from a set of offered options
- No pre-existing collection required — any card in the game is draftable; basic universals (Pokeball, Prof's Research, etc.) are always freely available without drafting
- Solitary, asynchronous draft — each player picks at their own pace, no timer, no inter-player dependency
- Matchmaking happens post-draft (after both players complete their picks)
- After drafting, each player builds one 20-card deck and plays a run (win/loss threshold TBD)
- Conquest format (3 decks, must win with each) is a future consideration for tournament play

### Phase 2 — Gameplay Simulation
- Simulate a full Pokemon TCG Pocket game between two decks
- No animations — static card images, clean readable board state
- No inherent delays (faster than the actual mobile game)
- Timed turns (countdown per player action)

### Phase 3 — Replay
- Export/import game logs
- Step through turns and game states in the app
- Share replays via URL or file

## Tech Stack

### Backend — Go
- Phase 1 v0.1 (async solitary draft) does NOT need a backend — pure frontend with state in localStorage. Backend joins at v0.2 for matchmaking and persistence.
- Go is chosen for Phase 2 (gameplay simulation): goroutine-based concurrency maps naturally to per-game state with turn timers and concurrent matches.
- WebSockets only needed for Phase 2 (turn-by-turn gameplay events). Phase 1/v0.2 backend uses plain HTTP/REST.
- Framework TBD (Gin or Echo) — decide when starting backend work
- WebSocket library TBD (gorilla/websocket or nhooyr.io/websocket) — decide at Phase 2

### Frontend — React + TypeScript
- Vite for build tooling
- Zustand for client-side state management
- React Query for card data fetching
- WebSocket client for real-time communication with backend

### Monorepo Structure
```
pocket-draft/
  server/           # Go backend
    cmd/
      main.go
    internal/
      draft/        # draft room logic, timers, pack opening simulation
      game/         # gameplay simulation engine
      api/          # HTTP handlers
      hub/          # WebSocket connection hub
  client/           # React + TypeScript frontend
    src/
      components/
      hooks/
      stores/
      types/
  docs/             # rules, design notes, API references
  Makefile          # unified commands to run/test both sides
```

### Type Sharing (Go ↔ TypeScript)
- Start with manual sync (Go structs + matching TS interfaces)
- Consider `tygo` for code generation if drift becomes a problem

## Card Data

### Metadata source: chase-manning/pokemon-tcg-pocket-cards (vendored)
- MIT-licensed GitHub repo that scrapes Limitless TCG and publishes structured JSON
- We vendor `v4.json` (all cards) + `expansions.json` (set/pack metadata) into `data/cards/` in this repo
- Refresh script pulls latest from upstream when new sets drop
- Derived transformations applied at load time: normalized 5-tier rarity, functional dedup of cosmetic variants, pack composition tables

### Image source: Limitless TCG CDN (hot-linked)
- Thumbnail: `https://limitlesstcg.nyc3.cdn.digitaloceanspaces.com/pocket/{SET}/{SET}_{NUM:003}_EN_SM.webp`
- Full-resolution (for zoom view): drop the `_SM` suffix and switch `.webp` → `.png`
- robots.txt is fully permissive; smoke test confirmed images load reliably
- Can switch to chase-manning's bundled image copies later if hot-linking ever becomes a problem

### Sources we rejected
- **TCGdex** — `api.tcgdex.net` was unreachable from our network during validation; Pocket coverage on TCGdex is also weaker than on Limitless
- **pokemontcg.io** — does not cover TCG Pocket (physical TCG only)

## Open Questions
- Backend introduction timing: Phase 1 v0.1 is pure frontend (no backend); backend joins for v0.2 (matchmaking)
- Draft format rules: see docs/rules.md (finalized)
- Go web framework: Gin vs Echo — decide when starting backend work
- Type sharing strategy: manual sync for now, revisit if drift becomes a problem
- Deployment target: TBD
