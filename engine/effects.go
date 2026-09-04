package engine

// EffectOp is one primitive verb an attack's text is built from — data, not
// bespoke per-card code (Approach A, docs/superpowers/specs/2026-07-13-effect-
// engine-slice-1-design.md). Apply mutates the resolution context for the
// attack currently resolving.
type EffectOp interface {
	Apply(ctx *EffectContext)
}

// EffectContext is the only seam verbs touch: the running damage for the
// attack currently resolving, plus access to the attacker/defender and a
// seeded coin flip so replay stays deterministic (ADR-0005).
type EffectContext struct {
	g        *Game
	attacker int
	Damage   int
}

// Flip draws one coin from the engine's seeded RNG — never a second RNG
// source, so folding the same log always reproduces the same flips.
func (ctx *EffectContext) Flip() bool { return ctx.g.rng.intn(2) == 0 }

func (ctx *EffectContext) attackerMon() *InPlay    { return ctx.g.S.Players[ctx.attacker].Active }
func (ctx *EffectContext) defenderMon() *InPlay    { return ctx.g.S.Players[1-ctx.attacker].Active }
func (ctx *EffectContext) attackerPlayer() *Player { return &ctx.g.S.Players[ctx.attacker] }

// narrate prefixes the attacker's side and Pokémon name, matching the
// existing narration style in applyAttack.
func (ctx *EffectContext) narrate(format string, args ...any) {
	prefix := []any{side(ctx.attacker), ctx.attackerMon().Card.Name}
	ctx.g.narrate("%s's %s "+format, append(prefix, args...)...)
}

// FlipForBonus: flip a coin; heads adds Bonus damage. E.g. Sneasel's Quick
// Attack (10+): flip a coin, if heads +20.
type FlipForBonus struct{ Bonus int }

func (v FlipForBonus) Apply(ctx *EffectContext) {
	if ctx.Flip() {
		ctx.Damage += v.Bonus
		ctx.narrate("flipped heads: +%d damage.", v.Bonus)
	} else {
		ctx.narrate("flipped tails: no bonus.")
	}
}

// DamagePerHeads: flip Coins coins, add Per damage for each heads. E.g.
// Petilil's Double Spin (20x): flip 2, 20 per heads.
type DamagePerHeads struct {
	Coins int
	Per   int
}

func (v DamagePerHeads) Apply(ctx *EffectContext) {
	heads := 0
	for i := 0; i < v.Coins; i++ {
		if ctx.Flip() {
			heads++
		}
	}
	bonus := heads * v.Per
	ctx.Damage += bonus
	ctx.narrate("flipped %d of %d heads: +%d damage.", heads, v.Coins, bonus)
}

// Status is a status condition a verb can inflict.
type Status uint8

const (
	Burn Status = iota
	Poison
)

// ApplyStatus inflicts a status condition on the defending Active. E.g.
// Ponyta's Singe: opponent's Active is now Burned.
type ApplyStatus struct{ Status Status }

func (v ApplyStatus) Apply(ctx *EffectContext) {
	switch v.Status {
	case Burn:
		ctx.defenderMon().Burned = true
	case Poison:
		ctx.defenderMon().Poisoned = true
	}
}

// EffectTarget names whose energy DamagePerEnergy reads.
type EffectTarget uint8

const (
	OppActive EffectTarget = iota
)

// DamagePerEnergy adds Per damage for each Energy attached to Where. E.g.
// Indeedee ex's Psychic (30+): +30 per Energy on the opponent's Active.
type DamagePerEnergy struct {
	Per   int
	Where EffectTarget
}

func (v DamagePerEnergy) Apply(ctx *EffectContext) {
	// Where is always OppActive in this slice (the only target this verb set
	// wires up); the field exists so a later slice can add Self without
	// redesigning the verb.
	ctx.Damage += v.Per * ctx.defenderMon().totalEnergy()
}

// DrawCards draws N cards from the attacker's deck into their hand. E.g.
// Munchlax's Hungrily Draw: draw a card.
type DrawCards struct{ N int }

func (v DrawCards) Apply(ctx *EffectContext) {
	pl := ctx.attackerPlayer()
	for i := 0; i < v.N && len(pl.Deck) > 0; i++ {
		pl.Hand = append(pl.Hand, pl.Deck[0])
		pl.Deck = pl.Deck[1:]
	}
}
