package engine

import (
	"errors"
	"fmt"
)

// LegalMoves is the single source of truth for what a player may do right now.
// Submit refuses any event that is not in this list, the screen lights spots
// from it and the bot picks from it.
//
// Each kind of move is one entry in moveRules: a generator that lists every
// candidate of that kind (legal or not) and a check that says why a candidate
// is illegal. A candidate is legal when its check passes. Adding a move kind
// means adding its Event type, one rule here, and its apply case in engine.go.

// moveRule is one kind of move, type-erased so the rules can sit in one list.
type moveRule struct {
	is    func(Event) bool
	list  func(g *Game, p int) []Event
	check func(g *Game, ev Event) error
	key   func(ev Event) string // identity for matching a submitted event to a listed one
}

// rule builds a moveRule for event type E. key may be nil when the event's
// printed value is already its identity.
func rule[E Event](list func(g *Game, p int) []E, check func(g *Game, e E) error, key func(e E) string) moveRule {
	if key == nil {
		key = func(e E) string { return fmt.Sprintf("%#v", e) }
	}
	return moveRule{
		is: func(ev Event) bool { _, ok := ev.(E); return ok },
		list: func(g *Game, p int) []Event {
			var out []Event
			for _, e := range list(g, p) {
				out = append(out, e)
			}
			return out
		},
		check: func(g *Game, ev Event) error { return check(g, ev.(E)) },
		key:   func(ev Event) string { return key(ev.(E)) },
	}
}

// moveRules is in the order LegalMoves lists moves.
var moveRules = []moveRule{
	rule(listSetupActive, (*Game).checkSetupActive, nil),
	rule(listSetupBench, (*Game).checkSetupBench, nil),
	rule(func(_ *Game, p int) []SetupReady { return []SetupReady{{Player: p}} }, (*Game).checkSetupReady, nil),
	rule(listChooseActive, (*Game).checkChooseActive, nil),
	rule(listPlayBasic, (*Game).checkPlayBasic, nil),
	rule(listAttach, (*Game).checkAttach, nil),
	rule(listEvolve, (*Game).checkEvolve, nil),
	rule(listPlayTrainer, (*Game).checkPlayTrainer, nil),
	rule(listDiscardFossil, (*Game).checkDiscardFossil, nil),
	rule(listRetreat, (*Game).checkRetreat, nil),
	rule(listRetreatDiscard, (*Game).checkRetreatDiscard, nil),
	rule(listUseAbility, (*Game).checkUseAbility, nil),
	rule(listAttack, (*Game).checkAttack, nil),
	rule(func(_ *Game, p int) []EndTurn { return []EndTurn{{Player: p}} }, (*Game).checkEndTurn, nil),
	rule(func(_ *Game, p int) []Concede { return []Concede{{Player: p}} }, (*Game).checkConcede, nil),
}

// LegalMoves returns every move player p may submit right now, each fully
// specified (card and target), with duplicates removed. Empty when p has
// nothing to do.
func (g *Game) LegalMoves(p int) []Event {
	var out []Event
	for _, r := range moveRules {
		seen := map[string]bool{}
		for _, ev := range r.list(g, p) {
			k := r.key(ev)
			if seen[k] || r.check(g, ev) != nil {
				continue
			}
			seen[k] = true
			out = append(out, ev)
		}
	}
	return out
}

// legality returns nil if ev is in LegalMoves, otherwise the reason it isn't.
func (g *Game) legality(ev Event) error {
	r, ok := ruleFor(ev)
	if !ok {
		return errors.New("unknown event")
	}
	if err := r.check(g, ev); err != nil {
		return err
	}
	k := r.key(ev)
	for p := 0; p < 2; p++ {
		for _, legal := range g.LegalMoves(p) {
			if r.is(legal) && r.key(legal) == k {
				return nil
			}
		}
	}
	return errors.New("that move is not available right now")
}

func ruleFor(ev Event) (moveRule, bool) {
	for _, r := range moveRules {
		if r.is(ev) {
			return r, true
		}
	}
	return moveRule{}, false
}

// --- gates: who may act at all ---

