# Pocket Draft — Vision & Intent

_Captured 2026-07-12. This is the "why" and the destination. Architecture and
build-order live in the ADRs (`docs/adr/`); this document is what they serve._

## What this is

Building the actual **TCG Pocket battle game, played for real — a better version
of it.** Not a tech demo, not a throwaway engine slice. When this is done, you
draft a deck and play a full, faithful game with it.

## Two aims, one loop

1. **Draft mode the real game doesn't have** — draft a deck from a shared pool,
   then play *that* deck.
2. **Faithful core battle gameplay** — including Pokémon **attack effects and
   abilities**, not just base-damage attacks.

These are not separate features. They are one loop: **draft → play your drafted
deck.** The whole thing becomes real the moment those two meet.

## Why (the motivations — these drive priorities)

- **Equal access.** The real game locks cards behind paying for hundreds/thousands
  of packs. A shared draft pool means everyone starts equal, zero grind.
- **Zero-lag play.** The real game round-trips to its server on every little tap,
  causing constant lag and frustrating UX. Play here must feel instant.
- **Always-on ranked.** The real game gives ~2 weeks of ranked per ~4-week
  expansion cycle. Here, ranked/ladder is always available.

## In scope / Out of scope

**In:** full battle rules — attacks, **attack effects**, **abilities**,
Trainer/Supporter/Item cards, status conditions, coin flips, weakness / retreat /
EX rules, evolution lines; draft mode; eventually PvP + an always-on ladder.

**Out:** collectibility (packs, economy, rarities — shiny / rainbow / full-art),
animations / visual flair, trade, and **official deck-construction enforcement**
(no 2-copy / 20-card gatekeeping — see Ethos).

> **Ethos — this is a personal sandbox.** The digital equivalent of paper proxies:
> play any deck you want, the tool never gates you on official rules. Draft mode
> is an *optional* fun constraint you opt into, not a cage.

> Effects and abilities being "incremental" (ADR-0002) is about **build order**,
> not scope. They are core to playing the game and are not optional.

## Where we are vs. where we're going

**Built and real:**
- Go rules engine skeleton — turn loop, energy zone, base-damage attacks, KO /
  points, setup, retreat, evolve, poison / burn. Tested, green.
- Seven ADRs locking architecture.
- Real-card data bridge (`server/carddata.go`) loading scraped card data.
- The mature draft tool (deployed on Vercel).
- React gameplay board mounted at `?play` (PR #1).

**Not built yet (the real work ahead):**
- **Effects / ability engine** — the bulk. Every card's text actually doing
  something. Incremental, bounded by the draftable card pool.
- **Playing your *chosen* deck.** Today the sim uses a hardcoded preset
  (Entei ex×6, Torkoal×7, Heatran×7) and ignores your draft — no legality issue,
  just the wrong deck.
- **Learnable play UX.** Today's board is a debug harness, not a designed game.
- **Low-latency deploy architecture** (see below).

## Key architecture decision pending

The sim is server-authoritative: every move is a round-trip to the Go server.
On localhost that's instant, but **deployed off localhost it reproduces the exact
per-tap lag this project exists to escape.**

**Leading candidate:** compile the Go engine to **WebAssembly** and run it in the
browser — play is fully local and instant; the server runs the *same* Go code
only as the authority for real-time PvP. One rules codebase, no input lag for
solo / vs-bot, network only where genuinely required. This supersedes ADR-0006's
"reimplement the rules in TypeScript too" plan (same responsiveness, half the
code, no engine drift). To be confirmed before multiplayer.

## Agreed first slice

**Wire the chosen/drafted deck into the sim** so you play a real deck instead of
the hardcoded preset. **No legality checks** — the engine plays whatever deck it's
handed. The smallest step that makes the draft→play loop real. Local either way,
so the WASM decision doesn't block it.
