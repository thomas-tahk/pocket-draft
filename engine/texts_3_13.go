package engine

import (
	"errors"
	"maps"
)

// Tool and Stadium texts, batch 3.13 (issue #73). Three capabilities the
// always-on hooks (passive.go) don't cover, and the ops the texts are built
// from; the wordings are mapped in server/texts_3_13.go.
//
//   - HP: HookMaxHP adds to a Pokémon's HP while its source is in play.
//   - Triggers: a Tool's or Stadium's text that fires on an event (end of a
//     turn, a Tool attached, an attack's damage, a knockout): Card.Triggers.
//   - Stadium use: "Once during each player's turn, that player may …" is a
//     UseStadium move that runs Card.StadiumUse, once per player per turn.

// maxHP is the Pokémon at h's HP with HookMaxHP.
func (g *Game) maxHP(h Holder) int { return h.Card.HP + g.sumHook(HookMaxHP, h, nil) }

func (g *Game) hpLeft(p, pos int) int {
	h := g.holder(p, pos)
	return g.maxHP(h) - h.Damage
}

// MaxHP is player p's Pokémon at pos's HP now, Tools and Stadium included.
func (g *Game) MaxHP(p, pos int) int { return g.maxHP(g.holder(p, pos)) }

// TriggerOn is the event a Trigger fires on.
type TriggerOn uint8

const (
	OnTurnEnd    TriggerOn = iota // the end of each turn, before the Checkup
	OnAttached                    // this Tool was attached
	OnAttackHit                   // its Pokémon, Active, took damage from the opponent's attack
	OnKnockedOut                  // its Pokémon was Knocked Out by damage from the opponent's attack
	OnKnocksOut                   // its Pokémon's attack Knocked Out the opponent's Pokémon
)

// Trigger is one piece of a Tool's or Stadium's text that fires on an event.
type Trigger struct {
	On   TriggerOn
	When func(TriggerCheck) bool // nil means always
	Do   TriggerOp
}

// TriggerCheck is what a Trigger's When is asked about: where its card sits
// (a Tool with its Pokémon; a Stadium at NoSpot) and whose turn it is.
type TriggerCheck struct {
	Source Holder
	Turn   int
}

// TriggerOp is what a Trigger does.
type TriggerOp interface {
	Fire(t *TriggerContext)
}

// TriggerContext is a firing Trigger's view of the game.
type TriggerContext struct {
	g      *Game
	Source Holder
	Card   Card
	Turn   int
	toHand bool // Rescue Scarf: the Knocked Out Pokémon goes to the hand
}

type triggerSource struct {
	at   Holder
	card Card
}

// triggerSources is every Tool in play, both players, Active first, then the
// Stadium.
func (g *Game) triggerSources() []triggerSource {
	var out []triggerSource
	for p := 0; p < 2; p++ {
		for _, pos := range targets() {
			if at := g.holder(p, pos); at.InPlay != nil && at.Tool != nil {
				out = append(out, triggerSource{at, *at.Tool})
			}
		}
	}
	if st := g.S.Stadium; st != nil {
		out = append(out, triggerSource{Holder{NoSpot, nil}, *st})
	}
	return out
}

// fire runs src's triggers on on; reports whether one sent the Pokémon to hand.
func (g *Game) fire(on TriggerOn, src triggerSource) bool {
	ctx := &TriggerContext{g: g, Source: src.at, Card: src.card, Turn: g.S.Active}
	for _, t := range src.card.Triggers {
		if t.On == on && (t.When == nil || t.When(TriggerCheck{src.at, ctx.Turn})) {
			t.Do.Fire(ctx)
		}
	}
	return ctx.toHand
}

// fireTool runs the triggers of the Tool on player p's Pokémon at pos.
func (g *Game) fireTool(on TriggerOn, p, pos int) bool {
	at := g.holder(p, pos)
	if at.InPlay == nil || at.Tool == nil {
		return false
	}
	return g.fire(on, triggerSource{at, *at.Tool})
}

// fireTurnEnd runs every OnTurnEnd trigger in play.
func (g *Game) fireTurnEnd() {
	for _, src := range g.triggerSources() {
		g.fire(OnTurnEnd, src)
	}
}

