package engine

import (
	"errors"
	"fmt"
)

// Game is a live match: the durable inputs (Seed, decks, Events) plus the
// derived State and a human-readable Narration feed. State is always
// reconstructible from Seed + decks + Events (ADR-0005); Narration is the
// no-theater "what happened" feed that replaces animation (ADR-0004).
type Game struct {
	Seed      uint64
	decks     [2][]Card
	rng       *rng
	S         *State
	Events    []Event
	Narration []string
	Actions   []Action // one per applied event, for replay and history (actions.go)

	flips []CoinFlip // coins flipped by the move being applied (coin.go)
}

var errBadPrompt = errors.New("invalid response to the current prompt")

func side(p int) string {
	if p == 0 {
		return "P1"
	}
	return "P2"
}

func (g *Game) narrate(format string, args ...any) {
	g.Narration = append(g.Narration, fmt.Sprintf(format, args...))
}

// NewGame deals a fresh match: shuffle both decks, draw opening hands
// (guaranteeing at least one Basic), flip for first player, and open the setup
// prompts. All randomness derives from seed. Each deck's Energy Zone uses its
// default energy types; see NewGameFromDecks to pass a chosen set.
func NewGame(seed uint64, deck0, deck1 []Card) *Game {
	return NewGameFromDecks(seed, [2]Deck{{Cards: deck0}, {Cards: deck1}})
}

// NewGameFromDecks is NewGame for decks that carry their chosen energy types
// (spec S2). The caller validates a choice with ValidateEnergyTypes.
func NewGameFromDecks(seed uint64, decks [2]Deck) *Game {
	g := &Game{Seed: seed, rng: newRng(seed)}
	g.decks[0] = decks[0].Cards
	g.decks[1] = decks[1].Cards
	st := &State{Winner: -1, Phase: PhaseSetup}
	g.S = st

	for i := 0; i < 2; i++ {
		pl := &st.Players[i]
		pl.Deck = append([]Card(nil), g.decks[i]...)
		pl.EnergyTypes = decks[i].energyTypes()
		pl.EnergyZone = NoEnergy
		pl.EnergyNext = NoEnergy
		g.dealOpeningHand(i)
	}
	g.flipForFirst()
	g.rollFirstEnergy()
	st.Pending = &Prompt{Player: 0, Kind: PromptSetup}
	g.narrate("Game start (seed %d). %s goes first.", seed, side(st.FirstPlayer))
	return g
}

// flipForFirst picks a random player to flip a coin (spec S4): the flipper
// goes first on heads, second on tails. Both draws come from the seeded rng.
func (g *Game) flipForFirst() {
	flip := CoinFlip{Flipper: g.rng.intn(2), Heads: g.rng.intn(2) == 0}
	g.S.OpeningFlip = flip
	g.S.FirstPlayer = flip.Flipper
	if !flip.Heads {
		g.S.FirstPlayer = 1 - flip.Flipper
	}
	result := "tails"
	if flip.Heads {
		result = "heads"
	}
	g.narrate("%s flips for first: %s.", side(flip.Flipper), result)
}

// rollFirstEnergy rolls each side's first energy as setup opens, so it shows
// from setup on (spec S4); the player going first rolls first.
func (g *Game) rollFirstEnergy() {
	first := g.S.FirstPlayer
	g.S.Players[first].EnergyNext = g.rollEnergy(first)
	g.S.Players[1-first].EnergyNext = g.rollEnergy(1 - first)
}

func (g *Game) dealOpeningHand(i int) {
	pl := &g.S.Players[i]
	for attempt := 0; attempt < 200; attempt++ {
		g.rng.shuffle(len(pl.Deck), func(a, b int) {
			pl.Deck[a], pl.Deck[b] = pl.Deck[b], pl.Deck[a]
		})
		hasBasic := false
		for _, c := range pl.Deck[:min(5, len(pl.Deck))] {
			if c.IsBasicPokemon() {
				hasBasic = true
				break
			}
		}
		if hasBasic {
			break
		}
	}
	n := min(5, len(pl.Deck))
	pl.Hand = append([]Card(nil), pl.Deck[:n]...)
	pl.Deck = pl.Deck[n:]
}

