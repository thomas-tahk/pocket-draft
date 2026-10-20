package engine

import (
	"errors"
	"fmt"
	"strings"
)

// Item texts (issue #72): the EffectOps Item cards are built from. Each op is
// one verb; server/texts_3_12.go maps each printed wording to its ops.

var (
	errOppNoEnergy = errors.New("your opponent's Active Pokémon has no Energy to discard")
	errNoMatch     = errors.New("there is no card it can take")
	errDeckEmpty   = errors.New("your deck is empty")
)

// ownSpots is every spot of player p's that holds a Pokémon passing ok.
func (g *Game) ownSpots(p int, ok func(Spot, *InPlay) bool) []Spot {
	var out []Spot
	for _, pos := range targets() {
		s := Spot{p, pos}
		if mon := targetMon(&g.S.Players[p], pos); mon != nil && ok(s, mon) {
			out = append(out, s)
		}
	}
	return out
}

func (ctx *EffectContext) opp() *Player { return &ctx.g.S.Players[1-ctx.attacker] }

func (g *Game) heal(mon *InPlay, n int) int {
	healed := min(n, mon.Damage)
	mon.Damage -= healed
	return healed
}

// HealTarget heals N damage from the targeted Pokémon of yours (Potion).
type HealTarget struct{ N int }

// Targets is each of your damaged Pokémon.
func (HealTarget) Targets(ctx *EffectContext) []Spot {
	return ctx.g.ownSpots(ctx.attacker, func(_ Spot, m *InPlay) bool { return m.Damage > 0 })
}

func (v HealTarget) Apply(ctx *EffectContext) {
	mon := ctx.targetMon()
	ctx.g.narrate("%s healed %d damage from %s.", ctx.Card.Name, ctx.g.heal(mon, v.N), mon.Card.Name)
}

// HealActive heals N damage from your Active. With FlipToReturn, healing any
// damage flips a coin, and heads puts the played card back in your hand
// (Lucky Ice Pop).
type HealActive struct {
	N            int
	FlipToReturn bool
}

func (v HealActive) Apply(ctx *EffectContext) {
	mon := ctx.attackerMon()
	healed := ctx.g.heal(mon, v.N)
	ctx.g.narrate("%s healed %d damage from %s.", ctx.Card.Name, healed, mon.Card.Name)
	if !v.FlipToReturn || healed == 0 {
		return
	}
	if !ctx.Flip() {
		ctx.g.narrate("Tails: %s stays in the discard pile.", ctx.Card.Name)
		return
	}
	pl := ctx.Player()
	if i := lastIndexOf(pl.Discard, ctx.Card.ID); i >= 0 {
		pl.Discard = removeAt(pl.Discard, i)
		pl.Hand = append(pl.Hand, ctx.Card)
		ctx.g.narrate("Heads: %s went back into %s's hand.", ctx.Card.Name, side(ctx.attacker))
	}
}

func lastIndexOf(cards []Card, id string) int {
	for i := len(cards) - 1; i >= 0; i-- {
		if cards[i].ID == id {
			return i
		}
	}
	return -1
}

// CureRandomCondition removes one random Special Condition from your Active
// (Big Malasada).
type CureRandomCondition struct{}

func (CureRandomCondition) Apply(ctx *EffectContext) {
	mon := ctx.attackerMon()
	conds := mon.Conditions()
	if len(conds) == 0 {
		return
	}
	s := conds[ctx.g.rng.intn(len(conds))]
	mon.Cure(s)
	ctx.g.narrate("%s is no longer %s.", mon.Card.Name, s)
}

// OppShuffleHandDraw: your opponent shuffles their hand into their deck and
// draws N cards (Red Card).
type OppShuffleHandDraw struct{ N int }

func (v OppShuffleHandDraw) Apply(ctx *EffectContext) {
	opp := ctx.opp()
	opp.Deck = append(opp.Deck, opp.Hand...)
	opp.Hand = nil
	ctx.g.shuffleDeck(1 - ctx.attacker)
	for i := 0; i < v.N && len(opp.Deck) > 0; i++ {
		opp.Hand = append(opp.Hand, opp.Deck[0])
		opp.Deck = opp.Deck[1:]
	}
	ctx.g.narrate("%s shuffled their hand into their deck and drew %d.", side(1-ctx.attacker), len(opp.Hand))
}