// resolveAttack finishes an attack that did dmg to the defending Active: the
// defender's Tool reacts to the hit, then knockouts resolve, with knockouts
// of the defender's Pokémon counted as by the attack.
func (g *Game) resolveAttack(attacker, dmg int) {
	if dmg > 0 {
		g.fireTool(OnAttackHit, 1-attacker, 0)
	}
	g.resolveAction()
}

// knockedOut runs the knockout triggers for owner's Active mon, still in
// place, and reports whether it goes to the hand instead of the discard.
func (g *Game) knockedOut(owner int) bool {
	if g.koAttacker == nil || owner == g.S.Active {
		return false
	}
	g.fireTool(OnKnocksOut, g.S.Active, 0)
	return g.fireTool(OnKnockedOut, owner, 0)
}

// --- Trigger ops ---

func (t *TriggerContext) owner() *Player { return &t.g.S.Players[t.Source.Player] }
func (t *TriggerContext) foe() *Player   { return &t.g.S.Players[1-t.Source.Player] }

func (t *TriggerContext) heal(mon *InPlay, n int) {
	n = t.g.heal(mon, n)
	if n > 0 {
		t.g.narrate("%s healed %d damage (%s).", mon.Card.Name, n, t.Card.Name)
	}
}

// discardThis discards the Tool firing from its Pokémon.
func (t *TriggerContext) discardThis() {
	mon := t.Source.InPlay
	t.owner().Discard = append(t.owner().Discard, *mon.Tool)
	t.g.narrate("%s was discarded from %s.", mon.Tool.Name, mon.Card.Name)
	mon.Tool = nil
}

// drawTo draws for player p until the hand holds n (at most handMax).
func (t *TriggerContext) drawTo(p, n int) {
	pl := &t.g.S.Players[p]
	drawn := 0
	for len(pl.Hand) < min(n, handMax) && len(pl.Deck) > 0 {
		pl.Hand, pl.Deck = append(pl.Hand, pl.Deck[0]), pl.Deck[1:]
		drawn++
	}
	if drawn > 0 {
		t.g.narrate("%s drew %d (%s).", side(p), drawn, t.Card.Name)
	}
}

// DamageFoeActive does N damage to the opponent's Active (Deceptive Needle).
type DamageFoeActive struct{ N int }

func (v DamageFoeActive) Fire(t *TriggerContext) {
	if mon := t.foe().Active; mon != nil {
		mon.Damage += v.N
		t.g.narrate("%s did %d damage to %s.", t.Card.Name, v.N, mon.Card.Name)
	}
}

// DamageAttacker does N damage to the attacking Pokémon (Rocky Helmet).
type DamageAttacker struct{ N int }

func (v DamageAttacker) Fire(t *TriggerContext) { DamageFoeActive(v).Fire(t) }

// PoisonAttacker Poisons the attacking Pokémon (Poison Barb).
type PoisonAttacker struct{}

func (PoisonAttacker) Fire(t *TriggerContext) {
	p := 1 - t.Source.Player
	if mon := t.foe().Active; mon != nil && t.g.canBeAffected(p, 0) {
		mon.Inflict(Poison)
		t.g.narrate("%s Poisoned %s.", t.Card.Name, mon.Card.Name)
	}
}

// FoeHandCardToDeck shuffles a random card from the opponent's hand into
// their deck (Dark Pendant).
type FoeHandCardToDeck struct{}

func (FoeHandCardToDeck) Fire(t *TriggerContext) {
	pl := t.foe()
	if len(pl.Hand) == 0 {
		return
	}
	i := t.g.rng.intn(len(pl.Hand))
	c := pl.Hand[i]
	pl.Hand = removeAt(pl.Hand, i)
	pl.Deck = append(pl.Deck, c)
	t.g.rng.shuffle(len(pl.Deck), func(a, b int) { pl.Deck[a], pl.Deck[b] = pl.Deck[b], pl.Deck[a] })
	t.g.narrate("%s revealed %s and shuffled it into their deck (%s).", side(1-t.Source.Player), c.Name, t.Card.Name)
}

