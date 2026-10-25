package main

import (
	"testing"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// One test per wording in batch 3.9, each on a real printing.

// put places a fresh copy of card at player p's pos (0 Active, n bench slot n-1).
func put(g *engine.Game, p, pos int, c engine.Card) *engine.InPlay {
	m := &engine.InPlay{Card: c, Energy: map[engine.Energy]int{}}
	pl := &g.S.Players[p]
	if pos == 0 {
		pl.Active = m
	} else {
		pl.Bench[pos-1] = m
	}
	return m
}

func play(t *testing.T, g *engine.Game, p int, id string, at engine.Spot) {
	t.Helper()
	if err := g.Submit(engine.PlayTrainer{Player: p, CardID: id, Target: at}); err != nil {
		t.Fatalf("play %s: %v", id, err)
	}
}

// fuel gives mon energy for every entry of attack i's cost (Fire for [C]).
func fuel(mon *engine.InPlay, i int) {
	for _, e := range mon.Card.Attacks[i].Cost {
		if e == engine.Colorless {
			e = engine.Fire
		}
		mon.Energy[e]++
	}
}

// buffedAttack plays supporter, then attacks with attacker's attack i into a
// Torkoal (Water weakness) and returns the damage done.
func buffedAttack(t *testing.T, supporter, attacker string, i int) (got, base int) {
	t.Helper()
	useRealCards(t)
	g, p := trainerGame(t, realCard(t, supporter))
	mon := put(g, p, 0, realCard(t, attacker))
	fuel(mon, i)
	def := g.S.Players[1-p].Active
	play(t, g, p, supporter, engine.NoSpot)
	mustSubmit(t, g, engine.UseAttack{Player: p, Index: i})
	return def.Damage, mon.Card.Attacks[i].Damage
}

func mustSubmit(t *testing.T, g *engine.Game, ev engine.Event) {
	t.Helper()
	if err := g.Submit(ev); err != nil {
		t.Fatalf("submit %#v: %v", ev, err)
	}
}

func TestBlaineAddsThirty(t *testing.T) {
	if got, base := buffedAttack(t, "A1-221", "B3b-009", 0); got != base+30 { // Ninetales
		t.Errorf("Ninetales did %d, want %d", got, base+30)
	}
}

func TestCynthiaAddsFifty(t *testing.T) {
	if got, base := buffedAttack(t, "A2-152", "A4-080", 0); got != base+50 { // Togekiss
		t.Errorf("Togekiss did %d, want %d", got, base+50)
	}
}

func TestGiovanniAddsTen(t *testing.T) {
	if got, base := buffedAttack(t, "A1-223", "A3b-010", 0); got != base+10 { // Torkoal
		t.Errorf("Torkoal did %d, want %d", got, base+10)
	}
}

func TestHauAddsThirty(t *testing.T) {
	if got, base := buffedAttack(t, "A3b-068", "A3-012", 1); got != base+30 { // Decidueye ex
		t.Errorf("Decidueye ex did %d, want %d", got, base+30)
	}
}

// A "this turn" buff is gone by the player's next turn.
func TestGiovanniEndsWithTheTurn(t *testing.T) {
	useRealCards(t)
	g, p := trainerGame(t, realCard(t, "A1-223"))
	play(t, g, p, "A1-223", engine.NoSpot)
	mustSubmit(t, g, engine.EndTurn{Player: p})
	mustSubmit(t, g, engine.EndTurn{Player: 1 - p})
	fuel(g.S.Players[p].Active, 0)
	def := g.S.Players[1-p].Active
	mustSubmit(t, g, engine.UseAttack{Player: p, Index: 0})
	if def.Damage != 70 {
		t.Errorf("Torkoal did %d next turn, want 70", def.Damage)
	}
}

func TestBarryCutsTwoColorless(t *testing.T) {
	useRealCards(t)
	g, p := trainerGame(t, realCard(t, "A2a-074"))
	snorlax := put(g, p, 0, realCard(t, "B3b-055")) // Mega Punch [CCC]
	snorlax.Energy[engine.Fire] = 1
	atk := engine.UseAttack{Player: p, Index: 0}
	if g.Submit(atk) == nil {
		t.Fatal("Mega Punch was usable with 1 Energy before Barry")
	}
	play(t, g, p, "A2a-074", engine.NoSpot)
	mustSubmit(t, g, atk)
}

// guardedHit: p plays supporter and ends the turn; the opponent's attacker
// (attack i, fully fueled) then hits p's Active, defender. Returns its damage.
func guardedHit(t *testing.T, supporter, defender, attacker string, i int) *engine.InPlay {
	t.Helper()
	useRealCards(t)
	g, p := trainerGame(t, realCard(t, supporter))
	def := put(g, p, 0, realCard(t, defender))
	play(t, g, p, supporter, engine.NoSpot)
	mustSubmit(t, g, engine.EndTurn{Player: p})
	att := put(g, 1-p, 0, realCard(t, attacker))
	fuel(att, i)
	mustSubmit(t, g, engine.UseAttack{Player: 1 - p, Index: i})
	return def
}

func TestBlueTakesTenLess(t *testing.T) {
	if d := guardedHit(t, "A1a-067", "A3b-010", "A3b-010", 0).Damage; d != 60 {
		t.Errorf("Torkoal took %d, want 70−10", d)
	}
}

func TestAdamanTakesTwentyLess(t *testing.T) {
	useRealCards(t)
	aggron := realCard(t, "A3a-050")
	want := 70 - 20
	if aggron.Weakness == engine.Fire {
		want += 20
	}
	if d := guardedHit(t, "A2a-075", "A3a-050", "A3b-010", 0).Damage; d != want {
		t.Errorf("Aggron took %d, want %d", d, want)
	}
}

func TestCherenBlocksExAttacks(t *testing.T) {
	if d := guardedHit(t, "B3-151", "B1-200", "A1-129", 0).Damage; d != 0 { // Watchog vs Mewtwo ex
		t.Errorf("Watchog took %d from Mewtwo ex, want 0", d)
	}
}

func TestHalaLeavesTenHP(t *testing.T) {
	useRealCards(t)
	g, p := trainerGame(t, realCard(t, "B1-222"))
	hariyama := put(g, p, 0, realCard(t, "A3-091"))
	hariyama.Damage = hariyama.Card.HP - 20
	play(t, g, p, "B1-222", engine.NoSpot)
	mustSubmit(t, g, engine.EndTurn{Player: p})
	fuel(g.S.Players[1-p].Active, 0)
	mustSubmit(t, g, engine.UseAttack{Player: 1 - p, Index: 0})
	if a := g.S.Players[p].Active; a != hariyama || a.Card.HP-a.Damage != 10 {
		t.Errorf("Hariyama was knocked out or not left at 10 HP")
	}
}

func TestErikaHealsGrass(t *testing.T) {
	useRealCards(t)
	g, p := trainerGame(t, realCard(t, "A1-219"))
	b := put(g, p, 1, realCard(t, "B1a-001")) // Bulbasaur
	b.Damage = 30
	play(t, g, p, "A1-219", engine.Spot{Player: p, Pos: 1})
	if b.Damage != 0 {
		t.Errorf("Bulbasaur has %d damage, want 0", b.Damage)
	}
}

func TestDianthaHealsAndDiscards(t *testing.T) {
	useRealCards(t)
	g, p := trainerGame(t, realCard(t, "B2-149"))
	m := put(g, p, 0, realCard(t, "A1-129")) // Mewtwo ex
	m.Damage, m.Energy[engine.Psychic] = 100, 3
	play(t, g, p, "B2-149", engine.Spot{Player: p, Pos: 0})
	if m.Damage != 10 || m.Energy[engine.Psychic] != 1 {
		t.Errorf("Mewtwo ex damage %d, [P] %d; want 10 and 1", m.Damage, m.Energy[engine.Psychic])
	}
}

func TestBrockAttachesFighting(t *testing.T) {
	useRealCards(t)
	g, p := trainerGame(t, realCard(t, "A1-224"))
	onix := put(g, p, 2, realCard(t, "B1a-038"))
	play(t, g, p, "A1-224", engine.Spot{Player: p, Pos: 2})
	if onix.Energy[engine.Fighting] != 1 {
		t.Errorf("Onix has %d [F], want 1", onix.Energy[engine.Fighting])
	}
}

func TestFantinaAttachesToEach(t *testing.T) {
	useRealCards(t)
	g, p := trainerGame(t, realCard(t, "B1-224"))
	d := put(g, p, 0, realCard(t, "B1-108"))  // Drifblim
	m := put(g, p, 3, realCard(t, "A4a-033")) // Mismagius
	play(t, g, p, "B1-224", engine.NoSpot)
	if d.Energy[engine.Psychic] != 1 || m.Energy[engine.Psychic] != 1 {
		t.Errorf("[P] on Drifblim %d, Mismagius %d; want 1 each", d.Energy[engine.Psychic], m.Energy[engine.Psychic])
	}
}

func TestBuddingExpeditionerReturnsMewEx(t *testing.T) {
	useRealCards(t)
	g, p := trainerGame(t, realCard(t, "A1a-066"))
	put(g, p, 0, realCard(t, "A1a-032"))
	bench := put(g, p, 2, realCard(t, "A3b-010"))
	play(t, g, p, "A1a-066", engine.NoSpot)
	pl := g.S.Players[p]
	if pl.Active != bench || indexOf(pl.Hand, "A1a-032") < 0 {
		t.Errorf("Mew ex not in hand with Torkoal promoted")
	}
}

func indexOf(cards []engine.Card, id string) int {
	for i, c := range cards {
		if c.ID == id {
			return i
		}
	}
	return -1
}

func TestCalemDrawsPerMega(t *testing.T) {
	useRealCards(t)
	g, p := trainerGame(t, realCard(t, "B2b-068"))
	put(g, p, 0, realCard(t, "B1-151"))   // Mega Absol ex
	put(g, 1-p, 1, realCard(t, "B1-151")) // and the opponent's
	hand := len(g.S.Players[p].Hand)
	play(t, g, p, "B2b-068", engine.NoSpot)
	if n := len(g.S.Players[p].Hand); n != hand-1+2 {
		t.Errorf("hand %d→%d, want +2 drawn", hand, n)
	}
}

func TestCelesticTownElderTakesBasic(t *testing.T) {
	useRealCards(t)
	g, p := trainerGame(t, realCard(t, "A2a-073"))
	pl := &g.S.Players[p]
	pl.Discard = []engine.Card{realCard(t, "A4b-373"), realCard(t, "B1a-001")}
	play(t, g, p, "A2a-073", engine.NoSpot)
	if indexOf(pl.Hand, "B1a-001") < 0 || indexOf(pl.Discard, "B1a-001") >= 0 {
		t.Error("Bulbasaur not moved from discard to hand")
	}
}

func TestFisherTakesWaterPerHeads(t *testing.T) {
	useRealCards(t)
	g, p := trainerGame(t, realCard(t, "A4-159"))
	pl := &g.S.Players[p]
	sq := realCard(t, "A1-053")
	pl.Discard = []engine.Card{sq, sq, sq, realCard(t, "B1a-001")}
	hand := len(pl.Hand)
	play(t, g, p, "A4-159", engine.NoSpot)
	heads := 0
	for _, f := range g.Actions[len(g.Actions)-1].Flips {
		if f.Heads {
			heads++
		}
	}
	if len(pl.Hand) != hand-1+heads || indexOf(pl.Hand, "B1a-001") >= 0 {
		t.Errorf("hand %d→%d with %d heads", hand, len(pl.Hand), heads)
	}
}

func TestCopycatMatchesOpponentHand(t *testing.T) {
	useRealCards(t)
	g, p := trainerGame(t, realCard(t, "B1-225"))
	opp := len(g.S.Players[1-p].Hand)
	play(t, g, p, "B1-225", engine.NoSpot)
	if n := len(g.S.Players[p].Hand); n != opp {
		t.Errorf("hand %d, want %d (opponent's)", n, opp)
	}
}

func TestCyrusSwitchesInDamaged(t *testing.T) {
	useRealCards(t)
	g, p := trainerGame(t, realCard(t, "A2-150"))
	hurt := put(g, 1-p, 2, realCard(t, "A3b-010"))
	hurt.Damage = 10
	play(t, g, p, "A2-150", engine.Spot{Player: 1 - p, Pos: 2})
	if g.S.Players[1-p].Active != hurt {
		t.Error("damaged Torkoal not switched in")
	}
}

func TestDawnMovesBenchEnergy(t *testing.T) {
	useRealCards(t)
	g, p := trainerGame(t, realCard(t, "A2-154"))
	b := put(g, p, 1, realCard(t, "A3b-010"))
	b.Energy[engine.Fire] = 1
	act := g.S.Players[p].Active
	before := act.Energy[engine.Fire]
	play(t, g, p, "A2-154", engine.Spot{Player: p, Pos: 1})
	if b.Energy[engine.Fire] != 0 || act.Energy[engine.Fire] != before+1 {
		t.Error("energy not moved from Bench to Active")
	}
}

func TestAcerolaMovesForty(t *testing.T) {
	useRealCards(t)
	g, p := trainerGame(t, realCard(t, "A3-148"))
	mim := put(g, p, 1, realCard(t, "A3-083")) // Mimikyu
	mim.Damage = 50
	def := g.S.Players[1-p].Active
	play(t, g, p, "A3-148", engine.Spot{Player: p, Pos: 1})
	if mim.Damage != 10 || def.Damage != 40 {
		t.Errorf("Mimikyu %d, opponent's Active %d; want 10 and 40", mim.Damage, def.Damage)
	}
}

func TestGuzmaDiscardsOpponentTools(t *testing.T) {
	useRealCards(t)
	g, p := trainerGame(t, realCard(t, "A3-151"))
	helmet := realCard(t, "A4b-322")
	g.S.Players[1-p].Active.Tool = &helmet
	play(t, g, p, "A3-151", engine.NoSpot)
	if g.S.Players[1-p].Active.Tool != nil || indexOf(g.S.Players[1-p].Discard, "A4b-322") < 0 {
		t.Error("opponent's Tool not discarded")
	}
}

func TestElesaReturnsAllTools(t *testing.T) {
	useRealCards(t)
	g, p := trainerGame(t, realCard(t, "B3b-066"))
	helmet := realCard(t, "A4b-322")
	h1, h2 := helmet, helmet
	g.S.Players[p].Active.Tool, g.S.Players[1-p].Active.Tool = &h1, &h2
	play(t, g, p, "B3b-066", engine.NoSpot)
	for q := 0; q < 2; q++ {
		if g.S.Players[q].Active.Tool != nil || indexOf(g.S.Players[q].Hand, "A4b-322") < 0 {
			t.Errorf("player %d's Tool not back in hand", q)
		}
	}
}

// searchWith plays supporter with deck as the player's deck; returns the hand.
func searchWith(t *testing.T, supporter string, deck ...string) (*engine.Game, []engine.Card) {
	t.Helper()
	useRealCards(t)
	g, p := trainerGame(t, realCard(t, supporter))
	pl := &g.S.Players[p]
	pl.Deck = nil
	for _, id := range deck {
		pl.Deck = append(pl.Deck, realCard(t, id))
	}
	play(t, g, p, supporter, engine.NoSpot)
	return g, pl.Hand
}

func TestCabbieFindsStadium(t *testing.T) {
	if _, hand := searchWith(t, "B3-150", "A3b-010", "B3-155", "A3b-010"); indexOf(hand, "B3-155") < 0 {
		t.Error("Bounded Field not put into hand")
	}
}

func TestGladionFindsTypeNull(t *testing.T) {
	if _, hand := searchWith(t, "A3a-067", "A3b-010", "A3a-060", "A3b-010"); indexOf(hand, "A3a-060") < 0 {
		t.Error("Type: Null not put into hand")
	}
}

func TestClemontFindsTwo(t *testing.T) {
	_, hand := searchWith(t, "B1a-068", "A3b-010", "A1-098", "A3b-010", "B1a-066")
	if indexOf(hand, "A1-098") < 0 || indexOf(hand, "B1a-066") < 0 {
		t.Error("Magneton and Clemont's Backpack not both put into hand")
	}
}

func TestArvenFindsItemOrTool(t *testing.T) {
	g, hand := searchWith(t, "B2a-091", "A2b-111", "A4b-322", "A3b-010")
	want := "A4b-322" // tails: Rocky Helmet
	if g.Actions[len(g.Actions)-1].Flips[0].Heads {
		want = "A2b-111" // heads: Poké Ball
	}
	if indexOf(hand, want) < 0 {
		t.Errorf("hand lacks %s", want)
	}
}
