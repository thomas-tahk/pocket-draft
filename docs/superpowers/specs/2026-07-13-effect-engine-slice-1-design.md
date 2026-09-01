# Effect Engine — Slice 1 (Design)

_2026-07-13. First vertical slice of Phase 2 (effect engine) from the draft-mode
milestone (`docs/superpowers/specs/2026-07-12-draft-mode-full-flow-design.md`).
Approach A (shared primitive toolbox) is locked by that spec; this slice proves the
architecture end-to-end on a handful of real cards before scaling._

## Goal

Make attacks **do what their text says** — coin flips, scaled damage, status,
draw — instead of only adding a printed number. Prove the toolbox architecture on
**6 verbs across 5 real draftable cards**, wired through the engine and visible on
the board in a real game. Then later slices widen the verb set and map more of the
pool.

**Success = one demo game where a coin-flip attack, a burn, and a draw all fire and
show on the board, plus deterministic per-verb unit tests are green.**

## Locked by the parent spec (not re-litigated here)

- **Approach A:** an attack's effect is an ordered list of primitive verb
  invocations — data, not bespoke per-card code — with an imperative escape hatch.
- Engine stays event-sourced / deterministic / seeded (ADR-0005); card text
  overrides base rules (ADR-0007 enforcement never removed).

## Representation — `EffectOp` interface (the foundational decision)

A verb is a small Go struct implementing one method:

```go
// An EffectOp is one primitive verb. Apply mutates the resolution context
// (running damage, statuses, draws) for the attack currently resolving.
type EffectOp interface {
    Apply(ctx *EffectContext)
}
```

An attack's effect is **data** — a slice of these structs:

```go
// Sneasel — "Quick Attack (10+): flip a coin, if heads +20"
Attacks: []Attack{{Name: "Quick Attack", Damage: 10, Effect: []EffectOp{
    FlipForBonus{Bonus: 20},
}}}
```

**Why this over alternatives:** it *is* "data not code" (structs are values), is
type-safe, lets each verb be unit-tested in isolation, and makes the escape hatch
just another struct implementing `Apply`. It scales to the full ~655 effect strings
without a hand-written parser or a giant `switch` interpreter (those pay off only if
effects were serialized to JSON/DB — they are authored in Go). Closures were
rejected: opaque, untestable, no escape-hatch story.

## The 6 verbs / 5 cards in this slice

All **Basics**, so no evolution is needed to demo them. Each proves one category:

| Card | Attack (text) | Verb |
|---|---|---|
| Sneasel | Quick Attack (10+): flip, heads → +20 | `FlipForBonus{Bonus:20}` |
| Petilil | Double Spin (20×): flip 2, 20 per heads | `DamagePerHeads{Coins:2, Per:20}` |
| Ponyta | Singe: opponent's Active is Burned | `ApplyStatus{Status:Burn}` |
| Indeedee ex | Psychic (30+): +30 per Energy on opp. Active | `DamagePerEnergy{Per:30, Where:OppActive}` |
| Munchlax | Hungrily Draw (10): draw a card | `DrawCards{N:1}` |

**Poison rides along for free:** `ApplyStatus{Status:Poison}` is the same verb with
a different param, and the model/board already handle poison — so the slice ships
6 verbs at the cost of 5.

## Resolution mechanics

One field is added to `Attack`: `Effect []EffectOp`. **`nil` means today's vanilla
base-damage path, unchanged** — every existing card and test keeps working.

`applyAttack` (`engine/engine.go:285`) changes from "add the number" to run the
effect list against a small context:

```go
ctx := &EffectContext{g: g, attacker: e.Player, Damage: atk.Damage}
for _, op := range atk.Effect {
    op.Apply(ctx)
}
// existing weakness handling, then apply ctx.Damage to the defender,
// then the existing KO / conclude-turn logic — unchanged.
```

`EffectContext` is the only seam verbs touch:

- holds the running `Damage`,
- exposes `Flip() bool`, pulling from the engine's **existing seeded RNG** so replay
  stays deterministic (coin flips are re-derived on fold, never stored — consistent
  with how draws already work),
