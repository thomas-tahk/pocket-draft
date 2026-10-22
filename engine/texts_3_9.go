package engine

import (
	"errors"
	"slices"
)

// Effect ops for Supporter texts, batch 3.9. Each is one wording's verb; the
// wordings are mapped to them in server/texts_3_9.go.

// errNoEffect is why a card whose text would do nothing can't be played.
var errNoEffect = errors.New("it would have no effect")

// spotsWhere is each of player p's Pokémon in play that ok accepts.
func spotsWhere(g *Game, p int, ok func(*InPlay) bool) []Spot {
	var out []Spot
	for _, pos := range targets() {
		if mon := targetMon(&g.S.Players[p], pos); mon != nil && ok(mon) {
			out = append(out, Spot{p, pos})
		}
	}
	return out
}

func (ctx *EffectContext) targetMon() *InPlay {
	return targetMon(&ctx.g.S.Players[ctx.Target.Player], ctx.Target.Pos)
}

func (ctx *EffectContext) opponent() *Player { return &ctx.g.S.Players[1-ctx.attacker] }

func named(names []string) func(*InPlay) bool {
	return func(m *InPlay) bool { return slices.Contains(names, m.Card.Name) }
}

// heal removes up to n damage from mon.
func (ctx *EffectContext) heal(mon *InPlay, n int) {
	if ctx.g.healBlocked() {
		n = 0
	}
	n = min(n, mon.Damage)
	mon.Damage -= n
	ctx.g.narrate("%s healed %d damage.", mon.Card.Name, n)
}

// HealType heals Amount from 1 of your damaged Pokémon of Type (Erika).
type HealType struct {
	Amount int
	Type   Energy
}

func (v HealType) Targets(ctx *EffectContext) []Spot {
	return spotsWhere(ctx.g, ctx.attacker, func(m *InPlay) bool { return m.Card.Type == v.Type && m.Damage > 0 })
}

func (v HealType) Apply(ctx *EffectContext) { ctx.heal(ctx.targetMon(), v.Amount) }

// HealThenDiscard heals Amount from 1 of your Pokémon of Type with at least N
// Energy of that type; if it healed any, N of that Energy are discarded
// (Diantha).
type HealThenDiscard struct {
	Amount int
	Type   Energy
	N      int
}

func (v HealThenDiscard) Targets(ctx *EffectContext) []Spot {
	return spotsWhere(ctx.g, ctx.attacker, func(m *InPlay) bool { return m.Card.Type == v.Type && m.Energy[v.Type] >= v.N })
}

func (v HealThenDiscard) Apply(ctx *EffectContext) {
	mon := ctx.targetMon()
	if mon.Damage == 0 {
		return
	}
	ctx.heal(mon, v.Amount)
	mon.Energy[v.Type] -= v.N
	for i := 0; i < v.N; i++ {
		ctx.Player().DiscardedEnergy = append(ctx.Player().DiscardedEnergy, v.Type)
	}
}

// AttachZoneEnergy takes Energy from the Energy Zone and attaches it to 1 of
// your Pokémon named in Names (Brock).
type AttachZoneEnergy struct {
	Energy Energy
	Names  []string
}

func (v AttachZoneEnergy) Targets(ctx *EffectContext) []Spot {
	return spotsWhere(ctx.g, ctx.attacker, named(v.Names))
}

func (v AttachZoneEnergy) Apply(ctx *EffectContext) {
	mon := ctx.targetMon()
	mon.Energy[v.Energy]++
	ctx.g.narrate("%s energy attached to %s.", v.Energy, mon.Card.Name)
}

// AttachZoneEnergyEach attaches Energy from the Energy Zone to each of your
// Pokémon named in Names (Fantina).
type AttachZoneEnergyEach struct {
	Energy Energy
	Names  []string
}

func (v AttachZoneEnergyEach) CanPlay(ctx *EffectContext) error {
	if len(spotsWhere(ctx.g, ctx.attacker, named(v.Names))) == 0 {
		return errNoEffect
	}
	return nil
}

func (v AttachZoneEnergyEach) Apply(ctx *EffectContext) {
	for _, s := range spotsWhere(ctx.g, ctx.attacker, named(v.Names)) {
		AttachZoneEnergy{Energy: v.Energy}.Apply(&EffectContext{g: ctx.g, attacker: ctx.attacker, Target: s})
	}
}

// ActiveToHand puts your Active into your hand if it is named Name (Budding
// Expeditioner); its Energy and Tool are discarded and you choose a new
// Active from your Bench.
type ActiveToHand struct{ Name string }