// DiscardOppEnergy discards one Energy from your opponent's Active: one of
// Type, or a random one when Type is NoEnergy. With BothHeads it first flips
// 2 coins and discards only on two heads (Hitting Hammer).
type DiscardOppEnergy struct {
	Type      Energy
	BothHeads bool
}

func (v DiscardOppEnergy) CanPlay(ctx *EffectContext) error {
	if len(v.choices(ctx.defenderMon())) == 0 {
		return errOppNoEnergy
	}
	return nil
}

// choices lists every Energy on mon that may be discarded, one entry each.
func (v DiscardOppEnergy) choices(mon *InPlay) []Energy {
	var out []Energy
	for _, e := range energyOrder {
		if v.Type != NoEnergy && e != v.Type {
			continue
		}
		for i := 0; i < mon.Energy[e]; i++ {
			out = append(out, e)
		}
	}
	return out
}

func (v DiscardOppEnergy) Apply(ctx *EffectContext) {
	if v.BothHeads {
		a, b := ctx.Flip(), ctx.Flip()
		if !a || !b {
			ctx.g.narrate("Not both heads: nothing was discarded.")
			return
		}
	}
	mon := ctx.defenderMon()
	es := v.choices(mon)
	e := es[ctx.g.rng.intn(len(es))]
	mon.Energy[e]--
	opp := ctx.opp()
	opp.DiscardedEnergy = append(opp.DiscardedEnergy, e)
	ctx.g.narrate("%s discarded a %s Energy from %s.", ctx.Card.Name, e, mon.Card.Name)
}

// AttachDiscardedEnergy attaches a Type Energy from your discard pile to your
// Active, if it is a Type Pokémon (Flame Patch).
type AttachDiscardedEnergy struct{ Type Energy }

func (v AttachDiscardedEnergy) CanPlay(ctx *EffectContext) error {
	if ctx.attackerMon().Card.Type != v.Type {
		return fmt.Errorf("your Active Pokémon isn't a %s Pokémon", v.Type)
	}
	if indexOfEnergy(ctx.Player().DiscardedEnergy, v.Type) < 0 {
		return fmt.Errorf("there is no %s Energy in your discard pile", v.Type)
	}
	return nil
}

func indexOfEnergy(es []Energy, t Energy) int {
	for i, e := range es {
		if e == t {
			return i
		}
	}
	return -1
}

func (v AttachDiscardedEnergy) Apply(ctx *EffectContext) {
	pl := ctx.Player()
	i := indexOfEnergy(pl.DiscardedEnergy, v.Type)
	pl.DiscardedEnergy = append(pl.DiscardedEnergy[:i:i], pl.DiscardedEnergy[i+1:]...)
	ctx.attackerMon().Energy[v.Type]++
	ctx.g.narrate("%s attached a %s Energy to %s.", ctx.Card.Name, v.Type, ctx.attackerMon().Card.Name)
}

// RandomFromDiscard puts a random card that Match accepts from a discard pile
// into your hand: yours, or your opponent's when FromOpp (Fishing Net, Team
// Rocket's Thieving Machine).
type RandomFromDiscard struct {
	FromOpp bool
	Match   func(Card) bool
}

func (v RandomFromDiscard) pile(ctx *EffectContext) *Player {
	if v.FromOpp {
		return ctx.opp()
	}
	return ctx.Player()
}

func (v RandomFromDiscard) matches(ctx *EffectContext) []int {
	var out []int
	for i, c := range v.pile(ctx).Discard {
		if v.Match(c) {
			out = append(out, i)
		}
	}
	return out
}

func (v RandomFromDiscard) CanPlay(ctx *EffectContext) error {
	if err := handFull(ctx); err != nil {
		return err
	}
	if len(v.matches(ctx)) == 0 {
		return errNoMatch
	}
	return nil
}

func (v RandomFromDiscard) Apply(ctx *EffectContext) {
	m := v.matches(ctx)
	pile := v.pile(ctx)
	i := m[ctx.g.rng.intn(len(m))]
	c := pile.Discard[i]
	pile.Discard = removeAt(pile.Discard, i)
	ctx.Player().Hand = append(ctx.Player().Hand, c)
	ctx.g.narrate("%s put %s into %s's hand.", ctx.Card.Name, c.Name, side(ctx.attacker))
}

// ShuffleRandomHandCard: a random card from both players' hands is revealed
// and shuffled into its owner's deck (Prank Spinner).
type ShuffleRandomHandCard struct{}