func (g *Game) onTurn(p int) error {
	if g.S.Pending != nil {
		return fmt.Errorf("waiting on %s to answer a prompt", side(g.S.Pending.Player))
	}
	if g.S.Phase != PhaseMain {
		return errors.New("not in the main phase")
	}
	if p != g.S.Active {
		return fmt.Errorf("it is not %s's turn", side(p))
	}
	return nil
}

func (g *Game) onPrompt(p int, k PromptKind) error {
	if g.S.Pending == nil {
		return errors.New("no prompt is pending")
	}
	if g.S.Pending.Kind != k || g.S.Pending.Player != p {
		return errBadPrompt
	}
	return nil
}

// onSetup gates setup placement. Setup is simultaneous (spec S5): any player
// not yet Ready may place, whichever player the setup prompt names.
func (g *Game) onSetup(p int) error {
	if g.S.Phase != PhaseSetup || g.S.Pending == nil || g.S.Pending.Kind != PromptSetup {
		return errors.New("not in setup")
	}
	if p != 0 && p != 1 {
		return fmt.Errorf("no player %d", p)
	}
	if g.S.setup[p] {
		return fmt.Errorf("%s is already Ready", side(p))
	}
	return nil
}

// --- candidate lists (legal or not; the checks filter them) ---

func listSetupActive(g *Game, p int) []SetupActive {
	var out []SetupActive
	for _, c := range g.S.Players[p].Hand {
		out = append(out, SetupActive{Player: p, CardID: c.ID})
	}
	return out
}

func listSetupBench(g *Game, p int) []SetupBench {
	var out []SetupBench
	for _, c := range g.S.Players[p].Hand {
		for slot := 0; slot < BenchSlots; slot++ {
			out = append(out, SetupBench{Player: p, CardID: c.ID, Slot: slot})
		}
	}
	return out
}

func listChooseActive(g *Game, p int) []ChooseNewActive {
	var out []ChooseNewActive
	for slot := 0; slot < BenchSlots; slot++ {
		out = append(out, ChooseNewActive{Player: p, BenchIndex: slot})
	}
	return out
}

// listPlayBasic offers every hand card into every slot; one placement per
// empty slot survives the check.
func listPlayBasic(g *Game, p int) []PlayBasic {
	var out []PlayBasic
	for _, c := range g.S.Players[p].Hand {
		for slot := 0; slot < BenchSlots; slot++ {
			out = append(out, PlayBasic{Player: p, CardID: c.ID, Slot: slot})
		}
	}
	return out
}

// targets lists every in-play spot: 0 the Active, n bench slot n-1.
func targets() []int {
	out := make([]int, 0, 1+BenchSlots)
	for t := 0; t <= BenchSlots; t++ {
		out = append(out, t)
	}
	return out
}

func listAttach(g *Game, p int) []AttachEnergy {
	var out []AttachEnergy
	for _, t := range targets() {
		out = append(out, AttachEnergy{Player: p, Target: t})
	}
	return out
}

func listEvolve(g *Game, p int) []Evolve {
	var out []Evolve
	for _, c := range g.S.Players[p].Hand {
		for _, t := range targets() {
			out = append(out, Evolve{Player: p, HandCardID: c.ID, Target: t})
		}
	}
	return out
}

// listRetreatDiscard is every distinct set of energy on p's Active that pays
// its retreat cost, while p owes that choice.
func listRetreatDiscard(g *Game, p int) []RetreatDiscard {
	pr := g.S.Pending
	if pr == nil || pr.Kind != PromptRetreatEnergy || pr.Player != p {
		return nil
	}
	mon := g.S.Players[p].Active
	var out []RetreatDiscard
	for _, set := range energySets(mon.Energy, energyOrder, g.RetreatCost(p)) {
		out = append(out, RetreatDiscard{Player: p, Energy: set})
	}
	return out
}

// energySets lists every multiset of n energy drawable from have, using only
// the types in order, each set listed in that order.
func energySets(have map[Energy]int, order []Energy, n int) [][]Energy {
	if n == 0 {
		return [][]Energy{{}}
	}
	if len(order) == 0 {
		return nil
	}
	t, rest := order[0], order[1:]
	var out [][]Energy
	for k := min(have[t], n); k >= 0; k-- {
		for _, tail := range energySets(have, rest, n-k) {
			set := make([]Energy, 0, n)
			for i := 0; i < k; i++ {
				set = append(set, t)
			}
			out = append(out, append(set, tail...))
		}
	}
	return out
}

