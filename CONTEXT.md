# Pocket Draft

A tool for drafting Pokémon TCG Pocket decks and then **simulating** the game's ruleset so those decks can be played out and reviewed. It reproduces Pocket's *rules* (the legal sequence and effect of card play), not the mobile game's visual presentation.

## Language

**Simulator**:
The component that plays out a game by enforcing Pocket's ruleset. The product's core value is rule fidelity and reviewability, not presentation.
_Avoid_: Game, engine-as-product, clone

**Draft play**:
A match played with decks built through the draft → shop → deckbuild flow. The first kind of play to support.
_Avoid_: Arena match (ambiguous)

**Constructed**:
A match played with decks built outside the draft flow (a deck source, not a different ruleset). Supported after draft play.
_Avoid_: Standard, freeform

**Lobby match**:
A networked game between two users — either by open matchmaking or a private room. Covers both draft play and constructed.
_Avoid_: Online game, PvP

**Hotseat**:
Local play where both seats are driven on one machine, with no networking. Used to exercise the simulator before lobby matches exist.
_Avoid_: Local mode, single-player

**Replay**:
A recorded match that can be stepped through and reviewed after the fact. Reviewing drafts is a first-class goal, not a later add-on.
_Avoid_: Log viewer, history

**Deck code**:
A compact, copy-pasteable text string encoding a 20-card decklist as (card id, count) pairs. The canonical decklist format — used to share drafted decks now, and to load constructed decks into the simulator later.
_Avoid_: Deck string, export blob, hash
