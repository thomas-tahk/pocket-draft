package engine

import "errors"

// Ability texts, part 3 (issue #79, plan 4.6). The EffectOps the wordings
// server/texts_4_6.go maps to that the earlier parts did not have.

// SwitchTypedActive switches your Active of Type with 1 of your Benched
// Pokémon (Metal Transport).
type SwitchTypedActive struct{ Type Energy }

func (v SwitchTypedActive) Targets(ctx *EffectContext) []Spot {
	if a := ctx.Player().Active; a == nil || a.Card.Type != v.Type {
		return nil
	}
	return spotsOf(ctx.g, ctx.attacker, benchPositions(), nil)
}

func (SwitchTypedActive) Apply(ctx *EffectContext) { switchIn(ctx.g, ctx.Target) }

// FlipGust flips a coin; heads switches in 1 of your opponent's Benched
// Pokémon (Captivating Rhythm).
type FlipGust struct{}

func (FlipGust) Targets(ctx *EffectContext) []Spot {
	return spotsOf(ctx.g, 1-ctx.attacker, benchPositions(), nil)
}

func (FlipGust) Apply(ctx *EffectContext) {
	if ctx.Flip() {
		switchIn(ctx.g, ctx.Target)
	}
}

// SwitchInOppBasic switches in 1 of your opponent's Benched Basic Pokémon
// (Fragrance Trap).
type SwitchInOppBasic struct{}

func (SwitchInOppBasic) Targets(ctx *EffectContext) []Spot {
	return spotsOf(ctx.g, 1-ctx.attacker, benchPositions(), func(m *InPlay) bool { return m.Card.Stage == Basic })
}

func (SwitchInOppBasic) Apply(ctx *EffectContext) { switchIn(ctx.g, ctx.Target) }

// SwitchSelfActive switches this Benched Pokémon with your Active (Rising Road).
type SwitchSelfActive struct{}

func (SwitchSelfActive) Apply(ctx *EffectContext) {
	switchIn(ctx.g, Spot{Player: ctx.attacker, Pos: ctx.pos})
}

// AttachZoneToSelf takes Energy of Type from your Energy Zone and attaches it
// to this Pokémon (Thunderclap Flash).
type AttachZoneToSelf struct{ Type Energy }

func (v AttachZoneToSelf) Apply(ctx *EffectContext) {
	pl := ctx.Player()
	if pl.EnergyZone != v.Type {
		return
	}
	ctx.self().Energy[v.Type]++
	pl.EnergyZone = NoEnergy
	ctx.g.narrate("%s attached a %s Energy to %s.", side(ctx.attacker), v.Type, ctx.self().Card.Name)
}

// DiscardOwnEnergy discards 1 Energy of Type from this Pokémon. It is a cost:
// the Ability can't be used without one (Passionate Voice).
type DiscardOwnEnergy struct{ Type Energy }

func (v DiscardOwnEnergy) CanPlay(ctx *EffectContext) error {
	if ctx.self().Energy[v.Type] == 0 {
		return errNothingMoves
	}
	return nil
}

func (v DiscardOwnEnergy) Apply(ctx *EffectContext) { ctx.self().Energy[v.Type]-- }

// ShieldSelfNextTurn: this Pokémon takes no damage and no effects of attacks
// until the end of your opponent's next turn (Stance).
type ShieldSelfNextTurn struct{}

func (ShieldSelfNextTurn) Apply(ctx *EffectContext) {
	mon := ctx.self()
	onMon := func(c PassiveCheck) bool { return c.Target.InPlay == mon }
	ForTurn{Next: true, Passive: Passive{Hook: HookDamageTaken, Value: -noDamageLimit, When: onMon}}.Apply(ctx)
	ForTurn{Next: true, Passive: Passive{Hook: HookEffectShield, When: onMon}}.Apply(ctx)
}

// noDamageLimit is more than any attack can do, for "takes no damage".
const noDamageLimit = 10000

// MoveOneTypeEnergyToActive moves 1 Energy of Type from a Benched Pokémon of
// yours to your Active, which must be of Type (Wash Out).
type MoveOneTypeEnergyToActive struct{ Type Energy }

func (v MoveOneTypeEnergyToActive) Targets(ctx *EffectContext) []Spot {
	if a := ctx.Player().Active; a == nil || a.Card.Type != v.Type {
		return nil
	}
	return spotsOf(ctx.g, ctx.attacker, benchPositions(), func(m *InPlay) bool {
		return m.Card.Type == v.Type && m.Energy[v.Type] > 0
	})
}

func (v MoveOneTypeEnergyToActive) Apply(ctx *EffectContext) {
	from, to := ctx.targetMon(), ctx.Player().Active
	from.Energy[v.Type]--
	to.Energy[v.Type]++
}

// MoveTypeEnergyHere moves all Energy of Type from each of your Pokémon to
// this one (Energy Plunder).
type MoveTypeEnergyHere struct{ Type Energy }

func (v MoveTypeEnergyHere) Apply(ctx *EffectContext) {
	to := ctx.self()
	for _, s := range spotsOf(ctx.g, ctx.attacker, targets(), nil) {
		from := targetMon(ctx.Player(), s.Pos)
		if from == to {
			continue
		}
		to.Energy[v.Type] += from.Energy[v.Type]
		from.Energy[v.Type] = 0
	}
}

// MoveFoeEnergyHere moves a random Energy from your opponent's Active to this
// Pokémon (Thieving Incisors).
type MoveFoeEnergyHere struct{}

func (MoveFoeEnergyHere) CanPlay(ctx *EffectContext) error {
	if foe := ctx.defenderMon(); foe == nil || foe.totalEnergy() == 0 {
		return errors.New("your opponent's Active has no Energy")
	}
	return nil
}

func (MoveFoeEnergyHere) Apply(ctx *EffectContext) {
	foe := ctx.defenderMon()
	var have []Energy
	for _, e := range energyOrder {
		if foe.Energy[e] > 0 {
			have = append(have, e)
		}
	}
	e := have[ctx.g.rng.intn(len(have))]
	foe.Energy[e]--
	ctx.self().Energy[e]++
}