func (ShuffleRandomHandCard) CanPlay(ctx *EffectContext) error {
	// The played card is still in hand while this is asked.
	if len(ctx.Player().Hand)-1+len(ctx.opp().Hand) == 0 {
		return errors.New("neither player has a card in hand")
	}
	return nil
}

func (ShuffleRandomHandCard) Apply(ctx *EffectContext) {
	mine, theirs := ctx.Player(), ctx.opp()
	i := ctx.g.rng.intn(len(mine.Hand) + len(theirs.Hand))
	owner, pl := ctx.attacker, mine
	if i >= len(mine.Hand) {
		owner, pl, i = 1-ctx.attacker, theirs, i-len(mine.Hand)
	}
	c := pl.Hand[i]
	pl.Hand = removeAt(pl.Hand, i)
	pl.Deck = append(pl.Deck, c)
	ctx.g.shuffleDeck(owner)
	ctx.g.narrate("%s revealed %s from %s's hand and shuffled it into their deck.", ctx.Card.Name, c.Name, side(owner))
}

// TopCardIf looks at the top card of your deck: if Match accepts it, it goes
// into your hand, otherwise to the bottom of your deck (Mythical Slab).
type TopCardIf struct{ Match func(Card) bool }

func (TopCardIf) CanPlay(ctx *EffectContext) error {
	if len(ctx.Player().Deck) == 0 {
		return errDeckEmpty
	}
	return handFull(ctx)
}

func (v TopCardIf) Apply(ctx *EffectContext) {
	pl := ctx.Player()
	c := pl.Deck[0]
	pl.Deck = pl.Deck[1:]
	if v.Match(c) {
		pl.Hand = append(pl.Hand, c)
		ctx.g.narrate("%s put %s into %s's hand.", ctx.Card.Name, c.Name, side(ctx.attacker))
		return
	}
	pl.Deck = append(pl.Deck, c)
	ctx.g.narrate("%s put the top card of %s's deck on the bottom.", ctx.Card.Name, side(ctx.attacker))
}

// RevealOppHand: your opponent reveals their hand (Hand Scope).
type RevealOppHand struct{}

func (RevealOppHand) Apply(ctx *EffectContext) {
	hand := ctx.opp().Hand
	if len(hand) == 0 {
		ctx.g.narrate("%s revealed an empty hand.", side(1-ctx.attacker))
		return
	}
	names := make([]string, len(hand))
	for i, c := range hand {
		names[i] = c.Name
	}
	ctx.g.narrate("%s revealed their hand: %s.", side(1-ctx.attacker), strings.Join(names, ", "))
}

// FlipForBenchEnergy flips a coin; heads attaches a Type Energy from your
// Energy Zone to the targeted Benched Type Pokémon (Electric Generator). The
// zone's energy for the turn is not used.
type FlipForBenchEnergy struct{ Type Energy }

// Targets is each of your Benched Type Pokémon.
func (v FlipForBenchEnergy) Targets(ctx *EffectContext) []Spot {
	return ctx.g.ownSpots(ctx.attacker, func(s Spot, m *InPlay) bool { return s.Pos > 0 && m.Card.Type == v.Type })
}

func (v FlipForBenchEnergy) Apply(ctx *EffectContext) {
	if !ctx.Flip() {
		ctx.g.narrate("Tails: no Energy was attached.")
		return
	}
	mon := ctx.targetMon()
	mon.Energy[v.Type]++
	ctx.g.narrate("Heads: %s attached a %s Energy to %s.", ctx.Card.Name, v.Type, mon.Card.Name)
}

// MoveEnergyToActive moves one Energy of any of Types from the targeted
// Benched Pokémon of yours to your Active (Elemental Switch). With more than
// one of the types there, you choose which.
type MoveEnergyToActive struct{ Types []Energy }

func (v MoveEnergyToActive) present(m *InPlay) []Energy {
	var out []Energy
	for _, t := range v.Types {
		if m.Energy[t] > 0 {
			out = append(out, t)
		}
	}
	return out
}

// Targets is each of your Benched Pokémon with one of Types attached.
func (v MoveEnergyToActive) Targets(ctx *EffectContext) []Spot {
	return ctx.g.ownSpots(ctx.attacker, func(s Spot, m *InPlay) bool { return s.Pos > 0 && len(v.present(m)) > 0 })
}

