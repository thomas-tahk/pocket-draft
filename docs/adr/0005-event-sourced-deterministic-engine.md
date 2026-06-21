# The engine is an event-sourced, deterministic state machine

The canonical game is an **append-only log of resolved events** (intents and prompt responses), not a mutable state blob. Current state is *derived* by folding the log from the initial position: `state = fold(initial, events)`. The engine — the per-event transition function — is the simulator core from ADR-0002.

## Why

Three first-class goals fall out for free:

- **Replay** (a stated first-class goal in CONTEXT.md): the log *is* the replay; the board at step N is the fold of the first N events. Stepping is just choosing where to stop.
- **Reconnection**: a rejoining client receives the log (or a snapshot + tail) and re-folds to "now" — the same fold everyone always runs, not a special resync path.
- **Anti-cheat** (the reason ADR-0001 puts the engine server-side): refereeing is one rule — validate an event against current state *before* appending. An illegal event never enters the log, so the log is always a perfect record of a legal game.

## Consequences

- **Determinism is a hard constraint from commit one.** Re-folding the same log must always produce the same state. Every source of randomness (deck shuffle, coin flips, "random" effects) must derive from a **seed stored in the log** — no ambient `rand`/`Math.random()` in effect code, ever, across all ~1500 cards. This is the entire price of event sourcing and it is paid per-effect.
- **The log stores outcomes and decisions, never the live process that produced them.** A coin flip logs "heads," not an animation; a turn timer logs a `Timeout` event, not a countdown. Wall-clock time may only *trigger* an event; once logged, the game is fully determined by the log again.
- **Reconciliation (ADR-0006) is cheap:** correcting a mispredicted client = discard optimistic events, re-fold the authoritative log.
