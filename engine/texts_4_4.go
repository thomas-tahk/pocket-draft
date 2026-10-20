package engine

import "errors"

// Ability texts, part 1 (issue #77). The triggers the engine runs for an
// Ability (a Checkup, being damaged by an attack, an Energy Zone attach) and
// the EffectOps the wordings are built from. server/texts_4_4.go maps each
// printed wording to them.

var (
	errNothingMoves = errors.New("that would do nothing")
	errWouldKOSelf  = errors.New("this Pokémon would be Knocked Out")
)

// healBlocked reports whether a passive says no Pokémon can be healed.
func (g *Game) healBlocked() bool { return g.hasHook(HookNoHeal, Holder{}, nil) }

// runTrigger runs ops for the Pokémon of player p at pos, which printed them.
func (g *Game) runTrigger(p, pos int, ops []EffectOp) {
	ctx := g.abilityCtx(p, pos, Spot{})
	for _, op := range ops {
		op.Apply(ctx)
	}
}

// checkupAbilities runs every AtCheckup Ability whose turn, spot and
// condition hold, after the status steps of the Checkup.
func (g *Game) checkupAbilities() {
	for p := 0; p < 2; p++ {
		for _, pos := range targets() {
			g.checkupAbility(p, pos)
		}
	}
}

func (g *Game) checkupAbility(p, pos int) {
	at := g.holder(p, pos)
	if at.InPlay == nil || at.remainingHP() <= 0 {
		return
	}
	a := g.abilityOf(at)
	if a == nil || len(a.AtCheckup) == 0 || abilitySpotOK(a, pos) != nil {
		return
	}
	own := p == g.S.Active
	if (a.CheckupOf == OwnCheckup && !own) || (a.CheckupOf == OppCheckup && own) {
		return
	}
	if a.When != nil && !a.When(AbilityCheck{Self: at, Game: g}) {
		return
	}
	g.narrate("%s's %s's %s.", side(p), at.Card.Name, a.Name)
	g.runTrigger(p, pos, a.AtCheckup)
}

// damagedTriggers runs the WhenDamaged Ability of player p's Active after an
// attack did it dmg damage. The ops see the attacker as p's opponent's Active.
func (g *Game) damagedTriggers(p, dmg int) {
	def := g.S.Players[p].Active
	if def == nil || dmg <= 0 {
		return
	}
	a := def.Card.Ability
	if a == nil || len(a.WhenDamaged) == 0 || abilitySpotOK(a, 0) != nil {
		return
	}
	g.narrate("%s's %s's %s.", side(p), def.Card.Name, a.Name)
	g.runTrigger(p, 0, a.WhenDamaged)
}

// zoneAttachTrigger runs an OnZoneAttach Ability after player p attached a t
// Energy from the Energy Zone to the Pokémon at pos.
func (g *Game) zoneAttachTrigger(p, pos int, t Energy) {
	at := g.holder(p, pos)
	a := g.abilityOf(at)
	if a == nil || a.OnZoneAttach == nil || a.OnZoneAttach.Type != t {
		return
	}
	g.narrate("%s's %s's %s.", side(p), at.Card.Name, a.Name)
	g.runTrigger(p, pos, a.OnZoneAttach.Ops)
	g.resolveAction()
}

// DamageOpp does N damage to your opponent's Active (an Ability's damage:
// no Weakness, no modifiers). Also the attacker, for a WhenDamaged Ability.
type DamageOpp struct{ N int }

func (v DamageOpp) Apply(ctx *EffectContext) {
	m := ctx.defenderMon()
	if m == nil {
		return
	}
	m.Damage += v.N
	ctx.g.narrate("%s took %d damage.", m.Card.Name, v.N)
}

// HealSelf heals N damage from the Pokémon using the Ability.
type HealSelf struct{ N int }

func (v HealSelf) Apply(ctx *EffectContext) {
	ctx.g.narrate("%s healed %d damage.", ctx.self().Card.Name, ctx.g.heal(ctx.self(), v.N))
}

// HealExDiscardEnergy heals N damage from 1 of your damaged ex that has
// Energy; if any was healed, a random Energy is discarded from it (Arboliva).
type HealExDiscardEnergy struct{ N int }

func (v HealExDiscardEnergy) Targets(ctx *EffectContext) []Spot {
	return ctx.g.ownSpots(ctx.attacker, func(_ Spot, m *InPlay) bool {
		return m.Card.IsEX && m.Damage > 0 && m.totalEnergy() > 0
	})
}

