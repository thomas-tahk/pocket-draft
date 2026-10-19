package engine

import (
	"errors"
	"fmt"
	"strings"
)

// Effect ops for Supporter texts, part 3 (issue #71). Each is one verb; the
// server maps printed wordings onto them (server/texts_3_11.go). A Fits
// field says which cards the text names ("your Miltank", "your Stage 2
// Pokémon"); nil means any.

func fits(f func(Card) bool, c Card) bool { return f == nil || f(c) }

// ownSpots is each of player p's Pokémon in play that ok accepts.
func ownSpots(g *Game, p int, ok func(*InPlay) bool) []Spot {
	var out []Spot
	for _, pos := range targets() {
		if mon := targetMon(&g.S.Players[p], pos); mon != nil && ok(mon) {
			out = append(out, Spot{p, pos})
		}
	}
	return out
}

// energyUnits lists mon's Energy one entry per unit, in energyOrder.
func energyUnits(mon *InPlay) []Energy {
	var out []Energy
	for _, e := range energyOrder {
		for i := 0; i < mon.Energy[e]; i++ {
			out = append(out, e)
		}
	}
	return out
}

// discardRandomEnergy discards one random Energy from owner's mon.
func (g *Game) discardRandomEnergy(owner int, mon *InPlay) {
	units := energyUnits(mon)
	if len(units) == 0 {
		return
	}
	e := units[g.rng.intn(len(units))]
	mon.Energy[e]--
	pl := &g.S.Players[owner]
	pl.DiscardedEnergy = append(pl.DiscardedEnergy, e)
	g.narrate("%s Energy was discarded from %s's %s.", e, side(owner), mon.Card.Name)
}

var errNoEnergyToDiscard = errors.New("your opponent's Active Pokémon has no Energy")

// NeedInPlay: "You can use this card only if you have Name in play." (Piers)
type NeedInPlay struct{ Name string }

func (NeedInPlay) Apply(*EffectContext) {}

func (v NeedInPlay) CanPlay(ctx *EffectContext) error {
	if len(ownSpots(ctx.g, ctx.attacker, func(m *InPlay) bool { return m.Card.Name == v.Name })) == 0 {
		return fmt.Errorf("you need %s in play", v.Name)
	}
	return nil
}

// NeedActiveAttack: "only if your Pokémon in the Active Spot has the Attack
// attack." (Psychic)
type NeedActiveAttack struct{ Attack string }

func (NeedActiveAttack) Apply(*EffectContext) {}

func (v NeedActiveAttack) CanPlay(ctx *EffectContext) error {
	if a := ctx.attackerMon(); a != nil {
		for _, atk := range a.Card.Attacks {
			if atk.Name == v.Attack {
				return nil
			}
		}
	}
	return fmt.Errorf("your Active Pokémon needs the %s attack", v.Attack)
}

// DiscardOppEnergyN discards N random Energy from the opponent's Active
// (Piers; Team Rocket Grunt per heads). It needs Energy there to play.
type DiscardOppEnergyN struct{ N int }

func (v DiscardOppEnergyN) CanPlay(ctx *EffectContext) error {
	if ctx.defenderMon().totalEnergy() == 0 {
		return errNoEnergyToDiscard
	}
	return nil
}

func (v DiscardOppEnergyN) Apply(ctx *EffectContext) {
	for i := 0; i < v.N; i++ {
		ctx.g.discardRandomEnergy(1-ctx.attacker, ctx.defenderMon())
	}
}

// PerHeadsUntilTails flips a coin until tails and runs Op once per heads
// (Team Rocket Grunt). Op's CanPlay is the card's.
type PerHeadsUntilTails struct{ Op EffectOp }

func (v PerHeadsUntilTails) CanPlay(ctx *EffectContext) error {
	if gate, ok := v.Op.(Gated); ok {
		return gate.CanPlay(ctx)
	}
	return nil
}

func (v PerHeadsUntilTails) Apply(ctx *EffectContext) {
	heads := 0
	for ctx.Flip() {
		heads++
		v.Op.Apply(ctx)
	}
	ctx.g.narrate("%s flipped %d heads before tails.", side(ctx.attacker), heads)
}

// Heal heals Amount damage from one of your Pokémon that Fits and cures it
// of Cures. It targets those with damage or one of those conditions
// (Pokémon Center Lady, Whitney).
type Heal struct {
	Amount int
	Cures  []Status
	Fits   func(Card) bool
}

func (v Heal) needs(m *InPlay) bool {
	if !fits(v.Fits, m.Card) {
		return false
	}
	if m.Damage > 0 {
		return true
	}
	for _, s := range v.Cures {
		if m.Has(s) {
			return true
		}
	}
	return false
}

