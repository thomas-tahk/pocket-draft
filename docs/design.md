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

### Source: scrape Limitless TCG directly
- `pocket.limitlesstcg.com` is fully permissive in robots.txt
- Build-time scraper (`scripts/scrape-limitless.mjs`, Node + cheerio) fetches `/cards`, `/cards/{SETID}`, and `/cards/{SETID}/{NUM}` and emits `data/cards.json` + `data/sets.json`
- Polite: 250 ms between requests, custom UA, sequential fetch, full re-run ~10 min
- Output is vendored into the repo and copied into `public/data/` for runtime serving
- Re-run when new sets drop; output is idempotent

### Card schema (per card)
- Identity: `id` (e.g. `A1-001`), `setId`, `number`, `name`
- Rarity model: 4 raw rarities (◊/◊◊/◊◊◊/◊◊◊◊) plus cosmetic stars (☆/☆☆/☆☆☆/♕) plus `Promo`; orthogonal `exKind` (regular / ex / mega-ex). Mega-EX is detected via `name.startsWith('Mega ') && name.endsWith(' ex')`
- Mechanical fields: `cardType` (energy type or `Trainer`), `trainerKind` (Item/Supporter), `hp`, `stage`, `evolvesFrom`, `attacks[]` (cost/name/damage/effect), `ability` (name/effect), `trainerText` (Trainer effect, null for Pokémon), `weakness`, `retreat`, `pack`
- Visual fields: `imageThumb`, `imageFull`, `artist`, `flavor`

### Draft pool filtering (at load time)
Applied by `src/data/loader.ts`:
- Drop cosmetic variants (☆/☆☆/☆☆☆/♕)
- Drop the entire `A4b` set (reprint-only "Deluxe Pack: ex" — every base-art card has a functional duplicate)
- Drop the 5 free-universal trainers (Potion, X Speed, Red Card, Poké Ball, Professor's Research) — they're available freely at deckbuild
- For Promo cards, infer draft rarity from `exKind` (EX/Mega-EX → ◊◊◊◊, regular → ◊)
- Dedupe the `pa-064` / `pa-065` Rayquaza ex pair (same functional signature) — keep lowest id
- No general functional dedup: multi-pack cards (e.g. A1 Eevee in three packs) are kept as separate entries for pack-composition tables

### Image source: Limitless TCG CDN (hot-linked)
- Thumbnail (`imageThumb`): `https://limitlesstcg.nyc3.cdn.digitaloceanspaces.com/pocket/{SET}/{SET}_{NUM:003}_EN_SM.webp`
- Full-res (`imageFull`): same path without `_SM` and `.png`
- Smoke test (v0.0) confirmed images load reliably from the CDN

### Sources we rejected
- **chase-manning/pokemon-tcg-pocket-cards** — was an early pivot; rejected 2026-05-18 because metadata is incomplete (no attack text, abilities, weakness, retreat, stage). We now own the scraper.
- **TCGdex** — `api.tcgdex.net` was unreachable from our network during validation; Pocket coverage on TCGdex is also weaker than on Limitless
- **pokemontcg.io** — does not cover TCG Pocket (physical TCG only)

## Open Questions
- Backend introduction timing: Phase 1 v0.1 is pure frontend (no backend); backend joins for v0.2 (matchmaking)
- Draft format rules: see docs/rules.md (finalized)
- Go web framework: Gin vs Echo — decide when starting backend work
- Type sharing strategy: manual sync for now, revisit if drift becomes a problem
- Deployment target: TBD