// Submit validates an event against the rules and applies it. On success it is
// appended to the canonical log; on failure nothing changes and an error is
// returned (block-the-illegal, ADR-0004/0007).
func (g *Game) Submit(ev Event) error {
	if g.S.Phase == PhaseOver {
		return errors.New("game is over")
	}
	act, before := g.describe(ev), len(g.Narration)
	g.flips = nil
	if err := g.validateAndApply(ev); err != nil {
		return err
	}
	act.Result = append([]string(nil), g.Narration[before:]...)
	act.Flips = g.flips
	g.Events = append(g.Events, ev)
	g.Actions = append(g.Actions, act)
	return nil
}

// validateAndApply refuses anything LegalMoves doesn't list, with the reason,
// then applies it. Apply functions assume the move is legal.
func (g *Game) validateAndApply(ev Event) error {
	if err := g.legality(ev); err != nil {
		return err
	}
	switch e := ev.(type) {
	case SetupActive:
		g.applySetupActive(e)
	case SetupBench:
		g.applySetupBench(e)
	case SetupReady:
		g.applySetupReady(e)
	case ChooseNewActive:
		g.applyChooseActive(e)
	case PlayBasic:
		g.applyPlayBasic(e)
	case AttachEnergy:
		g.applyAttach(e)
	case Evolve:
		g.applyEvolve(e)
	case Retreat:
		g.applyRetreat(e)
	case RetreatDiscard:
		g.S.Pending = nil
		g.finishRetreat(e)
	case PlayTrainer:
		g.applyPlayTrainer(e)
	case DiscardFossil:
		g.applyDiscardFossil(e)
	case UseAbility:
		g.applyUseAbility(e)
	case UseAttack:
		g.applyAttack(e)
	case EndTurn:
		g.concludeTurn()
	case Concede:
		g.applyConcede(e)
	}
	return nil
}

// --- Intent handlers ---

// Setup narration goes to both players, so it names no cards until both are
// Ready (spec S5).
func (g *Game) applySetupActive(e SetupActive) {
	pl := &g.S.Players[e.Player]
	idx := indexInHand(pl.Hand, e.CardID)
	pl.Active = newInPlay(pl.Hand[idx], 0)
	pl.Hand = removeAt(pl.Hand, idx)
	g.narrate("%s placed an Active.", side(e.Player))
}

func (g *Game) applySetupBench(e SetupBench) {
	pl := &g.S.Players[e.Player]
	idx := indexInHand(pl.Hand, e.CardID)
	pl.Bench[e.Slot] = newInPlay(pl.Hand[idx], 0)
	pl.Hand = removeAt(pl.Hand, idx)
	g.narrate("%s placed a Pokémon on the bench.", side(e.Player))
}

func (g *Game) applySetupReady(e SetupReady) {
	// Narration goes to both players, so it names no cards until both are Ready.
	g.narrate("%s is Ready.", side(e.Player))
	g.S.setup[e.Player] = true

	switch {
	case !g.S.setup[0]:
		g.S.Pending = &Prompt{Player: 0, Kind: PromptSetup}
	case !g.S.setup[1]:
		g.S.Pending = &Prompt{Player: 1, Kind: PromptSetup}
	default:
		g.narrateSetupReveal()
		g.S.Pending = nil
		g.S.Phase = PhaseMain
		g.S.Active = g.S.FirstPlayer
		g.S.Turn = 1
		g.startTurn(g.S.Active)
	}
}

// narrateSetupReveal names both players' setup picks once both are Ready.
func (g *Game) narrateSetupReveal() {
	for i := 0; i < 2; i++ {
		pl := &g.S.Players[i]
		g.narrate("%s placed %s as Active with %d on the bench.", side(i), pl.Active.Card.Name, pl.BenchCount())
	}
}