func listRetreat(g *Game, p int) []Retreat {
	var out []Retreat
	for slot := 0; slot < BenchSlots; slot++ {
		out = append(out, Retreat{Player: p, BenchIndex: slot})
	}
	return out
}

func listAttack(g *Game, p int) []UseAttack {
	pl := g.S.Players[p]
	if pl.Active == nil {
		return nil
	}
	var out []UseAttack
	for i := range pl.Active.Card.Attacks {
		out = append(out, UseAttack{Player: p, Index: i})
	}
	return out
}

// --- checks: nil when legal, otherwise the reason ---

// checkSetupCard says why hand card id can't be placed at setup, or nil.
func (g *Game) checkSetupCard(p int, id string) error {
	if err := g.onSetup(p); err != nil {
		return err
	}
	pl := &g.S.Players[p]
	idx := indexInHand(pl.Hand, id)
	if idx < 0 {
		return errors.New("card is not in hand")
	}
	if !pl.Hand[idx].IsBasicPokemon() {
		return errors.New("only Basic Pokémon may be placed at setup")
	}
	return nil
}

func (g *Game) checkSetupActive(e SetupActive) error {
	if err := g.checkSetupCard(e.Player, e.CardID); err != nil {
		return err
	}
	if g.S.Players[e.Player].Active != nil {
		return errors.New("the Active is already placed")
	}
	return nil
}

func (g *Game) checkSetupBench(e SetupBench) error {
	if err := g.checkSetupCard(e.Player, e.CardID); err != nil {
		return err
	}
	if e.Slot < 0 || e.Slot >= BenchSlots {
		return errors.New("invalid bench slot")
	}
	if g.S.Players[e.Player].Bench[e.Slot] != nil {
		return errors.New("that bench slot is taken")
	}
	return nil
}

func (g *Game) checkSetupReady(e SetupReady) error {
	if err := g.onSetup(e.Player); err != nil {
		return err
	}
	if g.S.Players[e.Player].Active == nil {
		return errors.New("place an Active before pressing Ready")
	}
	return nil
}

func (g *Game) checkChooseActive(e ChooseNewActive) error {
	if err := g.onPrompt(e.Player, PromptNewActive); err != nil {
		return err
	}
	if benchAt(&g.S.Players[e.Player], e.BenchIndex) == nil {
		return errors.New("no Pokémon in that bench slot")
	}
	return nil
}

var errSlotTaken = errors.New("that bench slot is taken")

func (g *Game) checkPlayBasic(e PlayBasic) error {
	if err := g.onTurn(e.Player); err != nil {
		return err
	}
	pl := &g.S.Players[e.Player]
	idx := indexInHand(pl.Hand, e.CardID)
	if idx < 0 {
		return errors.New("card is not in hand")
	}
	if !pl.Hand[idx].IsBasicPokemon() {
		return errors.New("only Basic Pokémon may be played to the bench")
	}
	if e.Slot < 0 || e.Slot >= BenchSlots {
		return errors.New("invalid bench slot")
	}
	if pl.Bench[e.Slot] != nil {
		return errSlotTaken
	}
	return nil
}

func (g *Game) checkAttach(e AttachEnergy) error {
	if err := g.onTurn(e.Player); err != nil {
		return err
	}
	pl := &g.S.Players[e.Player]
	if pl.EnergyUsed {
		return errors.New("energy was already attached this turn")
	}
	if pl.EnergyZone == NoEnergy {
		return errors.New("no energy is available this turn")
	}
	if targetMon(pl, e.Target) == nil {
		return errors.New("invalid energy target")
	}
	return nil
}