func (v MoveEnergyToActive) Apply(ctx *EffectContext) {
	from := ctx.targetMon()
	types := v.present(from)
	labels := make([]string, len(types))
	for i, t := range types {
		labels[i] = t.String() + " Energy"
	}
	p, name := ctx.attacker, ctx.Card.Name
	ctx.g.askChoice(p, "Choose the Energy to move", labels, func(g *Game, i int) {
		to := g.S.Players[p].Active
		from.Energy[types[i]]--
		to.Energy[types[i]]++
		g.narrate("%s moved a %s Energy from %s to %s.", name, types[i], from.Card.Name, to.Card.Name)
	})
}

// OnHeads flips a coin; heads runs Op (Order Pad). The card is gated as Op is.
type OnHeads struct{ Op EffectOp }

func (v OnHeads) CanPlay(ctx *EffectContext) error {
	if g, ok := v.Op.(Gated); ok {
		return g.CanPlay(ctx)
	}
	return nil
}

func (v OnHeads) Apply(ctx *EffectContext) {
	if !ctx.Flip() {
		ctx.g.narrate("Tails: %s did nothing.", ctx.Card.Name)
		return
	}
	v.Op.Apply(ctx)
}

// pokemonIn is the index of each Pokémon card among cards.
func pokemonIn(cards []Card) []int {
	var out []int
	for i, c := range cards {
		if c.Kind == KindPokemon {
			out = append(out, i)
		}
	}
	return out
}

// SwapHandPokemon: you choose a Pokémon in your hand and switch it with a
// random Pokémon in your deck (Pokémon Communication).
type SwapHandPokemon struct{}

func (SwapHandPokemon) CanPlay(ctx *EffectContext) error {
	pl := ctx.Player()
	if len(pokemonIn(pl.Hand)) == 0 {
		return errors.New("you have no Pokémon in your hand")
	}
	if len(pokemonIn(pl.Deck)) == 0 {
		return errors.New("there is no Pokémon in your deck")
	}
	return nil
}

func (SwapHandPokemon) Apply(ctx *EffectContext) {
	p, idx := ctx.attacker, pokemonIn(ctx.Player().Hand)
	ctx.g.askChoice(p, "Choose a Pokémon to switch with your deck", handLabelsAt(ctx.Player().Hand, idx), func(g *Game, i int) {
		pl := &g.S.Players[p]
		inDeck := pokemonIn(pl.Deck)
		d := inDeck[g.rng.intn(len(inDeck))]
		mine := pl.Hand[idx[i]]
		pl.Hand[idx[i]], pl.Deck[d] = pl.Deck[d], mine
		g.shuffleDeck(p)
		g.narrate("%s switched %s with a Pokémon from their deck.", side(p), mine.Name)
	})
}

func handLabelsAt(hand []Card, idx []int) []string {
	out := make([]string, len(idx))
	for i, j := range idx {
		out[i] = hand[j].Name
	}
	return out
}

// EvolveFromDeck evolves the targeted Type Pokémon of yours with a random
// Type Pokémon from your deck that evolves from it. Not on your first turn,
// nor on a Pokémon put into play this turn (Quick-Grow Extract).
type EvolveFromDeck struct{ Type Energy }

func (EvolveFromDeck) CanPlay(ctx *EffectContext) error {
	if ctx.g.S.Turn <= 2 {
		return errFirstTurnCandy
	}
	return nil
}

func (v EvolveFromDeck) evolutions(ctx *EffectContext, mon *InPlay) []int {
	var out []int
	for i, c := range ctx.Player().Deck {
		if c.Kind == KindPokemon && c.Type == v.Type && c.EvolvesFrom == mon.Card.Name {
			out = append(out, i)
		}
	}
	return out
}

// Targets is each of your Type Pokémon not put into play this turn with an
// evolution for it in your deck.
func (v EvolveFromDeck) Targets(ctx *EffectContext) []Spot {
	return ctx.g.ownSpots(ctx.attacker, func(_ Spot, m *InPlay) bool {
		return m.Card.Type == v.Type && m.PlacedTurn != ctx.g.S.Turn && len(v.evolutions(ctx, m)) > 0
	})
}

func (v EvolveFromDeck) Apply(ctx *EffectContext) {
	mon, pl := ctx.targetMon(), ctx.Player()
	idx := v.evolutions(ctx, mon)
	i := idx[ctx.g.rng.intn(len(idx))]
	ev := pl.Deck[i]
	pl.Deck = removeAt(pl.Deck, i)
	ctx.g.evolveMon(ctx.attacker, mon, ev)
}