func (v HealExDiscardEnergy) Apply(ctx *EffectContext) {
	mon := ctx.targetMon()
	if ctx.g.heal(mon, v.N) == 0 {
		return
	}
	pl := &ctx.g.S.Players[ctx.Target.Player]
	var held []Energy
	for _, e := range energyOrder {
		for i := 0; i < mon.Energy[e]; i++ {
			held = append(held, e)
		}
	}
	e := held[ctx.g.rng.intn(len(held))]
	mon.Energy[e]--
	pl.DiscardedEnergy = append(pl.DiscardedEnergy, e)
	ctx.g.narrate("%s healed and discarded a %s Energy.", mon.Card.Name, e)
}

// MoveDamageHere moves up to N damage from your Active to the Benched Pokémon
// using the Ability (Brambleghast).
type MoveDamageHere struct{ N int }

func (v MoveDamageHere) CanPlay(ctx *EffectContext) error {
	from, to := ctx.attackerMon(), ctx.self()
	switch {
	case from == nil || from == to || from.Damage == 0:
		return errNothingMoves
	case to.remainingHP() <= min(v.N, from.Damage):
		return errWouldKOSelf
	}
	return nil
}

func (v MoveDamageHere) Apply(ctx *EffectContext) {
	from, to := ctx.attackerMon(), ctx.self()
	n := min(v.N, from.Damage)
	from.Damage -= n
	to.Damage += n
	ctx.g.narrate("%d damage moved from %s to %s.", n, from.Card.Name, to.Card.Name)
}

// MoveAllDamageHere moves all of the damage from 1 of your other Pokémon to
// the one using the Ability (Dusknoir). A Benched user can't take enough to
// be Knocked Out: only the Active Spot is checked for Knock Outs.
type MoveAllDamageHere struct{}

func (MoveAllDamageHere) Targets(ctx *EffectContext) []Spot {
	self := ctx.self()
	return ctx.g.ownSpots(ctx.attacker, func(s Spot, m *InPlay) bool {
		return s.Pos != ctx.pos && m.Damage > 0 && (ctx.pos == 0 || self.remainingHP() > m.Damage)
	})
}

func (MoveAllDamageHere) Apply(ctx *EffectContext) {
	from, to := ctx.targetMon(), ctx.self()
	to.Damage += from.Damage
	ctx.g.narrate("%d damage moved from %s to %s.", from.Damage, from.Card.Name, to.Card.Name)
	from.Damage = 0
}

// Combust attaches a Type Energy from your discard pile to this Pokémon, then
// it takes Damage (Flareon ex).
type Combust struct {
	Type   Energy
	Damage int
}

func (v Combust) CanPlay(ctx *EffectContext) error {
	if indexOfEnergy(ctx.Player().DiscardedEnergy, v.Type) < 0 {
		return errNothingMoves
	}
	if ctx.pos != 0 && ctx.self().remainingHP() <= v.Damage {
		return errWouldKOSelf
	}
	return nil
}

func (v Combust) Apply(ctx *EffectContext) {
	pl, mon := ctx.Player(), ctx.self()
	i := indexOfEnergy(pl.DiscardedEnergy, v.Type)
	pl.DiscardedEnergy = append(pl.DiscardedEnergy[:i:i], pl.DiscardedEnergy[i+1:]...)
	mon.Energy[v.Type]++
	mon.Damage += v.Damage
	ctx.g.narrate("%s attached a %s Energy and took %d damage.", mon.Card.Name, v.Type, v.Damage)
}

// DiscardTopOppDeck discards the top card of your opponent's deck.
type DiscardTopOppDeck struct{}

func (DiscardTopOppDeck) CanPlay(ctx *EffectContext) error {
	if len(ctx.opp().Deck) == 0 {
		return errNothingMoves
	}
	return nil
}

func (DiscardTopOppDeck) Apply(ctx *EffectContext) {
	opp := ctx.opp()
	c := opp.Deck[0]
	opp.Deck = opp.Deck[1:]
	opp.Discard = append(opp.Discard, c)
	ctx.g.narrate("%s's %s was discarded from the top of their deck.", side(1-ctx.attacker), c.Name)
}

// RandomOppCondition puts one of Burned, Confused and Poisoned, chosen at
// random from those it doesn't already have, on your opponent's Active (Dustox).
type RandomOppCondition struct{}

func (RandomOppCondition) options(ctx *EffectContext) []Status {
	var out []Status
	for _, s := range []Status{Burn, Confused, Poison} {
		if !ctx.defenderMon().Has(s) {
			out = append(out, s)
		}
	}
	return out
}

func (v RandomOppCondition) CanPlay(ctx *EffectContext) error {
	if len(v.options(ctx)) == 0 {
		return errNothingMoves
	}
	return nil
}

func (v RandomOppCondition) Apply(ctx *EffectContext) {
	opts := v.options(ctx)
	if len(opts) == 0 || !ctx.g.canBeAffected(1-ctx.attacker, 0) {
		return
	}
	s := opts[ctx.g.rng.intn(len(opts))]
	ctx.defenderMon().Inflict(s)
	ctx.g.narrate("%s is now %s.", ctx.defenderMon().Card.Name, s)
}

