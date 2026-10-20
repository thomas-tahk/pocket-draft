package engine

import (
	"errors"
	"fmt"
)

// A card's choice (spec F9): a PromptChoose names its options by label, and
// the chooser answers with ChooseOption. What the answer does is a function
// kept in State; replay rebuilds it by folding the same events.

// ChooseOption answers a PromptChoose: Index picks Prompt.Labels[Index].
type ChooseOption struct {
	Player int
	Index  int
}

func (ChooseOption) event() {}

// ActChoose is the Action kind for a ChooseOption answer.
const ActChoose ActionKind = "choose"

// askChoice asks player p to pick one of labels; then runs with the pick. One
// option is picked at once and never shown (spec F8). Nothing to pick runs
// nothing.
func (g *Game) askChoice(p int, title string, labels []string, then func(g *Game, i int)) {
	switch len(labels) {
	case 0:
		return
	case 1:
		then(g, 0)
		return
	}
	g.showChoice(p, title, labels, then)
}

// showChoice raises the prompt even with one option: for a card that shows
// hidden cards to its player (Pokédex), the prompt is how they see them.
func (g *Game) showChoice(p int, title string, labels []string, then func(g *Game, i int)) {
	pr := &Prompt{Player: p, Kind: PromptChoose, Title: title, Labels: labels}
	for i := range labels {
		pr.Options = append(pr.Options, ChooseOption{Player: p, Index: i})
	}
	g.S.Pending, g.S.then = pr, then
}

func listChooseOption(g *Game, p int) []ChooseOption {
	pr := g.S.Pending
	if pr == nil || pr.Kind != PromptChoose || pr.Player != p {
		return nil
	}
	out := make([]ChooseOption, len(pr.Labels))
	for i := range out {
		out[i] = ChooseOption{Player: p, Index: i}
	}
	return out
}

func (g *Game) checkChooseOption(e ChooseOption) error {
	if err := g.onPrompt(e.Player, PromptChoose); err != nil {
		return err
	}
	if e.Index < 0 || e.Index >= len(g.S.Pending.Labels) {
		return errors.New("no such option")
	}
	return nil
}

func (g *Game) applyChooseOption(e ChooseOption) {
	then := g.S.then
	g.S.Pending, g.S.then = nil, nil
	then(g, e.Index)
}

// ChooseOne: you choose one of Labels, and the ops at the same index run
// (Eevee Bag).
type ChooseOne struct {
	Labels []string
	Ops    [][]EffectOp
}

func (v ChooseOne) Apply(ctx *EffectContext) {
	p, card := ctx.attacker, ctx.Card
	ctx.g.askChoice(p, "Choose 1", v.Labels, func(g *Game, i int) {
		g.narrate("%s chose: %s", side(p), v.Labels[i])
		inner := g.trainerCtx(p, card, NoSpot)
		for _, op := range v.Ops[i] {
			op.Apply(inner)
		}
	})
}

// handLabels names each card in a hand, for a choice among them.
func handLabels(hand []Card) []string {
	out := make([]string, len(hand))
	for i, c := range hand {
		out[i] = c.Name
	}
	return out
}

// HandToDeckThenDraw: you choose N cards from your hand, shuffle them into
// your deck, then draw Draw cards. Needs N cards to choose (Maintenance).
type HandToDeckThenDraw struct{ N, Draw int }

func (v HandToDeckThenDraw) CanPlay(ctx *EffectContext) error {
	if len(ctx.Player().Hand)-1 < v.N { // the played card is still in hand
		return fmt.Errorf("you need %d other cards in hand", v.N)
	}
	return nil
}

func (v HandToDeckThenDraw) Apply(ctx *EffectContext) {
	ctx.g.handToDeck(ctx.attacker, v.N, v.Draw)
}

