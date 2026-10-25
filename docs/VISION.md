# Pocket Draft — Vision & Intent

_Rewritten 2026-09-18, from the owner directly. The previous version framed this
as "two aims, one loop" with drafting as half the purpose. That framing was
wrong, and it propagated — agents read it and reproduced the error. Architecture
lives in `docs/adr/`. Build order lives in the milestone specs. This is the thing
they serve._

## What this is

**Play the battles of Pokémon TCG Pocket, without the bullshit.**

That's the whole purpose, and it is defined by subtraction. The battles are good.
What surrounds them is not:

| Removed | Why |
|---|---|
| **Collecting** | Cards gated behind packs, money, and grind |
| **Per-tap lag** | The real game round-trips to its server over every small interaction |
| **Long animations** | Tolerable once, excruciating across fifty games |
| **Ladder downtime** | Ranked runs roughly 2 weeks in every 4-week cycle |
| **Missing modes** | Drafting most of all — but not only drafting |

Every item there is friction that compounds with volume. This is built for
someone who intends to play *a lot* of games. That is the thread connecting all
five, and it is the tiebreaker when a design question is close.

**Also out of scope:** rarities (shiny / rainbow / full-art), trade, economy.
Those serve the business model, not the game.

**The one thing that makes it real rather than a demo:** cards doing what their
text says. Attack effects, abilities, Trainers, status. A game where only base
damage works is not the game. Effects being built incrementally (ADR-0002) is a
statement about build order, never about scope.

## Multiplayer is the point

This is a multiplayer game with solo surfaces — not a solo game that acquires
multiplayer later. Battles are meant to be against people.

Two surfaces are solo, permanently and by choice:

- **Drafting.** You draft alone against the pool. Not a group opening packs and
  passing them around a table — that belongs to paper card games.
- **Replay review.** Looking back over your own games.

So the multiplayer requirement is **battles only**. That is a far smaller surface
than "the app is multiplayer," and it should stay that way.

## The modes

Three modes of play, plus replay across all of them.

| Mode | Deck comes from | Wrapped by | People |
|---|---|---|---|
| **Ranked** | constructed, full pool | ladder + rating, always on | PvP, always |
| **Draft** | a solo draft session | a run (3 losses / 7 wins) | solo draft, then battle |
| **Custom** | constructed, or given | a preset's rules, or nothing | solo or invited |

**Ranked is PvP without exception.** A rating against a bot means nothing, so no
interim single-player version of Ranked exists or can be faked. Ranked ships
after rooms and transport, or it does not ship.

**VS Bot is a separate concept, not a Ranked substitute.** Its purpose is
practising a deck and building confidence in it. The bot is expected to be
braindead next to a person, and that is fine — it needs to play legally and not
be a wall. Bot intelligence is explicitly not worth investment.

**Casual mode** is likely and deliberately unspecified. See open questions.

### Custom is probably the substrate, not a third mode

*Proposed, not locked — this is a reading of the owner's intent that has not yet
been ratified.*

Custom is a normal game of Pocket with adjustable match and format settings, and
fixed combinations of those settings get promoted into named modes over time.
Follow that one step and Ranked is itself a preset — standard format, ladder
attached — and Draft is a preset plus a different way of getting a deck.

If that holds, only three things ever vary:

- **where the deck comes from** — drafted, constructed, or inherited from a
  replayed game
- **where the starting state comes from** — a fresh setup, or a prefix of a
  recorded game's event log
- **what wraps the game** — a ladder, a run, a preset's rules, or nothing

The engine never changes across any of it. "Add a mode" becomes a config entry
rather than a project.

**Caveat, held deliberately:** decide the *model* now, generalise the *code*
later. A settings engine built before anyone knows which knobs matter is
premature abstraction. Build one concrete mode end to end; leave the seam
visible.

## Replay, and playing from it

Every mode emits an event log, because the engine is event-sourced and
deterministic (ADR-0005). Replay is re-applying that log.

**Play-from-replay** is the feature worth naming: branch at any point in a
recorded game and keep playing from there — alone, to explore a different line,
or with someone invited in to see how it could have gone. Architecturally this is
not a new system. It is loading a game from a *prefix* of a log rather than from
a fresh setup. The July 2026 decision to make the engine event-sourced paid for
this in advance.

One fork inside it, cheap to build either way and not yet decided:

- **Study** — reuse the original seed, so draws stay identical and your decisions
  are the only variable. Answers "was that attack the mistake?"
- **What-if** — re-roll from the branch point, so luck varies too. Answers "how
  else could this have gone?"

They feel identical to a player and answer different questions.

## What this implies architecturally

Recorded here because these follow from the vision rather than from taste. The
detail belongs in ADRs.

- **Rooms are the primitive, not a lobby.** Four things need two clients in one
  game: Ranked matchmaking, Custom invites, shared replay branching, and any
  future battle mode. Build a joinable game instance with an id; matchmaking and
  friend-invites are then policies on top of it. A "lobby system" designed as a
  feature tends to swallow rating, chat and friends lists.
- **WebAssembly implements ADR-0006, it does not supersede it.** The previous
  vision claimed WASM replaced client-side prediction. That holds only if solo
  play is primary. It is not. Two people sharing a game means the server is the
  authority and the round-trip is unavoidable — WASM cannot remove latency the
  design requires. You beat that with prediction: play locally and instantly,
  reconcile with the server. WASM is simply the best way to *implement* that
  predictor, because it is the identical Go engine rather than a TypeScript
  reimplementation that drifts.
- **The engine was the right thing to build first.** Two people connected to a
  server that does not know the rules is nothing. Event sourcing, determinism and
  seeding are mild overkill for solo play and mandatory for netcode and replay.
- **The solo assumption shaped only the shell.** One global in-memory game,
  request/response HTTP instead of a connection, two separate boards. All of it
  sits outside the rules and is cheap to replace.

## Open questions

Real gaps, recorded rather than guessed at.

1. **What are the Custom sliders, concretely?** Three or four examples decide
   whether Custom means format variants (deck size, copy limit, points to win) or
   something wider (starting energy, HP scaling, ban lists). Very different
   amounts of work.
2. **Constructed deck-building does not exist.** Decks come from drafting only
   today. Ranked and Custom both need "build 20 from the whole pool," and that
   surface has never been scoped. Two of three modes depend on it.
3. **Where do replays live?** Per-browser local storage is free and private.
   Ranked replays you would probably want kept and shareable, which means server
   storage and running cost.
4. **Study or what-if** for play-from-replay, or both.
5. **Casual mode** — whether it is a distinct mode or just Custom's default
   preset.

## Where this is written down

- `docs/adr/` — locked architecture decisions.
- `docs/superpowers/specs/2026-07-12-draft-mode-full-flow-design.md` — the draft
  milestone. Predates this document; its "constructed mode" conflates what are
  now Ranked and Custom, and its deferral of rooms ("no rooms needed until
  multiplayer") is the line everything above crosses.
- `docs/decisions/2026-10-15-rework-answers.md` — the owner's answers on the battle
  rework, the real Pocket rules, and the draft changes. The battle spec is built from it.
