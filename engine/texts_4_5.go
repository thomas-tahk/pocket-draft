package engine

import "errors"

// Ability effect verbs, batch 4.5 (issue #78). Inside Apply, ctx.self() is the
// Pokémon using the Ability and ctx.Target the spot it is aimed at.

// ZoneToSelf takes N Energy of a type from the Energy Zone and attaches it to
// the Pokémon using the Ability (Magneton's Volt Charge).
type ZoneToSelf struct {
	Energy Energy
	N      int
}

func (v ZoneToSelf) Apply(ctx *EffectContext) {
	ctx.self().Energy[v.Energy] += v.N
	ctx.narrate("attached %d %s Energy from the Energy Zone.", v.N, v.Energy)
}

// DamageSelf puts N damage on the Pokémon using the Ability (Hydreigon's Roar
// in Unison).
type DamageSelf struct{ N int }

func (v DamageSelf) Apply(ctx *EffectContext) {
	ctx.self().Damage += v.N
	ctx.narrate("took %d damage.", v.N)
}

// DamageOneOpp does N damage to 1 of your opponent's Pokémon (Greninja's
// Water Shuriken). It isn't an attack: no Weakness, no damage modifiers. A
// Benched Pokémon it knocks out is discarded for its points.
type DamageOneOpp struct{ N int }

func (v DamageOneOpp) Targets(ctx *EffectContext) []Spot {
	return ownSpots(ctx.g, 1-ctx.attacker, func(*InPlay) bool { return true })
}

func (v DamageOneOpp) Apply(ctx *EffectContext) {
	m := ctx.targetMon()
	m.Damage += v.N
	ctx.narrate("did %d damage to %s.", v.N, m.Card.Name)
}

// MoveTypeEnergyToActive moves all Energy of a type from 1 of your Benched
// Pokémon of that type to your Active (Lunala ex's Psychic Connect).
type MoveTypeEnergyToActive struct{ Energy Energy }

func (v MoveTypeEnergyToActive) Targets(ctx *EffectContext) []Spot {
	var out []Spot
	for _, s := range ownSpots(ctx.g, ctx.attacker, func(m *InPlay) bool {
		return m.Card.Type == v.Energy && m.Energy[v.Energy] > 0
	}) {
		if s.Pos > 0 {
			out = append(out, s)
		}
	}
	return out
}

func (v MoveTypeEnergyToActive) Apply(ctx *EffectContext) {
	from, to := ctx.targetMon(), ctx.attackerMon()
	n := from.Energy[v.Energy]
	delete(from.Energy, v.Energy)
	to.Energy[v.Energy] += n
	ctx.narrate("moved %d %s Energy from %s to %s.", n, v.Energy, from.Card.Name, to.Card.Name)
}

// DiscardOppActiveTools discards the Tool on your opponent's Active (Klefki's
// Dismantling Keys); it can't be used when there is none.
type DiscardOppActiveTools struct{}

func (DiscardOppActiveTools) CanPlay(ctx *EffectContext) error {
	if m := ctx.defenderMon(); m == nil || m.Tool == nil {
		return errors.New("your opponent's Active Pokémon has no Pokémon Tool")
	}
	return nil
}

func (DiscardOppActiveTools) Apply(ctx *EffectContext) {
	DiscardTool{}.Apply(&EffectContext{g: ctx.g, attacker: ctx.attacker, Target: Spot{1 - ctx.attacker, 0}})
}

// DiscardSelf discards the Pokémon using the Ability, with its Tool and
// attachments; it scores nothing. It must be the last op.
type DiscardSelf struct{}

func (DiscardSelf) Apply(ctx *EffectContext) {
	pl, m := ctx.Player(), ctx.self()
	ctx.narrate("was discarded.")
	discardAttached(pl, m)
	pl.Discard = append(pl.Discard, m.Card)
	if ctx.pos == 0 {
		pl.Active = nil
		return
	}
	pl.Bench[ctx.pos-1] = nil
}

// SwitchInTakeAllEnergy switches the Pokémon using the Ability, from your
// Bench, with your Active, then moves all of your Energy in play to it
// (Miraidon ex's Legendary Drive).
type SwitchInTakeAllEnergy struct{}

func (SwitchInTakeAllEnergy) Apply(ctx *EffectContext) {
	me := ctx.self()
	switchIn(ctx.g, Spot{ctx.attacker, ctx.pos})
	for _, s := range ownSpots(ctx.g, ctx.attacker, func(m *InPlay) bool { return m != me }) {
		m := targetMon(ctx.Player(), s.Pos)
		for e, n := range m.Energy {
			me.Energy[e] += n
		}
		m.Energy = map[Energy]int{}
	}
	ctx.narrate("took all of your Energy in play.")
}
