# Incremental card coverage, gated by a playable-card set

The simulator targets full coverage of every card eventually, but gets there **incrementally** rather than all at once. The draft pool used for *playable* games is restricted to the set of cards the engine currently implements; that set widens over time until it is the whole game.

## Why

The rules engine splits into two layers:
- **Core** — turns, energy, damage, knockouts, win condition, board zones. Card-agnostic; built once, up front.
- **Per-card effects** — what each card's special text does. The long tail (~1500 cards).

Most cards are "vanilla" (an attack that only deals damage). They carry no special text, so the core layer plays them with no per-card code — they are just data (energy cost + damage). That means a large fraction of the pool is playable the moment the core exists. Only cards with special text ("draw 2", "heal 30", "flip a coin…") need bespoke effect logic, and those effects are reusable building blocks shared across many cards.

Building the full per-card layer before anything is playable would delay a working, reviewable game indefinitely. Incremental coverage lets a real draft be played and reviewed early.

## Consequences

- This is **not throwaway work**: the incremental build *is* the full version, partway done. The same core and effect building-blocks serve both the early slice and the eventual full coverage.
- The draft tool gains a "playable" filter (alongside its existing pool filters in `loader.ts`) so it only offers cards the engine can actually play. A reader seeing a smaller-than-full draft pool should look here for why.
- "What's currently playable" is a moving target that grows as effects are implemented; it is expected to lag full card coverage until the long tail is finished.