// HealToolPokemon heals N from the Tool's Pokémon (Leftovers).
type HealToolPokemon struct{ N int }

func (v HealToolPokemon) Fire(t *TriggerContext) {
	if t.Source.Damage > 0 {
		t.heal(t.Source.InPlay, v.N)
	}
}

// HealAtHalf heals N from the Tool's Pokémon if it has half its HP or less
// left, then discards the Tool (Sitrus Berry).
type HealAtHalf struct{ N int }

func (v HealAtHalf) Fire(t *TriggerContext) {
	left := t.g.hpLeft(t.Source.Player, t.Source.Pos)
	if left*2 > t.g.maxHP(t.Source) {
		return
	}
	t.heal(t.Source.InPlay, v.N)
	t.discardThis()
}

// CureThenDiscard cures every Special Condition of the Tool's Pokémon, if it
// has any, and discards the Tool (Lum Berry).
type CureThenDiscard struct{}

func (CureThenDiscard) Fire(t *TriggerContext) {
	if len(t.Source.Conditions()) == 0 {
		return
	}
	t.Source.CureAll()
	t.g.narrate("%s recovered from all Special Conditions (%s).", t.Source.Card.Name, t.Card.Name)
	t.discardThis()
}

// CureSelf cures every Special Condition of the Tool's Pokémon (Steel Apron).
type CureSelf struct{}

func (CureSelf) Fire(t *TriggerContext) { t.Source.CureAll() }

// DiscardThis discards the Tool (Metal Core Barrier).
type DiscardThis struct{}

func (DiscardThis) Fire(t *TriggerContext) { t.discardThis() }

// TurnPlayerDrawsTo: the player whose turn it is draws until N in hand
// (Hiking Trail).
type TurnPlayerDrawsTo struct{ N int }

func (v TurnPlayerDrawsTo) Fire(t *TriggerContext) { t.drawTo(t.Turn, v.N) }

// OwnerDrawsTo: the Tool's owner draws until N in hand (Lucky Egg).
type OwnerDrawsTo struct{ N int }

func (v OwnerDrawsTo) Fire(t *TriggerContext) { t.drawTo(t.Source.Player, v.N) }

// OwnerDraws: the Tool's owner draws N (Lucky Mittens).
type OwnerDraws struct{ N int }

func (v OwnerDraws) Fire(t *TriggerContext) {
	t.drawTo(t.Source.Player, len(t.owner().Hand)+v.N)
}

// HealEachWithEnergy: the player whose turn it is heals N from each of their
// Pokémon with Energy of Type attached (Soothing Shore).
type HealEachWithEnergy struct {
	N    int
	Type Energy
}

func (v HealEachWithEnergy) Fire(t *TriggerContext) {
	for _, pos := range targets() {
		if mon := targetMon(&t.g.S.Players[t.Turn], pos); mon != nil && mon.Energy[v.Type] > 0 && mon.Damage > 0 {
			t.heal(mon, v.N)
		}
	}
}

// ToHandInstead puts the Knocked Out Pokémon into its owner's hand instead of
// the discard pile (Rescue Scarf). At a full hand it is discarded.
type ToHandInstead struct{}

func (ToHandInstead) Fire(t *TriggerContext) { t.toHand = len(t.owner().Hand) < handMax }

// EnergyToBench moves up to N Energy of Type from the Knocked Out Pokémon, 1
// each to N of its owner's Benched Pokémon, left slot first (Electrical Cord).
type EnergyToBench struct {
	N    int
	Type Energy
}

func (v EnergyToBench) Fire(t *TriggerContext) {
	from, moved := t.Source.InPlay, 0
	for slot := 0; slot < BenchSlots && moved < v.N && from.Energy[v.Type] > 0; slot++ {
		if to := t.owner().Bench[slot]; to != nil {
			from.Energy[v.Type]--
			to.Energy[v.Type]++
			moved++
			t.g.narrate("%s moved a %s Energy to %s.", t.Card.Name, v.Type, to.Card.Name)
		}
	}
}

// --- Stadium use ---

