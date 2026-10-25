package engine

// Pokémon Tools (spec A3). A Tool is played from hand onto one of your
// Pokémon that has no Tool and stays there: its Card.Passive applies while it
// is attached (passive.go). It leaves play with its Pokémon, or when a card
// removes it; then that Pokémon can take a Tool again.

// toolTargets is each of p's Pokémon in play without a Tool.
func toolTargets(g *Game, p int) []Spot {
	var out []Spot
	for _, pos := range targets() {
		if mon := targetMon(&g.S.Players[p], pos); mon != nil && mon.Tool == nil {
			out = append(out, Spot{p, pos})
		}
	}
	return out
}

func (g *Game) attachTool(p int, tool Card, at Spot) {
	mon := targetMon(&g.S.Players[at.Player], at.Pos)
	mon.Tool = &tool
	g.narrate("%s attached %s to %s.", side(p), tool.Name, mon.Card.Name)
	g.fireTool(OnAttached, at.Player, at.Pos)
}

// DiscardTool discards the Tool from the targeted Pokémon, yours or your
// opponent's (Field Blower).
type DiscardTool struct{}

// Targets is every Pokémon in play with a Tool.
func (DiscardTool) Targets(ctx *EffectContext) []Spot {
	var out []Spot
	for p := 0; p < 2; p++ {
		for _, pos := range targets() {
			if mon := targetMon(&ctx.g.S.Players[p], pos); mon != nil && mon.Tool != nil {
				out = append(out, Spot{p, pos})
			}
		}
	}
	return out
}

func (DiscardTool) Apply(ctx *EffectContext) {
	owner := &ctx.g.S.Players[ctx.Target.Player]
	mon := targetMon(owner, ctx.Target.Pos)
	owner.Discard = append(owner.Discard, *mon.Tool)
	ctx.g.narrate("%s was discarded from %s.", mon.Tool.Name, mon.Card.Name)
	mon.Tool = nil
}