func (g *Game) handToDeck(p, left, draw int) {
	pl := &g.S.Players[p]
	if left == 0 {
		g.shuffleDeck(p)
		g.drawCards(p, draw)
		g.narrate("%s shuffled cards into their deck and drew %d.", side(p), draw)
		return
	}
	g.askChoice(p, "Choose a card to shuffle into your deck", handLabels(pl.Hand), func(g *Game, i int) {
		pl := &g.S.Players[p]
		pl.Deck = append(pl.Deck, pl.Hand[i])
		pl.Hand = removeAt(pl.Hand, i)
		g.handToDeck(p, left-1, draw)
	})
}

func (g *Game) drawCards(p, n int) {
	pl := &g.S.Players[p]
	for i := 0; i < n && len(pl.Deck) > 0 && len(pl.Hand) < handMax; i++ {
		pl.Hand = append(pl.Hand, pl.Deck[0])
		pl.Deck = pl.Deck[1:]
	}
}

// OppDiscardDownTo: your opponent discards cards of their choice from their
// hand until they have N (Nasty Notice).
type OppDiscardDownTo struct{ N int }

func (v OppDiscardDownTo) CanPlay(ctx *EffectContext) error {
	if len(ctx.opp().Hand) <= v.N {
		return fmt.Errorf("your opponent has %d or fewer cards in hand", v.N)
	}
	return nil
}

func (v OppDiscardDownTo) Apply(ctx *EffectContext) { ctx.g.discardDownTo(1-ctx.attacker, v.N) }

func (g *Game) discardDownTo(p, n int) {
	pl := &g.S.Players[p]
	if len(pl.Hand) <= n {
		return
	}
	g.askChoice(p, fmt.Sprintf("Discard down to %d cards", n), handLabels(pl.Hand), func(g *Game, i int) {
		pl := &g.S.Players[p]
		c := pl.Hand[i]
		pl.Hand = removeAt(pl.Hand, i)
		pl.Discard = append(pl.Discard, c)
		g.narrate("%s discarded %s.", side(p), c.Name)
		g.discardDownTo(p, n)
	})
}

// LookAtTop shows you the top N cards of your deck; with MayShuffle you then
// choose whether to shuffle it (Pokédex, Rotom Dex).
type LookAtTop struct {
	N          int
	MayShuffle bool
}

func (LookAtTop) CanPlay(ctx *EffectContext) error {
	if len(ctx.Player().Deck) == 0 {
		return errDeckEmpty
	}
	return nil
}

func (v LookAtTop) Apply(ctx *EffectContext) {
	p := ctx.attacker
	top := ctx.Player().Deck[:min(v.N, len(ctx.Player().Deck))]
	title := "Top of your deck: " + joinNames(top)
	ctx.g.narrate("%s looked at the top %d cards of their deck.", side(p), len(top))
	if !v.MayShuffle {
		ctx.g.showChoice(p, title, []string{"OK"}, func(*Game, int) {})
		return
	}
	ctx.g.showChoice(p, title, []string{"Shuffle your deck", "Don't shuffle"}, func(g *Game, i int) {
		if i == 0 {
			g.shuffleDeck(p)
			g.narrate("%s shuffled their deck.", side(p))
		}
	})
}

func joinNames(cards []Card) string {
	s := ""
	for i, c := range cards {
		if i > 0 {
			s += ", "
		}
		s += c.Name
	}
	return s
}

// BenchOppBasicFromDiscard puts a Basic Pokémon of your choice from your
// opponent's discard pile onto their Bench, in its leftmost empty slot
// (Pokémon Flute).
type BenchOppBasicFromDiscard struct{}

func (BenchOppBasicFromDiscard) CanPlay(ctx *EffectContext) error {
	opp := ctx.opp()
	if opp.BenchCount() == BenchSlots {
		return errors.New("your opponent's Bench is full")
	}
	if len(basicsIn(opp.Discard)) == 0 {
		return errors.New("there is no Basic Pokémon in your opponent's discard pile")
	}
	return nil
}

func basicsIn(cards []Card) []int {
	var out []int
	for i, c := range cards {
		if c.IsBasicPokemon() {
			out = append(out, i)
		}
	}
	return out
}

