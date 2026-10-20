package engine

import (
	"errors"
	"strings"
)

// Trainer verbs for Supporter texts (issue #70). Each is data: a filter says
// which Pokémon a text means, and the wording's ops are listed in
// trainerEffects (server).

// MonFilter picks the Pokémon a text means ("your [W] Pokémon", "your
// Carracosta or Jellicent"). Nil means any.
type MonFilter func(*InPlay) bool

func (f MonFilter) ok(m *InPlay) bool { return m != nil && (f == nil || f(m)) }

// Named matches Pokémon whose name is exactly one of names.
func Named(names ...string) MonFilter {
	return func(m *InPlay) bool {
		for _, n := range names {
			if m.Card.Name == n {
				return true
			}
		}
		return false
	}
}

// OfType matches Pokémon of type e.
func OfType(e Energy) MonFilter { return func(m *InPlay) bool { return m.Card.Type == e } }

// Damaged matches Pokémon with damage on them.
func Damaged(m *InPlay) bool { return m.Damage > 0 }

// Both matches Pokémon that pass every filter.
func Both(fs ...MonFilter) MonFilter {
	return func(m *InPlay) bool {
		for _, f := range fs {
			if !f.ok(m) {
				return false
			}
		}
		return true
	}
}

// ultraBeasts is every Ultra Beast in the card data, by name before " ex".
var ultraBeasts = []string{"Nihilego", "Buzzwole", "Pheromosa", "Xurkitree", "Celesteela", "Kartana",
	"Guzzlord", "Poipole", "Naganadel", "Stakataka", "Blacephalon"}

// IsUltraBeast reports whether c is an Ultra Beast (the scrape has no flag).
func (c Card) IsUltraBeast() bool {
	base := strings.TrimSuffix(c.Name, " ex")
	for _, n := range ultraBeasts {
		if base == n {
			return true
		}
	}
	return false
}

// Opponent is the side the acting player plays against.
func (ctx *EffectContext) Opponent() *Player { return &ctx.g.S.Players[1-ctx.attacker] }

// InPlay is every Pokémon the acting player has in play, Active first.
func (ctx *EffectContext) InPlay() []*InPlay { return inPlay(ctx.Player()) }

func inPlay(pl *Player) []*InPlay {
	var out []*InPlay
	for _, pos := range targets() {
		if m := targetMon(pl, pos); m != nil {
			out = append(out, m)
		}
	}
	return out
}

func (ctx *EffectContext) target() *InPlay {
	return targetMon(&ctx.g.S.Players[ctx.Target.Player], ctx.Target.Pos)
}

// spotsOf is player p's spots, from the given positions, whose Pokémon pass f.
func spotsOf(g *Game, p int, positions []int, f MonFilter) []Spot {
	var out []Spot
	for _, pos := range positions {
		if f.ok(targetMon(&g.S.Players[p], pos)) {
			out = append(out, Spot{p, pos})
		}
	}
	return out
}

func benchPositions() []int { return targets()[1:] }

// Require gates a Trainer: it can be played only while Test holds. Why is
// the reason shown when it can't (spec F18).
type Require struct {
	Test func(*EffectContext) bool
	Why  string
}

func (v Require) CanPlay(ctx *EffectContext) error {
	if v.Test(ctx) {
		return nil
	}
	return errors.New(v.Why)
}

func (Require) Apply(*EffectContext) {}

// HealOne heals Amount damage (0: all of it) from 1 of your damaged Pokémon
// that passes Which. DiscardEnergy then discards all its Energy (Mallow).
type HealOne struct {
	Amount        int
	Which         MonFilter
	DiscardEnergy bool
}

func (v HealOne) Targets(ctx *EffectContext) []Spot {
	return spotsOf(ctx.g, ctx.attacker, targets(), Both(Damaged, v.Which))
}

func (v HealOne) Apply(ctx *EffectContext) {
	m := ctx.target()
	heal(ctx.g, m, v.Amount)
	if v.DiscardEnergy {
		discardAllEnergy(ctx.Player(), m)
	}
}

// HealEach heals Amount damage from each of your damaged Pokémon that passes
// Which (Irida). It can't be played when none can be healed.
type HealEach struct {
	Amount int
	Which  MonFilter
}

func (v HealEach) CanPlay(ctx *EffectContext) error {
	if len(spotsOf(ctx.g, ctx.attacker, targets(), Both(Damaged, v.Which))) == 0 {
		return errors.New("none of your Pokémon can be healed")
	}
	return nil
}

func (v HealEach) Apply(ctx *EffectContext) {
	for _, s := range spotsOf(ctx.g, ctx.attacker, targets(), Both(Damaged, v.Which)) {
		heal(ctx.g, targetMon(ctx.Player(), s.Pos), v.Amount)
	}
}

func heal(g *Game, m *InPlay, amount int) {
	if g.healBlocked() {
		return
	}
	if amount == 0 || amount > m.Damage {
		amount = m.Damage
	}
	m.Damage -= amount
	g.narrate("%s healed %d damage.", m.Card.Name, amount)
}

