# Client-side prediction — responsive by default, wait only on hidden information

The client predicts the outcome of any action it has enough information to compute, showing the result **instantly**, then reconciles with the server's authoritative answer. Server round-trips are felt by the player *only* for genuinely unknowable outcomes (drawing, coin flips, the opponent's choices). This is a product cornerstone — "responsive, performant actions in and out of a match" — and the deliberate opposite of the real game's round-trip-everything lag.

## Why

Server-authoritative (ADR-0001) requires the server to have the *final say*, not that the client *wait* before showing a result. The bridge is the key insight: **prediction is possible exactly when there is no hidden information.** Your own client knows your hand, board, and the rules, so it can compute a no-hidden-info action's result with certainty — and it will match the server because both run the same rules on the same known inputs.

- **Predictable (no hidden info):** play a known card, attach energy, retreat, evolve, end turn, reject an illegal move. Shown instantly; server confirmation is a silent rubber stamp.
- **Unknowable (hidden info / other player's choice):** draw, coin flip, opponent actions, a prompt addressed to the *other* player. The client must wait — and these are the *only* delays a player feels. They are intrinsic to a hidden-information card game, not artificial.

## Consequences

- **The rules logic exists in two languages:** Go (server, authoritative) and TypeScript (client, predictive). This duplication risks drift; it must be managed (rules-as-data, a shared conformance test suite, or eventually one engine compiled to WASM). This is the real cost of prediction and the main argument against it — accepted for the responsiveness payoff.
- **Reconciliation/rollback** is required when a prediction is wrong. Event sourcing (ADR-0005) makes it cheap: discard optimistic events, re-fold the authoritative log. Turn-based play makes it rare and visually mild (a card snaps back) — unlike real-time games where rollback is hellish.
- **The predict-vs-wait line is the same line as the choice-ownership seam** in ADR-0007: actor-owned choices are predictable; other-owned choices (prompts) are not.
- **Deferred until networked play.** Single-machine hotseat (the first build) needs no prediction and no TS engine — prediction is introduced when lobby matches are.