- gives verbs access to attacker/defender `InPlay` + player state for status and
  draw.

**No new event types and no new prompt kinds** — every verb in this slice is
automatic (targets self or the opponent's Active). Player-choice verbs (pick a
bench target, discard from hand) reuse the existing prompt system (ADR-0007) in a
later slice.

## Board (minimal)

The board already renders what this slice produces: the model carries
`Poisoned`/`Burned`, `server/view.go` emits them, and `src/game/PlayerSide.tsx`
shows ☠/🔥 badges plus a narration log; hand counts render already.

Only change: **narration** names coin-flip outcomes (e.g. "Sneasel flipped heads:
+20 damage") so the player can see why damage varied. No board redesign in this
slice — that redesign is the cross-cutting Phase 1–3 work, out of scope here.

## Testing & verification

- **Per-verb seeded unit tests** in the engine — the real verification. Fixed seed →
  deterministic outcome (Sneasel heads = 30 / tails = 10; Petilil 2 heads = 40, 1 =
  20; Indeedee vs an N-energy Active = 30+30N; Ponyta leaves the defender Burned;
  Munchlax hand grows by 1). Set state directly, assert the number/flag.
- **Vanilla regression:** an attack with `Effect: nil` behaves exactly as today.
- **End-to-end:** a demo game using these cards where a coin-flip attack, a burn,
  and a draw all fire and are visible on the board. The demo deck **sets its Energy
  Zone types explicitly** to the types those cards need (see below), so attacks are
  never energy-starved during the demo.
- **UI:** manual browser check (project convention; no JS test runner).

## Out of scope (this slice)

Evolution-triggered effects, abilities (the 143 ability strings), Trainer/Supporter/
Item cards, and any player-choice (targeting/discard) verbs. All reuse the same
`EffectOp` interface, so nothing here blocks them.

## Backlog captured (not built here)

- **Energy-type set should be an explicit, player-chosen deck attribute.** Today the
  engine auto-derives Energy Zone types from the deck's Pokémon types
  (`energyTypesOf` in `engine/types.go`). The real rule: the player freely chooses
  the zone's energy type set at deck construction, decoupled from Pokémon types.
  Auto-derivation is not just imprecise but **strategically wrong** — a player
  deliberately excludes a type whose Pokémon is run only for its ability/effect
  (Greninja's Water Shuriken, Indeedee's heal), never to attack, to avoid diluting
  the zone. This is a **deck-bridge / construction** concern (Phase 1 territory),
  and becomes pressing once the **ability-effect slice** lands. Fix is small: the
  deck carries its chosen energy-type set as input. Not expanded into this slice.

- **Rule-overriding effects need a second mechanism, and it is undesigned.** Both
  this spec and the parent one assert that "card text overrides base rules"
  (ADR-0007), but nothing states how an override is *expressed*. `EffectOp` as
  built here cannot express one: `applyAttack` runs the verb list against a
  running `Damage`, then executes weakness, application, and KO **unchanged**.
  That is an *additive* model — it covers "+20 damage", "opponent is Burned",
  "draw a card", which is most effect text. It does not cover:
  - "this attack ignores Weakness" — weakness runs after the verbs, unconditionally
  - "prevent all damage to this Pokémon during your opponent's next turn" — no
    modifier survives past the attack that created it
  - "your opponent can't retreat next turn" / "this Pokémon has no Retreat Cost" —
    the decision point is not in `applyAttack` at all

  Two things are missing: **hooks at the other rule checkpoints** (retreat cost,
  weakness, damage application, status ticks) and **persistent modifiers** with an
  owner and an expiry, since an override usually outlives the attack. ADR-0007
  already names the right plug points — effects "read/modify state, may add intent
  parameters, and may raise prompts" — so this extends that model rather than
  replacing it.

  **Do not defer this past the verb set growing.** Retrofitting checkpoint hooks
  into 6 verbs is cheap; into 40 it is a rewrite. The trigger to design it is the
  first card whose text removes or replaces a rule rather than adding to one.