func (v Heal) Targets(ctx *EffectContext) []Spot { return ownSpots(ctx.g, ctx.attacker, v.needs) }

func (v Heal) Apply(ctx *EffectContext) {
	mon := ctx.targetMon()
	healed := min(v.Amount, mon.Damage)
	mon.Damage -= healed
	for _, s := range v.Cures {
		mon.Cure(s)
	}
	ctx.g.narrate("%s's %s healed %d.", side(ctx.attacker), mon.Card.Name, healed)
}

// AttachChosenFromDiscard attaches up to N Energy from your discard pile to one of
// your Pokémon that Fits. Type limits it to one type (NoEnergy: any);
// Distinct takes each type at most once. The oldest discarded go first
// (Volkner, Professor Sada).
type AttachChosenFromDiscard struct {
	N        int
	Type     Energy
	Distinct bool
	Fits     func(Card) bool
}

// pick is the index in pl.DiscardedEnergy of each Energy to attach.
func (v AttachChosenFromDiscard) pick(pl *Player) []int {
	var out []int
	taken := map[Energy]bool{}
	for i, e := range pl.DiscardedEnergy {
		if len(out) == v.N {
			break
		}
		if (v.Type != NoEnergy && e != v.Type) || (v.Distinct && taken[e]) {
			continue
		}
		taken[e] = true
		out = append(out, i)
	}
	return out
}

func (v AttachChosenFromDiscard) Targets(ctx *EffectContext) []Spot {
	if len(v.pick(ctx.Player())) == 0 {
		return nil
	}
	return ownSpots(ctx.g, ctx.attacker, func(m *InPlay) bool { return fits(v.Fits, m.Card) })
}

func (v AttachChosenFromDiscard) Apply(ctx *EffectContext) {
	pl, mon := ctx.Player(), ctx.targetMon()
	idx := v.pick(pl)
	used := map[int]bool{}
	for _, i := range idx {
		used[i] = true
		mon.Energy[pl.DiscardedEnergy[i]]++
	}
	var left []Energy
	for i, e := range pl.DiscardedEnergy {
		if !used[i] {
			left = append(left, e)
		}
	}
	pl.DiscardedEnergy = left
	ctx.g.narrate("%s attached %d Energy from the discard pile to %s.", side(ctx.attacker), len(idx), mon.Card.Name)
}

// AttachNewEnergy attaches one Type Energy (not this turn's) to one of your
// Pokémon that Fits (Wally: a [C] Energy from your Energy Zone).
type AttachNewEnergy struct {
	Type Energy
	Fits func(Card) bool
}

func (v AttachNewEnergy) Targets(ctx *EffectContext) []Spot {
	return ownSpots(ctx.g, ctx.attacker, func(m *InPlay) bool { return fits(v.Fits, m.Card) })
}

func (v AttachNewEnergy) Apply(ctx *EffectContext) {
	mon := ctx.targetMon()
	mon.Energy[v.Type]++
	ctx.g.narrate("%s attached %s Energy to %s.", side(ctx.attacker), v.Type, mon.Card.Name)
}

// ShuffleIntoDeck shuffles one of your Pokémon that Fits into your deck; its
// Tool and Energy are discarded. Your Active may go only when a Benched
// Pokémon can replace it, and you then choose which (Professor Turo).
type ShuffleIntoDeck struct{ Fits func(Card) bool }

func (v ShuffleIntoDeck) Targets(ctx *EffectContext) []Spot {
	pl := ctx.Player()
	return ownSpots(ctx.g, ctx.attacker, func(m *InPlay) bool {
		return fits(v.Fits, m.Card) && (m != pl.Active || pl.BenchCount() > 0)
	})
}

func (v ShuffleIntoDeck) Apply(ctx *EffectContext) {
	pl, mon := ctx.Player(), ctx.targetMon()
	pl.DiscardedEnergy = append(pl.DiscardedEnergy, energyUnits(mon)...)
	if mon.Tool != nil {
		pl.Discard = append(pl.Discard, *mon.Tool)
	}
	pl.Deck = append(pl.Deck, mon.Card)
	ctx.g.shuffleDeck(ctx.attacker)
	ctx.g.narrate("%s shuffled %s into their deck.", side(ctx.attacker), mon.Card.Name)
	if ctx.Target.Pos > 0 {
		pl.Bench[ctx.Target.Pos-1] = nil
		return
	}
	pl.Active = nil
	ctx.g.ask(ctx.attacker, PromptNewActive)
}

// MoveOppEnergyToActive moves a random Energy from one of the opponent's
// Benched Pokémon to their Active (Psychic).
type MoveOppEnergyToActive struct{}