// AttachZoneToActive takes a Type Energy from the Energy Zone and attaches it
// to your Active, if it is a Type Pokémon (Baxcalibur, Charmeleon).
type AttachZoneToActive struct{ Type Energy }

func (v AttachZoneToActive) CanPlay(ctx *EffectContext) error {
	if ctx.attackerMon().Card.Type != v.Type {
		return errNothingMoves
	}
	return nil
}

func (v AttachZoneToActive) Apply(ctx *EffectContext) {
	mon := ctx.attackerMon()
	mon.Energy[v.Type]++
	ctx.g.narrate("%s attached a %s Energy to %s.", side(ctx.attacker), v.Type, mon.Card.Name)
}

// AttachDiscardToActive attaches a random Energy from your discard pile to
// your Active, if it is a Type Pokémon (Dragonair).
type AttachDiscardToActive struct{ Type Energy }

func (v AttachDiscardToActive) CanPlay(ctx *EffectContext) error {
	if ctx.attackerMon().Card.Type != v.Type || len(ctx.Player().DiscardedEnergy) == 0 {
		return errNothingMoves
	}
	return nil
}

func (v AttachDiscardToActive) Apply(ctx *EffectContext) {
	pl, mon := ctx.Player(), ctx.attackerMon()
	i := ctx.g.rng.intn(len(pl.DiscardedEnergy))
	e := pl.DiscardedEnergy[i]
	pl.DiscardedEnergy = append(pl.DiscardedEnergy[:i:i], pl.DiscardedEnergy[i+1:]...)
	mon.Energy[e]++
	ctx.g.narrate("%s attached %s Energy from the discard pile to %s.", side(ctx.attacker), e, mon.Card.Name)
}

// SwitchUltraBeast switches your Active Ultra Beast with 1 of your Benched
// Ultra Beasts (Celesteela).
type SwitchUltraBeast struct{}

func (SwitchUltraBeast) Targets(ctx *EffectContext) []Spot {
	if a := ctx.Player().Active; a == nil || !a.Card.IsUltraBeast() {
		return nil
	}
	return spotsOf(ctx.g, ctx.attacker, benchPositions(), func(m *InPlay) bool { return m.Card.IsUltraBeast() })
}

func (SwitchUltraBeast) Apply(ctx *EffectContext) { switchIn(ctx.g, ctx.Target) }

// EvolveRandomFromDeck evolves the Pokémon using the Ability with a random
// card from your deck that evolves from it (Caterpie).
type EvolveRandomFromDeck struct{}

func (EvolveRandomFromDeck) Apply(ctx *EffectContext) {
	mon, pl := ctx.self(), ctx.Player()
	var idx []int
	for i, c := range pl.Deck {
		if c.Kind == KindPokemon && c.EvolvesFrom == mon.Card.Name {
			idx = append(idx, i)
		}
	}
	if len(idx) == 0 {
		return
	}
	i := idx[ctx.g.rng.intn(len(idx))]
	ev := pl.Deck[i]
	pl.Deck = removeAt(pl.Deck, i)
	ctx.g.evolveMon(ctx.attacker, mon, ev)
}

// RandomsFromDiscard puts up to N random cards that Match accepts from your
// discard pile into your hand (Perrserker).
type RandomsFromDiscard struct {
	N     int
	Match func(Card) bool
}

func (v RandomsFromDiscard) matches(ctx *EffectContext) []int {
	var out []int
	for i, c := range ctx.Player().Discard {
		if v.Match(c) {
			out = append(out, i)
		}
	}
	return out
}

func (v RandomsFromDiscard) CanPlay(ctx *EffectContext) error {
	if err := handFull(ctx); err != nil {
		return err
	}
	if len(v.matches(ctx)) == 0 {
		return errNoMatch
	}
	return nil
}

func (v RandomsFromDiscard) Apply(ctx *EffectContext) {
	pl := ctx.Player()
	for n := 0; n < v.N && len(pl.Hand) < handMax; n++ {
		m := v.matches(ctx)
		if len(m) == 0 {
			return
		}
		i := m[ctx.g.rng.intn(len(m))]
		c := pl.Discard[i]
		pl.Discard = removeAt(pl.Discard, i)
		pl.Hand = append(pl.Hand, c)
		ctx.g.narrate("%s put %s into %s's hand.", ctx.self().Card.Name, c.Name, side(ctx.attacker))
	}
}

// CureOwnRandomCondition removes a random Special Condition from your Active,
// and can't be used when it has none (Blissey).
type CureOwnRandomCondition struct{ CureRandomCondition }

func (CureOwnRandomCondition) CanPlay(ctx *EffectContext) error {
	if len(ctx.attackerMon().Conditions()) == 0 {
		return errNothingMoves
	}
	return nil
}