func (g *Game) checkEvolve(e Evolve) error {
	ev, err := g.evolveGate(e.Player, e.HandCardID)
	if err != nil {
		return err
	}
	mon := targetMon(&g.S.Players[e.Player], e.Target)
	if mon == nil {
		return errors.New("invalid evolution target")
	}
	if mon.Card.Name != ev.EvolvesFrom {
		return fmt.Errorf("%s does not evolve from %s", ev.Name, mon.Card.Name)
	}
	if g.hasHook(HookEvolveEarly, g.holder(e.Player, e.Target), nil) {
		return nil
	}
	// T8: game turn 1 is the first player's first turn, turn 2 the second's.
	if g.S.Turn <= 2 {
		return errFirstTurnEvolve
	}
	if mon.PlacedTurn == g.S.Turn {
		return errors.New("a Pokémon cannot evolve the turn it was played")
	}
	return nil
}

var errFirstTurnEvolve = errors.New("no Pokémon may evolve on the first turn")

// evolveGate is every Evolve check that doesn't depend on the target.
func (g *Game) evolveGate(p int, handCardID string) (Card, error) {
	if err := g.onTurn(p); err != nil {
		return Card{}, err
	}
	pl := &g.S.Players[p]
	hidx := indexInHand(pl.Hand, handCardID)
	if hidx < 0 {
		return Card{}, errors.New("evolution card is not in hand")
	}
	ev := pl.Hand[hidx]
	if ev.IsTrainer() || ev.Stage == Basic {
		return Card{}, errors.New("that card is not an evolution")
	}
	return ev, nil
}

func (g *Game) checkRetreat(e Retreat) error {
	if err := g.onTurn(e.Player); err != nil {
		return err
	}
	pl := &g.S.Players[e.Player]
	if pl.Retreated {
		return errors.New("already retreated this turn")
	}
	if pl.Active == nil {
		return errors.New("no Active Pokémon to retreat")
	}
	if pl.Active.Card.Fossil {
		return fmt.Errorf("%s can't retreat", pl.Active.Card.Name)
	}
	if pl.Active.cantAct() {
		return fmt.Errorf("%s is %s and can't retreat", pl.Active.Card.Name, pl.Active.special)
	}
	if benchAt(pl, e.BenchIndex) == nil {
		return errors.New("no Pokémon in that bench slot")
	}
	if pl.Active.totalEnergy() < g.RetreatCost(e.Player) {
		return errors.New("not enough energy to retreat")
	}
	return nil
}

func (g *Game) checkRetreatDiscard(e RetreatDiscard) error {
	if err := g.onPrompt(e.Player, PromptRetreatEnergy); err != nil {
		return err
	}
	mon := g.S.Players[e.Player].Active
	if cost := g.RetreatCost(e.Player); len(e.Energy) != cost {
		return fmt.Errorf("retreating costs %d energy", cost)
	}
	need := map[Energy]int{}
	for _, t := range e.Energy {
		need[t]++
		if need[t] > mon.Energy[t] {
			return fmt.Errorf("not enough %s energy attached", t)
		}
	}
	return nil
}

func (g *Game) checkAttack(e UseAttack) error {
	if err := g.onTurn(e.Player); err != nil {
		return err
	}
	pl := &g.S.Players[e.Player]
	if pl.Active == nil {
		return errors.New("no Active Pokémon to attack with")
	}
	if pl.Active.cantAct() {
		return fmt.Errorf("%s is %s and can't attack", pl.Active.Card.Name, pl.Active.special)
	}
	if e.Index < 0 || e.Index >= len(pl.Active.Card.Attacks) {
		return errors.New("invalid attack index")
	}
	if !canPay(pl.Active.Energy, pl.Active.Card.Attacks[e.Index].Cost) {
		return errors.New("not enough energy for that attack")
	}
	if g.S.Players[1-e.Player].Active == nil {
		return errors.New("there is no defending Pokémon")
	}
	return nil
}

func (g *Game) checkEndTurn(e EndTurn) error { return g.onTurn(e.Player) }

// checkConcede: either player, at any time, prompts included (spec E7).
func (g *Game) checkConcede(e Concede) error {
	if g.S.Phase == PhaseOver {
		return errors.New("game is over")
	}
	if e.Player != 0 && e.Player != 1 {
		return fmt.Errorf("no player %d", e.Player)
	}
	return nil
}