func (v ActiveToHand) CanPlay(ctx *EffectContext) error {
	pl := ctx.Player()
	switch {
	case pl.Active == nil || pl.Active.Card.Name != v.Name:
		return errors.New("your Active isn't " + v.Name)
	case pl.BenchCount() == 0:
		return errors.New("you have no Benched Pokémon to take its place")
	}
	return handFull(ctx)
}

func (v ActiveToHand) Apply(ctx *EffectContext) {
	pl := ctx.Player()
	mon := pl.Active
	pl.Active = nil
	discardAttached(pl, mon)
	pl.Hand = append(pl.Hand, mon.Card)
	ctx.g.narrate("%s returned to %s's hand.", mon.Card.Name, side(ctx.attacker))
	ctx.g.promptPromotion()
}

// discardAttached discards mon's Energy and Tool as it leaves play.
func discardAttached(pl *Player, mon *InPlay) {
	for _, e := range energyOrder {
		for i := 0; i < mon.Energy[e]; i++ {
			pl.DiscardedEnergy = append(pl.DiscardedEnergy, e)
		}
	}
	if mon.Tool != nil {
		pl.Discard = append(pl.Discard, *mon.Tool)
	}
}

// DrawPerMegaInPlay draws a card for each Mega Evolution Pokémon ex in play,
// both players' (Calem).
type DrawPerMegaInPlay struct{}

func (DrawPerMegaInPlay) count(ctx *EffectContext) int {
	n := 0
	for p := 0; p < 2; p++ {
		n += len(spotsWhere(ctx.g, p, func(m *InPlay) bool { return m.Card.IsMegaEX }))
	}
	return n
}

func (v DrawPerMegaInPlay) CanPlay(ctx *EffectContext) error {
	if v.count(ctx) == 0 {
		return errNoEffect
	}
	return handFull(ctx)
}

func (v DrawPerMegaInPlay) Apply(ctx *EffectContext) { DrawCards{N: v.count(ctx)}.Apply(ctx) }

// CardMatch says which cards a search or pick accepts.
type CardMatch func(Card) bool

// FromDiscard puts random cards that Match from your discard pile into your
// hand: 1, or with Coins one per heads of that many flips (Celestic Town
// Elder, Fisher).
type FromDiscard struct {
	Match CardMatch
	Coins int
}

func (v FromDiscard) CanPlay(ctx *EffectContext) error {
	if !slices.ContainsFunc(ctx.Player().Discard, v.Match) {
		return errNoEffect
	}
	return handFull(ctx)
}

func (v FromDiscard) Apply(ctx *EffectContext) {
	n := 1
	if v.Coins > 0 {
		n = 0
		for i := 0; i < v.Coins; i++ {
			if ctx.Flip() {
				n++
			}
		}
	}
	for i := 0; i < n; i++ {
		ctx.discardToHand(v.Match)
	}
}

// discardToHand moves one random card that match accepts from the discard
// pile to the hand; nothing at a full hand or with none left.
func (ctx *EffectContext) discardToHand(match CardMatch) {
	pl := ctx.Player()
	var idx []int
	for i, c := range pl.Discard {
		if match(c) {
			idx = append(idx, i)
		}
	}
	if len(idx) == 0 || len(pl.Hand) >= handMax {
		return
	}
	i := idx[ctx.g.rng.intn(len(idx))]
	c := pl.Discard[i]
	pl.Discard = removeAt(pl.Discard, i)
	pl.Hand = append(pl.Hand, c)
	ctx.g.narrate("%s put %s from the discard pile into their hand.", side(ctx.attacker), c.Name)
}

// ShuffleHandDrawPerOpp shuffles your hand into your deck, then draws a card
// for each card in your opponent's hand (Copycat).
type ShuffleHandDrawPerOpp struct{}

func (ShuffleHandDrawPerOpp) Apply(ctx *EffectContext) {
	pl := ctx.Player()
	pl.Deck = append(pl.Deck, pl.Hand...)
	pl.Hand = nil
	ctx.g.rng.shuffle(len(pl.Deck), func(a, b int) { pl.Deck[a], pl.Deck[b] = pl.Deck[b], pl.Deck[a] })
	DrawCards{N: len(ctx.opponent().Hand)}.Apply(ctx)
}

// SwitchInDamagedBench switches 1 of your opponent's damaged Benched Pokémon
// into the Active Spot (Cyrus). The old Active goes to that bench slot.
type SwitchInDamagedBench struct{}

func (SwitchInDamagedBench) Targets(ctx *EffectContext) []Spot {
	var out []Spot
	for _, s := range spotsWhere(ctx.g, 1-ctx.attacker, func(m *InPlay) bool { return m.Damage > 0 }) {
		if s.Pos > 0 {
			out = append(out, s)
		}
	}
	return out
}