// UseStadium uses the Stadium in play's text (Card.StadiumUse), once per
// player per turn. CardID is the card picked from hand for a text that picks
// one (HandPick), else "".
type UseStadium struct {
	Player int
	CardID string
}

func (UseStadium) event() {}

// ActUseStadium is the Action kind for a used Stadium.
const ActUseStadium ActionKind = "use_stadium"

// HandPick is an optional op method for a Stadium's use: which hand cards it
// may pick. The picked card is ctx.Picked.
type HandPick interface {
	Picks(ctx *EffectContext, c Card) bool
}

func (g *Game) stadiumCtx(e UseStadium) *EffectContext {
	ctx := g.trainerCtx(e.Player, *g.S.Stadium, NoSpot)
	ctx.Picked = e.CardID
	return ctx
}

func listUseStadium(g *Game, p int) []UseStadium {
	st := g.S.Stadium
	if st == nil || st.StadiumUse == nil {
		return nil
	}
	picker := handPicker(st.StadiumUse)
	if picker == nil {
		return []UseStadium{{Player: p}}
	}
	var out []UseStadium
	for _, c := range g.S.Players[p].Hand {
		out = append(out, UseStadium{Player: p, CardID: c.ID})
	}
	return out
}

func handPicker(ops []EffectOp) HandPick {
	for _, op := range ops {
		if hp, ok := op.(HandPick); ok {
			return hp
		}
	}
	return nil
}