func (g *Game) applyPlayBasic(e PlayBasic) {
	pl := &g.S.Players[e.Player]
	idx := indexInHand(pl.Hand, e.CardID)
	c := pl.Hand[idx]
	pl.Hand = removeAt(pl.Hand, idx)
	pl.Bench[e.Slot] = newInPlay(c, g.S.Turn)
	g.narrate("%s played %s to the bench.", side(e.Player), c.Name)
}

func (g *Game) applyAttach(e AttachEnergy) {
	pl := &g.S.Players[e.Player]
	mon := targetMon(pl, e.Target)
	mon.Energy[pl.EnergyZone]++
	g.narrate("%s attached %s energy to %s.", side(e.Player), pl.EnergyZone, mon.Card.Name)
	pl.EnergyZone = NoEnergy
	pl.EnergyUsed = true
}

func (g *Game) applyEvolve(e Evolve) {
	pl := &g.S.Players[e.Player]
	hidx := indexInHand(pl.Hand, e.HandCardID)
	ev := pl.Hand[hidx]
	mon := targetMon(pl, e.Target)
	pl.Hand = removeAt(pl.Hand, hidx)
	mon.Card = ev
	mon.PlacedTurn = g.S.Turn
	mon.AbilityUsedTurn = 0 // an evolved Pokémon is a new card (A4)
	g.narrate("%s evolved into %s.", side(e.Player), ev.Name)
}

// applyRetreat asks which energy pays the retreat cost (T9); the swap happens
// once it is chosen, in finishRetreat.
func (g *Game) applyRetreat(e Retreat) {
	g.S.Players[e.Player].Retreated = true
	g.S.retreatTo = e.BenchIndex
	g.ask(e.Player, PromptRetreatEnergy)
}

func (g *Game) finishRetreat(e RetreatDiscard) {
	pl := &g.S.Players[e.Player]
	for _, t := range e.Energy {
		pl.Active.Energy[t]--
	}
	pl.DiscardedEnergy = append(pl.DiscardedEnergy, e.Energy...)
	slot := g.S.retreatTo
	pl.Active, pl.Bench[slot] = pl.Bench[slot], pl.Active
	g.narrate("%s retreated to %s.", side(e.Player), pl.Active.Card.Name)
}

func (g *Game) applyAttack(e UseAttack) {
	if g.confusionFails(e.Player) {
		g.concludeTurn()
		return
	}
	pl := &g.S.Players[e.Player]
	atk := pl.Active.Card.Attacks[e.Index]
	opp := &g.S.Players[1-e.Player]
	ctx := &EffectContext{g: g, attacker: e.Player, Damage: atk.Damage}
	for _, op := range atk.Effect {
		op.Apply(ctx)
	}
	dmg, weakness := g.attackDamage(e.Player, ctx.Damage)
	opp.Active.Damage += dmg
	g.surviveKO(e.Player)
	if weakness != "" {
		g.narrate("%s's %s used %s for %d (%s).", side(e.Player), pl.Active.Card.Name, atk.Name, dmg, weakness)
	} else {
		g.narrate("%s's %s used %s for %d.", side(e.Player), pl.Active.Card.Name, atk.Name, dmg)
	}

	g.resolveAction()
	// Attacking ends the turn — but a knockout may owe a prompt first.
	if g.S.Phase == PhaseOver {
		return
	}
	if g.S.Pending != nil {
		g.S.resume = resumeConclude
		return
	}
	g.concludeTurn()
}

// confusionFails flips for a Confused attacker (C8): tails, the attack fails
// with no self-damage and the condition stays.
func (g *Game) confusionFails(p int) bool {
	mon := g.S.Players[p].Active
	if !mon.Has(Confused) || g.flipCoin(p, true) {
		return false
	}
	g.narrate("%s's %s is Confused: tails, the attack failed.", side(p), mon.Card.Name)
	return true
}

