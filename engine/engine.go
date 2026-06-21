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
// prompts. All randomness derives from seed.
func NewGame(seed uint64, deck0, deck1 []Card) *Game {
	g := &Game{Seed: seed, rng: newRng(seed)}
	g.decks[0] = deck0
	g.decks[1] = deck1
	st := &State{Winner: -1, Phase: PhaseSetup}
	g.S = st

	for i := 0; i < 2; i++ {
		pl := &st.Players[i]
		pl.Deck = append([]Card(nil), g.decks[i]...)
		pl.EnergyTypes = energyTypesOf(pl.Deck)
		pl.EnergyZone = NoEnergy
		g.dealOpeningHand(i)
	}
	st.firstPlayer = g.rng.intn(2)
	st.Pending = &Prompt{Player: 0, Kind: PromptSetup}
	g.narrate("Game start (seed %d). %s goes first.", seed, side(st.firstPlayer))
	return g
}

func (g *Game) dealOpeningHand(i int) {
	pl := &g.S.Players[i]
	for attempt := 0; attempt < 200; attempt++ {
		g.rng.shuffle(len(pl.Deck), func(a, b int) {
			pl.Deck[a], pl.Deck[b] = pl.Deck[b], pl.Deck[a]
		})
		hasBasic := false
		for _, c := range pl.Deck[:min(5, len(pl.Deck))] {
			if c.Stage == Basic {
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
	if err := g.validateAndApply(ev); err != nil {
		return err
	}
	g.Events = append(g.Events, ev)
	return nil
}

func (g *Game) validateAndApply(ev Event) error {
	st := g.S
	// Mode 3: blocked on a prompt — only the owed response is accepted.
	if st.Pending != nil {
		switch e := ev.(type) {
		case SetupPlace:
			if st.Pending.Kind != PromptSetup || e.Player != st.Pending.Player {
				return errBadPrompt
			}
			return g.applySetup(e)
		case ChooseNewActive:
			if st.Pending.Kind != PromptNewActive || e.Player != st.Pending.Player {
				return errBadPrompt
			}
			return g.applyChooseActive(e)
		default:
			return fmt.Errorf("waiting on %s to answer a prompt", side(st.Pending.Player))
		}
	}
	// Modes 1: free-form intents from the active player.
	switch e := ev.(type) {
	case PlayBasic:
		return g.requireTurn(e.Player, func() error { return g.applyPlayBasic(e) })
	case AttachEnergy:
		return g.requireTurn(e.Player, func() error { return g.applyAttach(e) })
	case Evolve:
		return g.requireTurn(e.Player, func() error { return g.applyEvolve(e) })
	case Retreat:
		return g.requireTurn(e.Player, func() error { return g.applyRetreat(e) })
	case UseAttack:
		return g.requireTurn(e.Player, func() error { return g.applyAttack(e) })
	case EndTurn:
		return g.requireTurn(e.Player, func() error { g.concludeTurn(); return nil })
	case SetupPlace, ChooseNewActive:
		return errors.New("no prompt is pending")
	default:
		return errors.New("unknown event")
	}
}

func (g *Game) requireTurn(p int, fn func() error) error {
	if g.S.Phase != PhaseMain {
		return errors.New("not in the main phase")
	}
	if p != g.S.Active {
		return fmt.Errorf("it is not %s's turn", side(p))
	}
	return fn()
}

// --- Intent handlers ---

func (g *Game) applySetup(e SetupPlace) error {
	pl := &g.S.Players[e.Player]
	idx := indexInHand(pl.Hand, e.ActiveCardID)
	if idx < 0 {
		return errors.New("active card is not in hand")
	}
	if pl.Hand[idx].Stage != Basic {
		return errors.New("the Active must be a Basic Pokémon")
	}
	if len(e.BenchCardIDs) > 3 {
		return errors.New("the bench holds at most 3")
	}
	remove := map[int]bool{idx: true}
	var benchCards []Card
	for _, bid := range e.BenchCardIDs {
		bi := indexInHandExcept(pl.Hand, bid, remove)
		if bi < 0 {
			return errors.New("bench card is not in hand")
		}
		if pl.Hand[bi].Stage != Basic {
			return errors.New("only Basics may be benched")
		}
		remove[bi] = true
		benchCards = append(benchCards, pl.Hand[bi])
	}

	activeCard := pl.Hand[idx]
	pl.Hand = removeFromHand(pl.Hand, remove)
	pl.Active = newInPlay(activeCard, 0)
	for _, c := range benchCards {
		pl.Bench = append(pl.Bench, newInPlay(c, 0))
	}
	g.narrate("%s placed %s as Active with %d on the bench.", side(e.Player), activeCard.Name, len(benchCards))
	g.S.setup[e.Player] = true

	switch {
	case !g.S.setup[0]:
		g.S.Pending = &Prompt{Player: 0, Kind: PromptSetup}
	case !g.S.setup[1]:
		g.S.Pending = &Prompt{Player: 1, Kind: PromptSetup}
	default:
		g.S.Pending = nil
		g.S.Phase = PhaseMain
		g.S.Active = g.S.firstPlayer
		g.S.Turn = 1
		g.startTurn(g.S.Active)
	}
	return nil
}

func (g *Game) applyPlayBasic(e PlayBasic) error {
	pl := &g.S.Players[e.Player]
	idx := indexInHand(pl.Hand, e.CardID)
	if idx < 0 {
		return errors.New("card is not in hand")
	}
	if pl.Hand[idx].Stage != Basic {
		return errors.New("only Basic Pokémon may be played to the bench")
	}
	if len(pl.Bench) >= 3 {
		return errors.New("the bench is full")
	}
	c := pl.Hand[idx]
	pl.Hand = removeAt(pl.Hand, idx)
	pl.Bench = append(pl.Bench, newInPlay(c, g.S.Turn))
	g.narrate("%s played %s to the bench.", side(e.Player), c.Name)
	return nil
}

func (g *Game) applyAttach(e AttachEnergy) error {
	pl := &g.S.Players[e.Player]
	if pl.EnergyUsed {
		return errors.New("energy was already attached this turn")
	}
	if pl.EnergyZone == NoEnergy {
		return errors.New("no energy is available this turn")
	}
	mon := targetMon(pl, e.Target)
	if mon == nil {
		return errors.New("invalid energy target")
	}
	mon.Energy[pl.EnergyZone]++
	g.narrate("%s attached %s energy to %s.", side(e.Player), pl.EnergyZone, mon.Card.Name)
	pl.EnergyZone = NoEnergy
	pl.EnergyUsed = true
	return nil
}

func (g *Game) applyEvolve(e Evolve) error {
	pl := &g.S.Players[e.Player]
	hidx := indexInHand(pl.Hand, e.HandCardID)
	if hidx < 0 {
		return errors.New("evolution card is not in hand")
	}
	ev := pl.Hand[hidx]
	if ev.Stage == Basic {
		return errors.New("that card is not an evolution")
	}
	mon := targetMon(pl, e.Target)
	if mon == nil {
		return errors.New("invalid evolution target")
	}
	if mon.Card.Name != ev.EvolvesFrom {
		return fmt.Errorf("%s does not evolve from %s", ev.Name, mon.Card.Name)
	}
	if mon.PlacedTurn == g.S.Turn {
		return errors.New("a Pokémon cannot evolve the turn it was played")
	}
	if g.S.Turn == 1 {
		return errors.New("no Pokémon may evolve on the first turn")
	}
	pl.Hand = removeAt(pl.Hand, hidx)
	mon.Card = ev
	mon.PlacedTurn = g.S.Turn
	g.narrate("%s evolved into %s.", side(e.Player), ev.Name)
	return nil
}

func (g *Game) applyRetreat(e Retreat) error {
	pl := &g.S.Players[e.Player]
	if pl.Retreated {
		return errors.New("already retreated this turn")
	}
	if pl.Active == nil {
		return errors.New("no Active Pokémon to retreat")
	}
	if e.BenchIndex < 0 || e.BenchIndex >= len(pl.Bench) {
		return errors.New("invalid bench index")
	}
	cost := pl.Active.Card.RetreatCost
	if pl.Active.totalEnergy() < cost {
		return errors.New("not enough energy to retreat")
	}
	discardEnergy(pl.Active, cost)
	pl.Active, pl.Bench[e.BenchIndex] = pl.Bench[e.BenchIndex], pl.Active
	pl.Retreated = true
	g.narrate("%s retreated to %s.", side(e.Player), pl.Active.Card.Name)
	return nil
}

func (g *Game) applyAttack(e UseAttack) error {
	pl := &g.S.Players[e.Player]
	if pl.Active == nil {
		return errors.New("no Active Pokémon to attack with")
	}
	if e.Index < 0 || e.Index >= len(pl.Active.Card.Attacks) {
		return errors.New("invalid attack index")
	}
	atk := pl.Active.Card.Attacks[e.Index]
	if !canPay(pl.Active.Energy, atk.Cost) {
		return errors.New("not enough energy for that attack")
	}
	opp := &g.S.Players[1-e.Player]
	if opp.Active == nil {
		return errors.New("there is no defending Pokémon")
	}
	dmg := atk.Damage
	weak := opp.Active.Card.Weakness != NoEnergy && opp.Active.Card.Weakness == pl.Active.Card.Type
	if weak {
		dmg += 20
	}
	opp.Active.Damage += dmg
	if weak {
		g.narrate("%s's %s used %s for %d (+20 weakness).", side(e.Player), pl.Active.Card.Name, atk.Name, dmg)
	} else {
		g.narrate("%s's %s used %s for %d.", side(e.Player), pl.Active.Card.Name, atk.Name, dmg)
	}

	if opp.Active.remainingHP() <= 0 {
		g.koActive(1 - e.Player)
	}
	// Attacking ends the turn — but a knockout may owe a prompt first.
	if g.S.Phase == PhaseOver {
		return nil
	}
	if g.S.Pending != nil {
		g.S.resume = resumeConclude
		return nil
	}
	g.concludeTurn()
	return nil
}

func (g *Game) applyChooseActive(e ChooseNewActive) error {
	pl := &g.S.Players[e.Player]
	if e.BenchIndex < 0 || e.BenchIndex >= len(pl.Bench) {
		return errors.New("invalid bench index")
	}
	mon := pl.Bench[e.BenchIndex]
	pl.Bench = append(pl.Bench[:e.BenchIndex], pl.Bench[e.BenchIndex+1:]...)
	pl.Active = mon
	g.S.Pending = nil
	g.narrate("%s promoted %s to Active.", side(e.Player), mon.Card.Name)

	if g.S.Phase == PhaseOver {
		return nil
	}
	switch g.S.resume {
	case resumeConclude:
		g.S.resume = resumeNone
		g.concludeTurn()
	case resumeAdvance:
		g.S.resume = resumeNone
		g.advanceTurn()
	}
	return nil
}

// --- Turn flow & resolution ---

// koActive knocks out a player's Active, scores for the opponent, and either
// ends the game or raises a "promote a new Active" prompt (ADR-0007, mode 3).
func (g *Game) koActive(owner int) {
	pl := &g.S.Players[owner]
	mon := pl.Active
	if mon == nil {
		return
	}
	pl.Discard = append(pl.Discard, mon.Card)
	pl.Active = nil
	g.narrate("%s's %s was Knocked Out.", side(owner), mon.Card.Name)

	scorer := 1 - owner
	pts := 1
	if mon.Card.IsEX {
		pts = 2
	}
	g.S.Players[scorer].Points += pts
	g.narrate("%s scored %d (%d/3).", side(scorer), pts, g.S.Players[scorer].Points)

	if g.S.Players[scorer].Points >= 3 {
		g.gameOver(scorer)
		return
	}
	if len(pl.Bench) == 0 {
		g.gameOver(scorer) // no Pokémon left to promote
		return
	}
	g.S.Pending = &Prompt{Player: owner, Kind: PromptNewActive}
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
// automatic), chaining into a knockout prompt if it KOs a Pokémon.
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
			if g.rng.intn(2) == 0 {
				pl.Active.Burned = false
				g.narrate("%s's %s recovered from burn.", side(i), pl.Active.Card.Name)
			}
		}
		if pl.Active.remainingHP() <= 0 {
			g.koActive(i)
			if g.S.Pending != nil || g.S.Phase == PhaseOver {
				return // handle one knockout at a time
			}
		}
	}
}

func (g *Game) advanceTurn() {
	g.S.Active = 1 - g.S.Active
	g.S.Turn++
	g.startTurn(g.S.Active)
}

// startTurn begins a player's turn: reset per-turn flags, draw, and generate
// the Energy Zone energy (skipped on the very first turn of the game).
func (g *Game) startTurn(p int) {
	pl := &g.S.Players[p]
	pl.EnergyUsed = false
	pl.Retreated = false

	if len(pl.Deck) > 0 {
		c := pl.Deck[0]
		pl.Deck = pl.Deck[1:]
		pl.Hand = append(pl.Hand, c)
	}

	if g.S.Turn > 1 {
		et := pl.EnergyTypes
		pl.EnergyZone = et[g.rng.intn(len(et))]
	} else {
		pl.EnergyZone = NoEnergy
	}
	g.narrate("--- Turn %d: %s ---", g.S.Turn, side(p))
}

func (g *Game) gameOver(winner int) {
	g.S.Phase = PhaseOver
	g.S.Winner = winner
	g.S.Pending = nil
	g.narrate("Game over — %s wins %d-%d.", side(winner), g.S.Players[winner].Points, g.S.Players[1-winner].Points)
}

// Replay rebuilds a game from its seed, decks and event log, then applies every
// event. Same inputs → identical game (ADR-0005). This is also how reconnection
// and replay viewing work.
func Replay(seed uint64, deck0, deck1 []Card, events []Event) (*Game, error) {
	g := NewGame(seed, deck0, deck1)
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
	if target-1 < len(pl.Bench) {
		return pl.Bench[target-1]
	}
	return nil
}

func discardEnergy(mon *InPlay, n int) {
	for _, t := range energyOrder {
		for n > 0 && mon.Energy[t] > 0 {
			mon.Energy[t]--
			n--
		}
	}
}

func indexInHand(hand []Card, id string) int {
	for i, c := range hand {
		if c.ID == id {
			return i
		}
	}
	return -1
}

func indexInHandExcept(hand []Card, id string, used map[int]bool) int {
	for i, c := range hand {
		if c.ID == id && !used[i] {
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

func removeFromHand(hand []Card, idxs map[int]bool) []Card {
	out := make([]Card, 0, len(hand))
	for i, c := range hand {
		if !idxs[i] {
			out = append(out, c)
		}
	}
	return out
}
