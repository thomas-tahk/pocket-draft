# Engine interaction model — three modes, and choice ownership decides intent-vs-prompt

The engine's main loop handles three modes, all first-class from the first build:

1. **Free-form intents.** During a player's turn they send atomic actions in any valid order (play Basic, evolve, attach energy, play Item/Supporter, use ability, retreat, attack). The engine silently validates each against the rules and instantly refuses illegal ones (e.g. a second energy attach, a second Supporter, attacking with no energy). This is the normal 95% loop. Attacking ends the turn.
2. **Automatic resolution.** Some rules fire with no player input — end-of-turn burn/poison damage, and any knockout that damage causes. The engine performs and logs these itself between turns.
3. **Prompts (pending decisions).** Occasionally the rules require **one specific player** to make **one specific choice** before play can continue. The engine pauses, addresses the prompt to that player, and accepts only a valid response. Examples: setup (place ≥1 Basic as Active), your Active is knocked out (choose the new Active), Sabrina (the *opponent* chooses their new Active).

## The load-bearing rule: who owns the choice?

A choice lives in one of two places, determined solely by **which player makes it**:

- **Actor-owned → a parameter of the intent.** The acting player decides up front, so the pick travels with the action: `PlayTrainer(Cyrus, target: opponent's damaged Charizard)`, `AttachEnergy(to: Pikachu)`, `Retreat(to: bench slot 2)`. The other player is not asked — they receive the result. (Cyrus *feels* forced to its target, but mechanically the actor chose; it is mode 1, not a prompt.)
- **Other-owned → a prompt (mode 3).** The choice belongs to a player other than the one acting, so the engine pauses and asks them. (Sabrina; knocked-out-Active.)

This single distinction coincides with three others, so the engine implements **one** seam, not three:

| | Actor-owned (Cyrus) | Other-owned (Sabrina) |
|---|---|---|
| Protocol form | intent parameter | prompt + response |
| Prediction (ADR-0006), actor's client | predicts fully | must wait |
| Prediction, other client | renders result | shows the prompt |

## Consequences

- The engine **drives** the interaction only in mode 3 — it is otherwise a silent validator of player-ordered intents. "Engine-authoritative" does not mean "engine bosses the player."
- Both intents and prompt responses are events in the log (ADR-0005); replay and reconnection cover them for free.
- The **core rules alone require mode 3** (knockout → choose new Active), so prompts are not deferrable to a "card effects" phase — they are needed in the first vanilla-only slice.
- Per-card effects (ADR-0002) are expressed in terms of these primitives: they read/modify state, may add intent parameters, and may *raise prompts* addressed to a chosen player.