func (g *Game) applyChooseActive(e ChooseNewActive) {
	g.S.Pending = nil
	g.promote(e)

	if g.S.Phase == PhaseOver || g.promptPromotion() {
		return
	}
	switch g.S.resume {
	case resumeConclude:
		g.S.resume = resumeNone
		g.concludeTurn()
	case resumeAdvance:
		g.S.resume = resumeNone
		g.advanceTurn()
	}
}

// --- Turn flow & resolution ---

// koPoints is what knocking out c scores (spec E1): 1, an ex 2, a Mega ex 3.
func koPoints(c Card) int {
	switch {
	case c.IsMegaEX:
		return 3
	case c.IsEX:
		return 2
	}
	return 1
}

// resolveAction finishes one action (spec E4) once all its damage has landed:
// remove every knocked-out Active, then award points, then check the result.
func (g *Game) resolveAction() {
	var pts [2]int
	for owner := 0; owner < 2; owner++ {
		pts[1-owner] += g.removeIfKnockedOut(owner)
	}
	for scorer := 0; scorer < 2; scorer++ {
		if pts[scorer] == 0 {
			continue
		}
		g.S.Players[scorer].Points += pts[scorer]
		g.narrate("%s scored %d (%d/3).", side(scorer), pts[scorer], g.S.Players[scorer].Points)
	}
	if g.checkResult() {
		return
	}
	g.promptPromotion()
}

// removeIfKnockedOut discards a player's Active if it has no HP left and
// returns what the KO scores for the opponent (0 if it is still standing).
func (g *Game) removeIfKnockedOut(owner int) int {
	pl := &g.S.Players[owner]
	mon := pl.Active
	if mon == nil || mon.remainingHP() > 0 {
		return 0
	}
	pl.Discard = append(pl.Discard, mon.Card)
	if mon.Tool != nil {
		pl.Discard = append(pl.Discard, *mon.Tool)
	}
	pl.Active = nil
	g.narrate("%s's %s was Knocked Out.", side(owner), mon.Card.Name)
	return koPoints(mon.Card)
}

// checkResult ends the game if an action decided it (spec E2, E3, E5): both at
// 3 or more is a tie whatever the Benches hold; otherwise first to 3 wins;
// otherwise a player with no Pokémon left loses (both without = tie).
func (g *Game) checkResult() bool {
	won0, won1 := g.S.Players[0].Points >= 3, g.S.Players[1].Points >= 3
	switch {
	case won0 && won1:
		g.tie(ReasonPoints)
	case won0:
		g.gameOver(0, ReasonPoints)
	case won1:
		g.gameOver(1, ReasonPoints)
	default:
		return g.checkNoPokemon()
	}
	return true
}

func (g *Game) checkNoPokemon() bool {
	has0, has1 := g.S.Players[0].hasPokemon(), g.S.Players[1].hasPokemon()
	switch {
	case has0 && has1:
		return false
	case has0:
		g.gameOver(0, ReasonNoPokemon)
	case has1:
		g.gameOver(1, ReasonNoPokemon)
	default:
		g.tie(ReasonNoPokemon)
	}
	return true
}

func (g *Game) promote(e ChooseNewActive) {
	pl := &g.S.Players[e.Player]
	mon := benchAt(pl, e.BenchIndex)
	pl.Bench[e.BenchIndex] = nil
	pl.Active = mon
	g.S.hidden = nil
	if g.S.Players[1-e.Player].Active == nil { // double KO: the other still picks (E6)
		g.S.hidden = &HiddenPromotion{e.Player, e.BenchIndex, len(g.Actions), len(g.Narration)}
	}
	g.narrate("%s promoted %s to Active.", side(e.Player), mon.Card.Name)
}