func (MoveOppEnergyToActive) Targets(ctx *EffectContext) []Spot {
	var out []Spot
	for _, s := range ownSpots(ctx.g, 1-ctx.attacker, func(m *InPlay) bool { return m.totalEnergy() > 0 }) {
		if s.Pos > 0 {
			out = append(out, s)
		}
	}
	return out
}

func (MoveOppEnergyToActive) Apply(ctx *EffectContext) {
	from, to := ctx.targetMon(), ctx.defenderMon()
	units := energyUnits(from)
	e := units[ctx.g.rng.intn(len(units))]
	from.Energy[e]--
	to.Energy[e]++
	ctx.g.narrate("%s Energy moved from %s to %s.", e, from.Card.Name, to.Card.Name)
}

// TopCardsToHand looks at the top N cards of your deck, puts every one that
// Fits into your hand (up to 10 in hand) and shuffles the rest back
// (Sightseer, Traveling Merchant, Puppy-Loving Girl).
type TopCardsToHand struct {
	N    int
	Fits func(Card) bool
}

func (TopCardsToHand) CanPlay(ctx *EffectContext) error { return handFull(ctx) }

func (v TopCardsToHand) Apply(ctx *EffectContext) {
	pl := ctx.Player()
	n := min(v.N, len(pl.Deck))
	rest := append([]Card(nil), pl.Deck[n:]...)
	var found []string
	for _, c := range pl.Deck[:n] {
		if fits(v.Fits, c) && len(pl.Hand) < handMax {
			pl.Hand = append(pl.Hand, c)
			found = append(found, c.Name)
		} else {
			rest = append(rest, c)
		}
	}
	pl.Deck = rest
	ctx.g.shuffleDeck(ctx.attacker)
	ctx.g.narrate("%s found %d: %s.", side(ctx.attacker), len(found), strings.Join(found, ", "))
}

// SwitchOppActive switches the opponent's Active to the Bench; the opponent
// chooses the new Active (Sabrina). Leaving the Active clears its conditions
// (spec C9).
type SwitchOppActive struct{}

func (SwitchOppActive) CanPlay(ctx *EffectContext) error {
	if ctx.g.S.Players[1-ctx.attacker].BenchCount() == 0 {
		return errors.New("your opponent has no Benched Pokémon")
	}
	return nil
}

func (SwitchOppActive) Apply(ctx *EffectContext) {
	opp := &ctx.g.S.Players[1-ctx.attacker]
	opp.Active.CureAll()
	ctx.g.S.switchedOut, opp.Active = opp.Active, nil
	ctx.g.ask(1-ctx.attacker, PromptNewActive)
}

// SwitchOwnActive switches your Active, which must Fit, with the targeted
// Benched Pokémon (Skyla: your Active Stage 1 Pokémon).
type SwitchOwnActive struct{ Fits func(Card) bool }

func (v SwitchOwnActive) CanPlay(ctx *EffectContext) error {
	if a := ctx.attackerMon(); a == nil || !fits(v.Fits, a.Card) {
		return errors.New("your Active Pokémon can't be switched by this card")
	}
	return nil
}

func (v SwitchOwnActive) Targets(ctx *EffectContext) []Spot {
	var out []Spot
	for _, s := range ownSpots(ctx.g, ctx.attacker, func(*InPlay) bool { return true }) {
		if s.Pos > 0 {
			out = append(out, s)
		}
	}
	return out
}

func (v SwitchOwnActive) Apply(ctx *EffectContext) {
	pl, slot := ctx.Player(), ctx.Target.Pos-1
	pl.Active.CureAll()
	pl.Active, pl.Bench[slot] = pl.Bench[slot], pl.Active
	ctx.g.narrate("%s switched in %s.", side(ctx.attacker), pl.Active.Card.Name)
}

// ShuffleOppSupporter: the opponent reveals their hand; a Supporter found
// there is shuffled into their deck (Silver). The first one in their hand is
// taken: there is no card picker yet.
type ShuffleOppSupporter struct{}

func (ShuffleOppSupporter) Apply(ctx *EffectContext) {
	opp := 1 - ctx.attacker
	pl := &ctx.g.S.Players[opp]
	names := make([]string, len(pl.Hand))
	for i, c := range pl.Hand {
		names[i] = c.Name
	}
	ctx.g.narrate("%s revealed their hand: %s.", side(opp), strings.Join(names, ", "))
	for i, c := range pl.Hand {
		if c.Kind == KindSupporter {
			pl.Hand = removeAt(pl.Hand, i)
			pl.Deck = append(pl.Deck, c)
			ctx.g.shuffleDeck(opp)
			ctx.g.narrate("%s was shuffled into %s's deck.", c.Name, side(opp))
			return
		}
	}
}

