# Design & Architecture

## Project Overview

An online tool for conducting/simulating a drafting format for Pokemon TCG Pocket (the mobile game), with a stretch goal of full gameplay simulation and replay functionality.

## Core Features

### Phase 1 — Draft Tool
- **Arena-style format:** players draft a single deck by repeatedly picking 1 card from a set of offered options
- No pre-existing collection required — any card in the game is draftable; basic universals (Pokeball, Prof's Research, etc.) are always freely available without drafting
- Timed draft picks (countdown per player turn)
- Online/multiplayer — players draft simultaneously in the same session
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
- Chosen for goroutine-based concurrency model, which maps naturally to:
  - Per-room draft sessions with independent timers
  - Per-game simulation state with turn timers
- WebSockets for real-time events (draft picks, timer ticks, game actions)
- Framework TBD (Gin or Echo)
- WebSocket library TBD (gorilla/websocket or nhooyr.io/websocket)

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

### Source: TCGdex
- Public community-maintained API, covers TCG Pocket sets
- Docs: https://tcgdex.dev
- Card endpoint: `GET https://api.tcgdex.net/v2/en/cards/{setId}-{number}`
- Card images: `https://assets.tcgdex.net/en/tcgp/{setId}/{number}/high.webp`
- Confirmed live and returning card images

### Notes
- pokemontcg.io does NOT cover TCG Pocket (physical TCG only)
- Limitless TCG (pocket.limitlesstcg.com) self-hosts card images on their own DigitalOcean CDN — not a public resource we can use

## Open Questions
- Multiplayer support: deferred — start with single-session / local simulation
- Draft format rules: TBD — see docs/rules.md once sourced
- Go web framework: Gin vs Echo — decide at project init
- Type sharing strategy: manual sync for now, revisit if needed
- Deployment target: TBD