func (SwitchInDamagedBench) Apply(ctx *EffectContext) {
	opp := ctx.opponent()
	slot := ctx.Target.Pos - 1
	old := opp.Active
	opp.Active, opp.Bench[slot] = opp.Bench[slot], old
	if old != nil {
		old.CureAll() // leaving the Active spot cures (C9)
	}
	ctx.g.narrate("%s's %s was switched into the Active Spot.", side(1-ctx.attacker), opp.Active.Card.Name)
}

// MoveBenchEnergyToActive moves an Energy from 1 of your Benched Pokémon to
// your Active (Dawn). With several types on it, the first type the Active's
// own type matches moves, else the first in energyOrder.
type MoveBenchEnergyToActive struct{}

func (MoveBenchEnergyToActive) Targets(ctx *EffectContext) []Spot {
	if ctx.Player().Active == nil {
		return nil
	}
	var out []Spot
	for _, s := range spotsWhere(ctx.g, ctx.attacker, func(m *InPlay) bool { return m.totalEnergy() > 0 }) {
		if s.Pos > 0 {
			out = append(out, s)
		}
	}
	return out
}

func (MoveBenchEnergyToActive) Apply(ctx *EffectContext) {
	from, to := ctx.targetMon(), ctx.Player().Active
	e := to.Card.Type
	if from.Energy[e] == 0 {
		for _, e = range energyOrder {
			if from.Energy[e] > 0 {
				break
			}
		}
	}
	from.Energy[e]--
	to.Energy[e]++
	ctx.g.narrate("%s energy moved from %s to %s.", e, from.Card.Name, to.Card.Name)
}

// MoveDamageToOppActive moves up to Amount damage from 1 of your damaged
// Pokémon named in Names to your opponent's Active (Acerola). A knockout
// scores as usual.
type MoveDamageToOppActive struct {
	Amount int
	Names  []string
}

func (v MoveDamageToOppActive) Targets(ctx *EffectContext) []Spot {
	if ctx.opponent().Active == nil {
		return nil
	}
	return spotsWhere(ctx.g, ctx.attacker, func(m *InPlay) bool { return m.Damage > 0 && named(v.Names)(m) })
}

func (v MoveDamageToOppActive) Apply(ctx *EffectContext) {
	from, to := ctx.targetMon(), ctx.opponent().Active
	n := min(v.Amount, from.Damage)
	from.Damage -= n
	to.Damage += n
	ctx.g.narrate("%d damage moved from %s to %s.", n, from.Card.Name, to.Card.Name)
	ctx.g.resolveAction()
}

// toolSpots is every spot of player p's Pokémon with a Tool.
func toolSpots(g *Game, p int) []Spot {
	return spotsWhere(g, p, func(m *InPlay) bool { return m.Tool != nil })
}

// DiscardOppTools discards every Tool on your opponent's Pokémon (Guzma).
type DiscardOppTools struct{}

func (DiscardOppTools) CanPlay(ctx *EffectContext) error {
	if len(toolSpots(ctx.g, 1-ctx.attacker)) == 0 {
		return errNoEffect
	}
	return nil
}

func (DiscardOppTools) Apply(ctx *EffectContext) {
	for _, s := range toolSpots(ctx.g, 1-ctx.attacker) {
		DiscardTool{}.Apply(&EffectContext{g: ctx.g, attacker: ctx.attacker, Target: s})
	}
}

// ReturnAllTools returns every Tool in play to its owner's hand (Elesa). A
// Tool whose owner's hand is full stays attached (T3).
type ReturnAllTools struct{}

func (ReturnAllTools) CanPlay(ctx *EffectContext) error {
	if len(toolSpots(ctx.g, 0))+len(toolSpots(ctx.g, 1)) == 0 {
		return errNoEffect
	}
	return nil
}

func (ReturnAllTools) Apply(ctx *EffectContext) {
	for p := 0; p < 2; p++ {
		pl := &ctx.g.S.Players[p]
		for _, s := range toolSpots(ctx.g, p) {
			mon := targetMon(pl, s.Pos)
			if len(pl.Hand) >= handMax {
				continue
			}
			pl.Hand = append(pl.Hand, *mon.Tool)
			ctx.g.narrate("%s returned to %s's hand.", mon.Tool.Name, side(p))
			mon.Tool = nil
		}
	}
}

// FlipSearch flips a coin and searches the deck for a random card matching
// Heads or Tails (Arven).
type FlipSearch struct{ Heads, Tails CardMatch }

func (FlipSearch) CanPlay(ctx *EffectContext) error { return handFull(ctx) }

func (v FlipSearch) Apply(ctx *EffectContext) {
	m := v.Tails
	if ctx.Flip() {
		m = v.Heads
	}
	SearchDeck{Match: m}.Apply(ctx)
}