// BenchOppBasics puts the Basic Pokémon in the opponent's hand onto their
// Bench, as many as fit, left to right (Team Rocket's Boss: "any number";
// all of them until a picker exists).
type BenchOppBasics struct{}

func (BenchOppBasics) CanPlay(ctx *EffectContext) error {
	pl := &ctx.g.S.Players[1-ctx.attacker]
	if pl.BenchCount() == BenchSlots {
		return errors.New("your opponent's Bench is full")
	}
	return nil
}

func (BenchOppBasics) Apply(ctx *EffectContext) {
	opp := 1 - ctx.attacker
	pl := &ctx.g.S.Players[opp]
	var keep []Card
	for _, c := range pl.Hand {
		slot := freeSlot(pl)
		if !c.IsBasicPokemon() || slot < 0 {
			keep = append(keep, c)
			continue
		}
		pl.Bench[slot] = newInPlay(c, ctx.g.S.Turn)
		ctx.g.narrate("%s's %s was put onto the Bench.", side(opp), c.Name)
	}
	pl.Hand = keep
}

func freeSlot(pl *Player) int {
	for s := 0; s < BenchSlots; s++ {
		if pl.Bench[s] == nil {
			return s
		}
	}
	return -1
}

// SelfStatusOnTails flips a coin; tails, your Active gets Status (Team
// Rocket's Master Plan).
type SelfStatusOnTails struct{ Status Status }

func (v SelfStatusOnTails) Apply(ctx *EffectContext) {
	if ctx.Flip() || !ctx.g.canBeAffected(ctx.attacker, 0) {
		return
	}
	ctx.attackerMon().Inflict(v.Status)
	ctx.g.narrate("%s's %s is now %s.", side(ctx.attacker), ctx.attackerMon().Card.Name, v.Status)
}

// DiscardEnergyAmong discards one random Energy from among all the Energy on
// the opponent's Pokémon that Fit (Team Star Grunt: those with an Ability).
type DiscardEnergyAmong struct{ Fits func(Card) bool }

func (v DiscardEnergyAmong) units(ctx *EffectContext) (mons []*InPlay) {
	opp := 1 - ctx.attacker
	for _, s := range ownSpots(ctx.g, opp, func(m *InPlay) bool { return fits(v.Fits, m.Card) }) {
		mon := targetMon(&ctx.g.S.Players[opp], s.Pos)
		for range energyUnits(mon) {
			mons = append(mons, mon)
		}
	}
	return mons
}

func (v DiscardEnergyAmong) CanPlay(ctx *EffectContext) error {
	if len(v.units(ctx)) == 0 {
		return errors.New("no Energy to discard")
	}
	return nil
}

// Apply picks a unit uniformly among all of them, then a random one of that
// Pokémon's: the same as one draw over every unit.
func (v DiscardEnergyAmong) Apply(ctx *EffectContext) {
	mons := v.units(ctx)
	ctx.g.discardRandomEnergy(1-ctx.attacker, mons[ctx.g.rng.intn(len(mons))])
}

// EvolveFitFromDeck evolves one of your Pokémon that Fits with a random card
// from your deck that Evolves into it (Wallace). The usual timing applies: not
// on a first turn, nor the turn the Pokémon was played or evolved.
type EvolveFitFromDeck struct {
	Fits    func(Card) bool
	Evolves func(Card) bool
}

func (v EvolveFitFromDeck) evolutions(pl *Player, mon *InPlay) []int {
	var out []int
	for i, c := range pl.Deck {
		if c.Kind == KindPokemon && c.EvolvesFrom == mon.Card.Name && fits(v.Evolves, c) {
			out = append(out, i)
		}
	}
	return out
}

func (v EvolveFitFromDeck) Targets(ctx *EffectContext) []Spot {
	pl := ctx.Player()
	g := ctx.g
	return ownSpots(g, ctx.attacker, func(m *InPlay) bool {
		return fits(v.Fits, m.Card) && len(v.evolutions(pl, m)) > 0 && g.S.Turn > 2 && m.PlacedTurn != g.S.Turn
	})
}

func (v EvolveFitFromDeck) Apply(ctx *EffectContext) {
	pl, mon := ctx.Player(), ctx.targetMon()
	opts := v.evolutions(pl, mon)
	i := opts[ctx.g.rng.intn(len(opts))]
	ev := pl.Deck[i]
	pl.Deck = removeAt(pl.Deck, i)
	mon.Card, mon.PlacedTurn = ev, ctx.g.S.Turn
	mon.CureAll() // evolving cures (spec C9)
	ctx.g.narrate("%s evolved into %s.", side(ctx.attacker), ev.Name)
}
