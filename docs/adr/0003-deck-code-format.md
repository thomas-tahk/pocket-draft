# Deck codes are the canonical decklist format, shared with constructed play

A drafted deck can be exported as a compact, copy-pasteable **deck code** (and an equivalent shareable link). The code encodes only the final 20-card decklist as `(card id, count)` pairs, base64-encoded with a leading version marker. It is fully client-side — no backend, storage, or accounts.

## Why

Card IDs (e.g. `A1-001`) are stable, so encoding IDs survives card-data refreshes. A version marker lets the format evolve without breaking codes already shared.

Crucially, a *constructed* deck (built outside the draft flow, to play on the simulator) is also just a decklist. So this same deck-code format is intended to be **reused** as the constructed-deck import format later — paste a code, get a playable deck. Designing one canonical, versioned format now avoids inventing a second one for the simulator.

## Consequences

- The encode/decode logic and the format spec should live in a shared module, not buried in the deckbuild UI, so the future simulator can import the same codes.
- Only the final deck is captured — not the draft picks or shop purchases. Sharing a *draft session* is a separate, deferred concern.
- **Drafted decks are export-only.** The draft flow never accepts a code as input — injecting a deck would defeat the randomized pulls and pick decisions that are the point of drafting. Export exists to *show* a drafted deck to others, nothing more.
- **Opening a shared link/code shows a read-only deck view (display), not an import.** Displaying a deck is distinct from loading one into a buildable/playable flow.
- **Import — loading a code into a deck you can build or play — belongs exclusively to the future constructed mode**, and is deferred. `decodeDeck` is built now (the read-only viewer needs it) and reused there later.
- Delivery is both a raw code (copy/paste) and a shareable link (`?deck=…` opens the read-only view).
