# Presentation shows state, never theater; feedback is always instant

The client renders the **facts** of the game and nothing more. Every visual must earn its place by conveying game state; anything that exists only for drama is cut. This is a cornerstone of the product, not a polish-phase preference — it is how the simulator deliberately departs from the real game's experience (and stays clearly a rules-and-review tool, not a clone of it).

The rule, applied to any card or interaction — including ones not yet anticipated:

1. **Show the state.** For anything that happens, ask "what *fact* about the game changed?" and show exactly that, in the plainest form. Cards in hand (the count), a card moving hand→board, a coin landing heads, damage applied — all state, all shown.
2. **Cut the theater.** Add nothing on top of the state change. No attack cinematics, no card-play flourishes, no spinning coin — the coin shows *heads/tails*, not a spin.
3. **Never block; only refuse.** Visuals are instant and never something the player waits through. The game never makes you *wait* to act — but it does instantly *refuse* an illegal action (e.g. a second Trainer in one turn), which is itself just showing a state fact ("already played a Trainer this turn").

"Minimal animation" is the *appearance* of this rule in practice, not the rule itself. The rule is a test that scales to ~1500 cards no one has enumerated: meet a new effect, ask what state changed, show that, add nothing, never make the player wait.

## Why

- **Latency.** The real game stacks slow, *blocking* animations on top of every network round-trip; players tolerate it because it's Pokémon. Cutting theater removes a whole class of artificial delay. (It also raises the stakes on client-side prediction — a no-theater game has nowhere to hide a stall, so predictable actions genuinely must be instant.)
- **Differentiation / IP.** A plain rules-and-review tool is unmistakably not a clone of the mobile game. CONTEXT.md already frames the Simulator as "rule fidelity and reviewability, **not presentation**"; this ADR is the concrete, enforceable form of that stance.

## Consequences

- **No animation system gets built.** A future contributor *will* be tempted to "helpfully" add flourish — this ADR is the documented "no," with the test to apply instead.
- **Two of animation's jobs need replacing.** Animation normally provides (a) latency cover and (b) narration — letting you follow *what just happened*. (a) is replaced by client-side prediction; (b) is replaced by a live, on-screen projection of the event log (e.g. "P2 played Pikachu → attached Lightning → attacked for 20") plus plain state display, so the player is never lost despite zero theater.
- **"Block" is two different things, kept distinct.** *Block-the-wait* (theater the player sits through) is banned; *block-the-illegal* (refusing a rule-breaking action) is required and must itself be instant and informative.
- **Illegal-move rejection is local first.** Because the client owns a copy of the rules, it refuses obvious illegal actions instantly with no round-trip; the server re-validates for anti-cheat. Speed on the client, truth on the server.