// promptPromotion asks each player left without an Active, in turn, to
// promote a new one (ADR-0007, mode 3). Reports whether a prompt is left
// pending. After a double KO the player whose turn is next picks first (spec
// E6); every KO resolves before the turn passes, so that is the non-active one.
func (g *Game) promptPromotion() bool {
	first := 1 - g.S.Active
	for _, p := range [2]int{first, 1 - first} {
		if g.S.Players[p].Active == nil && g.ask(p, PromptNewActive) {
			return true
		}
	}
	return false
}

// ask raises a prompt for player p (spec F8). Its options are p's legal
// answers to it. When there is only one, the engine makes it at once and the
// prompt never reaches the player; like a coin flip it is re-derived on
// replay, never logged. Reports whether the prompt is left pending.
func (g *Game) ask(p int, k PromptKind) bool {
	pr := &Prompt{Player: p, Kind: k}
	g.S.Pending = pr
	pr.Options = g.answers(p)
	if len(pr.Options) != 1 {
		return true
	}
	g.S.Pending = nil
	g.choose(pr.Options[0])
	return false
}

// answers is every legal move of p's except Concede, which is always open and
// answers nothing.
func (g *Game) answers(p int) []Event {
	var out []Event
	for _, ev := range g.LegalMoves(p) {
		if _, quit := ev.(Concede); !quit {
			out = append(out, ev)
		}
	}
	return out
}

// choose carries out the one possible answer to a prompt. A prompt kind raised
// through ask needs a case here.
func (g *Game) choose(ev Event) {
	switch e := ev.(type) {
	case ChooseNewActive:
		g.promote(e)
	case RetreatDiscard:
		g.finishRetreat(e)
	}
}

// concludeTurn ends the current turn: resolve end-of-turn status, then advance
// (unless a status knockout owes a prompt first).
func (g *Game) concludeTurn() {
	g.resolveStatus()
	if g.S.Phase == PhaseOver {
		return
	}
	if g.S.Pending != nil {
		g.S.resume = resumeAdvance
		return
	}
	g.advanceTurn()
}

// resolveStatus applies poison/burn to both Actives at end of turn (mode 2,
// automatic), then resolves any knockouts as one action.
func (g *Game) resolveStatus() {
	for i := 0; i < 2; i++ {
		pl := &g.S.Players[i]
		if pl.Active == nil {
			continue
		}
		if pl.Active.Poisoned {
			pl.Active.Damage += 10
			g.narrate("%s's %s took 10 from poison.", side(i), pl.Active.Card.Name)
		}
		if pl.Active.Burned {
			pl.Active.Damage += 20
			g.narrate("%s's %s took 20 from burn.", side(i), pl.Active.Card.Name)
			if g.flipCoin(i, true) {
				pl.Active.Burned = false
				g.narrate("%s's %s recovered from burn.", side(i), pl.Active.Card.Name)
			}
		}
		g.wearOffParalysis(i)
	}
	g.resolveAction() // one Checkup is one action (spec E4)
}

// wearOffParalysis cures player i's Paralyzed Active at the Checkup that ends
// i's turn, once a whole turn of i's has passed with it Paralyzed (C7).
func (g *Game) wearOffParalysis(i int) {
	mon := g.S.Players[i].Active
	if i != g.S.Active || !mon.Has(Paralyzed) || !mon.ownTurnParalyzed {
		return
	}
	mon.Cure(Paralyzed)
	g.narrate("%s's %s is no longer Paralyzed.", side(i), mon.Card.Name)
}

func (g *Game) advanceTurn() {
	if g.S.Turn >= g.S.Mode.turnCap() {
		g.tie(ReasonTurnCap)
		return
	}
	g.S.Active = 1 - g.S.Active
	g.S.Turn++
	g.startTurn(g.S.Active)
}

// handMax is the hand size at which draws are blocked (spec T3); the card
// stays in the deck.
const handMax = 10