func (g *Game) checkUseStadium(e UseStadium) error {
	if err := g.onTurn(e.Player); err != nil {
		return err
	}
	st := g.S.Stadium
	switch {
	case st == nil || st.StadiumUse == nil:
		return errNoStadiumUse
	case g.S.Players[e.Player].StadiumUseTurn == g.S.Turn:
		return errStadiumUsed
	}
	ctx := g.stadiumCtx(e)
	if err := g.checkPick(ctx, st.StadiumUse); err != nil {
		return err
	}
	for _, op := range st.StadiumUse {
		if gate, ok := op.(Gated); ok {
			if err := gate.CanPlay(ctx); err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *Game) checkPick(ctx *EffectContext, ops []EffectOp) error {
	picker := handPicker(ops)
	if picker == nil {
		return nil
	}
	hand := ctx.Player().Hand
	if i := indexInHand(hand, ctx.Picked); i < 0 || !picker.Picks(ctx, hand[i]) {
		return errCantPick
	}
	return nil
}

func (g *Game) applyUseStadium(e UseStadium) {
	g.S.Players[e.Player].StadiumUseTurn = g.S.Turn
	g.narrate("%s used %s.", side(e.Player), g.S.Stadium.Name)
	ctx := g.stadiumCtx(e)
	for _, op := range g.S.Stadium.StadiumUse {
		op.Apply(ctx)
	}
}

var (
	errNoStadiumUse = errors.New("no Stadium in play can be used")
	errStadiumUsed  = errors.New("you already used the Stadium this turn")
	errCantPick     = errors.New("that card can't be picked")
)

// --- Stadium-use ops ---

// FlipAllDrawTo flips Coins coins; if all are heads, draws until N in hand
// (Arcade).
type FlipAllDrawTo struct{ Coins, N int }

func (v FlipAllDrawTo) CanPlay(ctx *EffectContext) error {
	if len(ctx.Player().Hand) >= min(v.N, handMax) {
		return errHandFull
	}
	return nil
}

func (v FlipAllDrawTo) Apply(ctx *EffectContext) {
	heads := 0
	for i := 0; i < v.Coins; i++ {
		if ctx.Flip() {
			heads++
		}
	}
	ctx.g.narrate("%s flipped %d of %d heads.", side(ctx.attacker), heads, v.Coins)
	if heads == v.Coins {
		DrawCards{N: v.N - len(ctx.Player().Hand)}.Apply(ctx)
	}
}

// BasicToDeckDraw shuffles the picked Basic Pokémon from hand into the deck,
// then draws a card (Area Zero).
type BasicToDeckDraw struct{}

func (BasicToDeckDraw) Picks(_ *EffectContext, c Card) bool { return c.IsBasicPokemon() }

func (BasicToDeckDraw) Apply(ctx *EffectContext) {
	ctx.handToDeck(ctx.Picked)
	DrawCards{N: 1}.Apply(ctx)
}

// SwapForTool swaps the picked hand card with a random Pokémon Tool from the
// deck (Kid's Room). It can't be used with no Tool in the deck.
type SwapForTool struct{}

func (SwapForTool) Picks(_ *EffectContext, _ Card) bool { return true }

func (SwapForTool) CanPlay(ctx *EffectContext) error {
	for _, c := range ctx.Player().Deck {
		if c.Kind == KindTool {
			return nil
		}
	}
	return errors.New("there is no Pokémon Tool in your deck")
}

func (SwapForTool) Apply(ctx *EffectContext) {
	pl := ctx.Player()
	var tools []int
	for i, c := range pl.Deck {
		if c.Kind == KindTool {
			tools = append(tools, i)
		}
	}
	i := tools[ctx.g.rng.intn(len(tools))]
	tool := pl.Deck[i]
	pl.Deck = removeAt(pl.Deck, i)
	ctx.handToDeck(ctx.Picked)
	pl.Hand = append(pl.Hand, tool)
	ctx.g.narrate("%s took %s from the deck.", side(ctx.attacker), tool.Name)
}

// handToDeck shuffles the hand card id into the acting player's deck.
func (ctx *EffectContext) handToDeck(id string) {
	pl := ctx.Player()
	i := indexInHand(pl.Hand, id)
	pl.Deck = append(pl.Deck, pl.Hand[i])
	pl.Hand = removeAt(pl.Hand, i)
	ctx.g.rng.shuffle(len(pl.Deck), func(a, b int) { pl.Deck[a], pl.Deck[b] = pl.Deck[b], pl.Deck[a] })
	ctx.g.narrate("%s shuffled a card from their hand into the deck.", side(ctx.attacker))
}

// RerollEnergy discards this turn's Energy in the Energy Zone; the next one
// moves up and a new next one is rolled (Rainbow Cave).
type RerollEnergy struct{}

func (RerollEnergy) CanPlay(ctx *EffectContext) error {
	if ctx.Player().EnergyZone == NoEnergy {
		return errors.New("there is no Energy in your Energy Zone")
	}
	return nil
}

func (RerollEnergy) Apply(ctx *EffectContext) {
	pl := ctx.Player()
	pl.DiscardedEnergy = append(pl.DiscardedEnergy, pl.EnergyZone)
	pl.EnergyZone, pl.EnergyNext = pl.EnergyNext, ctx.g.rollEnergy(ctx.attacker)
	ctx.g.narrate("%s discarded their Energy; %s Energy was produced.", side(ctx.attacker), pl.EnergyZone)
}

// stadiumCard is the Stadium in play, or the zero Card.
func (g *Game) stadiumCard() Card {
	if g.S.Stadium == nil {
		return Card{}
	}
	return *g.S.Stadium
}

// Attacks is every attack the Pokémon can use: its card's, then those its
// Tool lends (Memory Light).
func (m *InPlay) Attacks() []Attack {
	if m.Tool == nil || m.Tool.LendAttacks == nil {
		return m.Card.Attacks
	}
	return append(m.Card.Attacks[:len(m.Card.Attacks):len(m.Card.Attacks)], m.Tool.LendAttacks(m.Card)...)
}

// runAttackEffects runs an attack's effect ops. If the defending Active has
// HookEffectShield (Clear Veil), what they do to it is undone: the attack's
// damage is worked out from ctx afterwards and still lands.
func (g *Game) runAttackEffects(attacker int, ops []EffectOp, ctx *EffectContext) {
	def := g.S.Players[1-attacker].Active
	if def == nil || !g.hasHook(HookEffectShield, g.holder(1-attacker, 0), g.S.Players[attacker].Active) {
		for _, op := range ops {
			op.Apply(ctx)
		}
		return
	}
	saved := *def
	saved.Energy = maps.Clone(def.Energy)
	for _, op := range ops {
		op.Apply(ctx)
	}
	*def = saved
}