func (BenchOppBasicFromDiscard) Apply(ctx *EffectContext) {
	o := 1 - ctx.attacker
	idx := basicsIn(ctx.opp().Discard)
	labels := make([]string, len(idx))
	for i, j := range idx {
		labels[i] = ctx.opp().Discard[j].Name
	}
	ctx.g.askChoice(ctx.attacker, "Choose a Basic Pokémon for your opponent's Bench", labels, func(g *Game, i int) {
		pl := &g.S.Players[o]
		c := pl.Discard[idx[i]]
		pl.Discard = removeAt(pl.Discard, idx[i])
		for slot := range pl.Bench {
			if pl.Bench[slot] == nil {
				pl.Bench[slot] = newInPlay(c, g.S.Turn)
				break
			}
		}
		g.narrate("%s put %s onto %s's Bench.", side(1-o), c.Name, side(o))
	})
}

// SwitchOppActiveBasic switches your opponent's Active Basic Pokémon to the
// Bench; they choose the new Active (Repel). Leaving the Active spot clears
// its conditions (spec C9).
type SwitchOppActiveBasic struct{}

func (SwitchOppActiveBasic) CanPlay(ctx *EffectContext) error {
	if ctx.defenderMon().Card.Stage != Basic {
		return errors.New("your opponent's Active Pokémon isn't a Basic Pokémon")
	}
	if ctx.opp().BenchCount() == 0 {
		return errors.New("your opponent has no Benched Pokémon")
	}
	return nil
}

func (SwitchOppActiveBasic) Apply(ctx *EffectContext) {
	o := 1 - ctx.attacker
	opp := ctx.opp()
	opp.Active.CureAll()
	ctx.g.S.switchedOut, opp.Active = opp.Active, nil
	ctx.g.narrate("%s switched out %s's Active.", ctx.Card.Name, side(o))
	ctx.g.ask(o, PromptNewActive)
}

// RareCandy evolves the targeted Basic of yours straight into a Stage 2 from
// your hand that evolves from it, skipping the Stage 1; you choose which when
// several do. Not on your first turn, nor on a Basic put into play this turn.
// BasicOf names the Basic a Stage 1 evolves from (the card data knows it).
type RareCandy struct{ BasicOf func(stage1 string) string }

var errFirstTurnCandy = errors.New("you can't use this card during your first turn")

func (RareCandy) CanPlay(ctx *EffectContext) error {
	if ctx.g.S.Turn <= 2 {
		return errFirstTurnCandy
	}
	return nil
}

// stage2s is the index of each Stage 2 in your hand that evolves from mon.
func (v RareCandy) stage2s(ctx *EffectContext, mon *InPlay) []int {
	var out []int
	for i, c := range ctx.Player().Hand {
		if c.Kind == KindPokemon && c.Stage == Stage2 && v.BasicOf(c.EvolvesFrom) == mon.Card.Name {
			out = append(out, i)
		}
	}
	return out
}

// Targets is each Basic of yours not put into play this turn with a Stage 2
// for it in hand.
func (v RareCandy) Targets(ctx *EffectContext) []Spot {
	return ctx.g.ownSpots(ctx.attacker, func(_ Spot, m *InPlay) bool {
		return m.Card.Stage == Basic && m.PlacedTurn != ctx.g.S.Turn && len(v.stage2s(ctx, m)) > 0
	})
}

func (v RareCandy) Apply(ctx *EffectContext) {
	mon, p := ctx.targetMon(), ctx.attacker
	hand := ctx.Player().Hand
	idx := v.stage2s(ctx, mon)
	labels := make([]string, len(idx))
	for i, j := range idx {
		labels[i] = hand[j].Name
	}
	ctx.g.askChoice(p, "Choose the Stage 2 to evolve into", labels, func(g *Game, i int) {
		pl := &g.S.Players[p]
		ev := pl.Hand[idx[i]]
		pl.Hand = removeAt(pl.Hand, idx[i])
		g.evolveMon(p, mon, ev)
	})
}