// startTurn begins a player's turn: reset per-turn flags, draw, and advance
// the Energy Zone.
func (g *Game) startTurn(p int) {
	pl := &g.S.Players[p]
	pl.EnergyUsed = false
	pl.Retreated = false
	pl.SupporterUsed = false
	pl.StadiumUsed = false
	if pl.Active != nil && pl.Active.Has(Paralyzed) {
		pl.Active.ownTurnParalyzed = true
	}

	if len(pl.Deck) > 0 && len(pl.Hand) < handMax {
		c := pl.Deck[0]
		pl.Deck = pl.Deck[1:]
		pl.Hand = append(pl.Hand, c)
	}

	g.rollEnergyZone(p)
	g.narrate("--- Turn %d: %s ---", g.S.Turn, side(p))
}

// rollEnergyZone advances a player's Energy Zone at the start of their turn
// (spec T4–T6). On turn 1 the player going first gets none now; both next
// energies were rolled at setup (rollFirstEnergy). After that, energy left unattached last turn is
// discarded, next moves into now, and a new next is rolled.
func (g *Game) rollEnergyZone(p int) {
	if g.S.Turn == 1 {
		g.S.Players[p].EnergyZone = NoEnergy
		return
	}
	pl := &g.S.Players[p]
	if pl.EnergyZone != NoEnergy {
		pl.DiscardedEnergy = append(pl.DiscardedEnergy, pl.EnergyZone)
	}
	pl.EnergyZone = pl.EnergyNext
	pl.EnergyNext = g.rollEnergy(p)
}

func (g *Game) rollEnergy(p int) Energy {
	et := g.S.Players[p].EnergyTypes
	return et[g.rng.intn(len(et))]
}

// applyConcede ends the game in the opponent's favor when a player forfeits.
func (g *Game) applyConcede(e Concede) {
	g.narrate("%s conceded.", side(e.Player))
	g.gameOver(1-e.Player, ReasonConcede)
}

func (g *Game) gameOver(winner int, reason EndReason) {
	g.S.Phase = PhaseOver
	g.S.Winner = winner
	g.S.Reason = reason
	g.S.Pending = nil
	g.narrate("Game over — %s wins %d-%d.", side(winner), g.S.Players[winner].Points, g.S.Players[1-winner].Points)
}

// Replay rebuilds a game from its seed, decks and event log, then applies every
// event. Same inputs → identical game (ADR-0005). This is also how reconnection
// and replay viewing work.
func Replay(seed uint64, deck0, deck1 []Card, events []Event) (*Game, error) {
	return ReplayDecks(seed, [2]Deck{{Cards: deck0}, {Cards: deck1}}, events)
}

// ReplayDecks is Replay for decks that carry their chosen energy types.
func ReplayDecks(seed uint64, decks [2]Deck, events []Event) (*Game, error) {
	g := NewGameFromDecks(seed, decks)
	for i, ev := range events {
		if err := g.Submit(ev); err != nil {
			return nil, fmt.Errorf("event %d (%T): %w", i, ev, err)
		}
	}
	return g, nil
}

// --- small helpers ---

func newInPlay(c Card, placedTurn int) *InPlay {
	return &InPlay{Card: c, Energy: map[Energy]int{}, PlacedTurn: placedTurn}
}

func targetMon(pl *Player, target int) *InPlay {
	if target == 0 {
		return pl.Active
	}
	return benchAt(pl, target-1)
}

// benchAt returns the Pokémon in a bench slot, or nil if the slot is empty or
// does not exist.
func benchAt(pl *Player, slot int) *InPlay {
	if slot < 0 || slot >= BenchSlots {
		return nil
	}
	return pl.Bench[slot]
}

func indexInHand(hand []Card, id string) int {
	for i, c := range hand {
		if c.ID == id {
			return i
		}
	}
	return -1
}

func removeAt(hand []Card, i int) []Card {
	out := make([]Card, 0, len(hand)-1)
	out = append(out, hand[:i]...)
	return append(out, hand[i+1:]...)
}