func discardAllEnergy(pl *Player, m *InPlay) {
	for _, e := range energyOrder {
		for ; m.Energy[e] > 0; m.Energy[e]-- {
			pl.DiscardedEnergy = append(pl.DiscardedEnergy, e)
		}
	}
}

// ToHand puts 1 of your Pokémon that passes Which into your hand (Ilima,
// Parasol Lady); ActiveOnly limits it to the Active (Koga). Its Energy and
// Tool are discarded. The Active can be taken only while a Benched Pokémon
// can replace it, and its owner then promotes one.
type ToHand struct {
	Which      MonFilter
	ActiveOnly bool
}

func (v ToHand) Targets(ctx *EffectContext) []Spot {
	positions := targets()
	if v.ActiveOnly {
		positions = positions[:1]
	}
	if ctx.Player().BenchCount() == 0 {
		positions = positions[1:]
	}
	return spotsOf(ctx.g, ctx.attacker, positions, v.Which)
}

func (v ToHand) Apply(ctx *EffectContext) {
	pl := ctx.Player()
	m := ctx.target()
	discardAllEnergy(pl, m)
	if m.Tool != nil {
		pl.Discard = append(pl.Discard, *m.Tool)
	}
	pl.Hand = append(pl.Hand, m.Card)
	ctx.g.narrate("%s put %s into their hand.", side(ctx.attacker), m.Card.Name)
	if ctx.Target.Pos == 0 {
		pl.Active = nil
		ctx.g.promptPromotion()
		return
	}
	pl.Bench[ctx.Target.Pos-1] = nil
}

// SwitchActive switches your Active, when it passes Which, with the Benched
// Pokémon the card is dropped on (Lyra). Leaving the Active cures (C9).
type SwitchActive struct{ Which MonFilter }

func (v SwitchActive) Targets(ctx *EffectContext) []Spot {
	if !v.Which.ok(ctx.Player().Active) {
		return nil
	}
	return spotsOf(ctx.g, ctx.attacker, benchPositions(), nil)
}

func (SwitchActive) Apply(ctx *EffectContext) { switchIn(ctx.g, ctx.Target) }

// GustBench switches in the opponent's Benched Pokémon the card is dropped on
// to their Active Spot (Lana).
type GustBench struct{}

func (GustBench) Targets(ctx *EffectContext) []Spot {
	return spotsOf(ctx.g, 1-ctx.attacker, benchPositions(), nil)
}

func (GustBench) Apply(ctx *EffectContext) { switchIn(ctx.g, ctx.Target) }

// switchIn swaps the Benched Pokémon at s with its owner's Active.
func switchIn(g *Game, s Spot) {
	pl := &g.S.Players[s.Player]
	slot := s.Pos - 1
	pl.Active.CureAll()
	pl.Active, pl.Bench[slot] = pl.Bench[slot], pl.Active
	g.narrate("%s's %s was switched in to the Active Spot.", side(s.Player), pl.Active.Card.Name)
}

// MoveBenchEnergy moves all Energy of Type (NoEnergy: every type) from each
// of your Benched Pokémon to your Active (Lt. Surge, Juggler).
type MoveBenchEnergy struct{ Type Energy }

func (v MoveBenchEnergy) Apply(ctx *EffectContext) {
	pl := ctx.Player()
	for _, b := range pl.Bench {
		if b == nil {
			continue
		}
		for _, e := range energyOrder {
			if v.Type != NoEnergy && e != v.Type {
				continue
			}
			pl.Active.Energy[e] += b.Energy[e]
			delete(b.Energy, e)
		}
	}
	ctx.g.narrate("%s moved Energy to %s.", side(ctx.attacker), pl.Active.Card.Name)
}

// AttachFromZone attaches N Energy of Type, taken from the Energy Zone (not
// the turn's energy), to 1 of your Pokémon that passes Which (Kiawe).
type AttachFromZone struct {
	Which MonFilter
	Type  Energy
	N     int
}

func (v AttachFromZone) Targets(ctx *EffectContext) []Spot {
	return spotsOf(ctx.g, ctx.attacker, targets(), v.Which)
}

func (v AttachFromZone) Apply(ctx *EffectContext) { attachN(ctx, v.Type, v.N) }

func attachN(ctx *EffectContext, e Energy, n int) {
	m := ctx.target()
	m.Energy[e] += n
	ctx.g.narrate("%s attached %d %s Energy to %s.", side(ctx.attacker), n, e, m.Card.Name)
}

// AttachPerHeads flips a coin until tails and attaches an Energy of Type from
// the Energy Zone per heads to 1 of your Pokémon that passes Which (Misty).
type AttachPerHeads struct {
	Which MonFilter
	Type  Energy
}

func (v AttachPerHeads) Targets(ctx *EffectContext) []Spot {
	return spotsOf(ctx.g, ctx.attacker, targets(), v.Which)
}

func (v AttachPerHeads) Apply(ctx *EffectContext) {
	heads := 0
	for ctx.Flip() {
		heads++
	}
	attachN(ctx, v.Type, heads)
}

