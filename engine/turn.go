package engine

// Will's coin effect, which lasts for the rest of the turn it was played.

// NextFlipHeads makes the acting player's next effect coin flip this turn
// land heads (Will). Confusion and Checkup flips aren't an effect's.
type NextFlipHeads struct{}

func (NextFlipHeads) Apply(ctx *EffectContext) { ctx.Player().nextHeads = true }

func (p *Player) endTurnEffects() {
	p.nextHeads = false
}