// AttachFromDiscard attaches N random Energy from your discard pile to 1 of
// your Pokémon that passes Which (Lusamine). Fewer are attached when fewer
// are there.
type AttachFromDiscard struct {
	Which MonFilter
	N     int
}

func (v AttachFromDiscard) Targets(ctx *EffectContext) []Spot {
	return spotsOf(ctx.g, ctx.attacker, targets(), v.Which)
}

func (v AttachFromDiscard) Apply(ctx *EffectContext) {
	pl, m := ctx.Player(), ctx.target()
	for i := 0; i < v.N && len(pl.DiscardedEnergy) > 0; i++ {
		k := ctx.g.rng.intn(len(pl.DiscardedEnergy))
		e := pl.DiscardedEnergy[k]
		pl.DiscardedEnergy = append(pl.DiscardedEnergy[:k:k], pl.DiscardedEnergy[k+1:]...)
		m.Energy[e]++
		ctx.g.narrate("%s attached %s Energy from the discard pile to %s.", side(ctx.attacker), e, m.Card.Name)
	}
}

// EndTurnNow ends the turn once the card's other ops are done (Kiawe).
type EndTurnNow struct{}

func (EndTurnNow) Apply(ctx *EffectContext) { ctx.g.concludeTurn() }

// ShuffleHandsDraw: each player shuffles their hand into their deck, then
// draws that many cards (Iono).
type ShuffleHandsDraw struct{}

func (ShuffleHandsDraw) Apply(ctx *EffectContext) {
	for _, p := range [2]int{ctx.attacker, 1 - ctx.attacker} {
		n := len(ctx.g.S.Players[p].Hand)
		ctx.g.redraw(p, n)
	}
}

// OppRedrawPointsLeft: your opponent shuffles their hand into their deck and
// draws a card for each point they still need to win (Mars).
type OppRedrawPointsLeft struct{}

func (OppRedrawPointsLeft) Apply(ctx *EffectContext) {
	opp := 1 - ctx.attacker
	ctx.g.redraw(opp, max(0, 3-ctx.g.S.Players[opp].Points))
}

// redraw shuffles player p's hand into their deck, then draws n.
func (g *Game) redraw(p, n int) {
	pl := &g.S.Players[p]
	pl.Deck = append(pl.Deck, pl.Hand...)
	pl.Hand = nil
	g.shuffleDeck(p)
	for i := 0; i < n && len(pl.Deck) > 0 && len(pl.Hand) < handMax; i++ {
		pl.Hand = append(pl.Hand, pl.Deck[0])
		pl.Deck = pl.Deck[1:]
	}
	g.narrate("%s shuffled their hand into their deck and drew %d.", side(p), len(pl.Hand))
}

func (g *Game) shuffleDeck(p int) {
	d := g.S.Players[p].Deck
	g.rng.shuffle(len(d), func(a, b int) { d[a], d[b] = d[b], d[a] })
}

// RevealOppDeck: your opponent reveals every card of Kind in their deck
// (Looker). Both players see the names in the log.
type RevealOppDeck struct{ Kind CardKind }

func (v RevealOppDeck) Apply(ctx *EffectContext) {
	var names []string
	for _, c := range ctx.Opponent().Deck {
		if c.Kind == v.Kind {
			names = append(names, c.Name)
		}
	}
	if len(names) == 0 {
		names = []string{"none"}
	}
	ctx.g.narrate("%s revealed from their deck: %s.", side(1-ctx.attacker), strings.Join(names, ", "))
}

// CopyOppSupporter looks at a random Supporter not named like this card from
// the opponent's deck, shuffles it back, and uses its effect as this card's
// (Penny). A Supporter whose text needs a target, isn't built, or can't be
// used right now does nothing.
type CopyOppSupporter struct{}

func (CopyOppSupporter) Apply(ctx *EffectContext) {
	opp := 1 - ctx.attacker
	var pool []Card
	for _, c := range ctx.Opponent().Deck {
		if c.Kind == KindSupporter && c.Name != ctx.Card.Name {
			pool = append(pool, c)
		}
	}
	if len(pool) == 0 {
		ctx.g.narrate("%s has no Supporter to copy.", side(opp))
		return
	}
	c := pool[ctx.g.rng.intn(len(pool))]
	ctx.g.shuffleDeck(opp)
	ctx.g.narrate("%s's %s copies %s.", side(ctx.attacker), ctx.Card.Name, c.Name)
	if !copyable(ctx, c) {
		ctx.g.narrate("%s has no effect here.", c.Name)
		return
	}
	for _, op := range c.Trainer {
		op.Apply(ctx)
	}
}

func copyable(ctx *EffectContext, c Card) bool {
	if c.Trainer == nil {
		return false
	}
	for _, op := range c.Trainer {
		if _, ok := op.(Targeted); ok {
			return false
		}
		if gate, ok := op.(Gated); ok && gate.CanPlay(ctx) != nil {
			return false
		}
	}
	return true
}
