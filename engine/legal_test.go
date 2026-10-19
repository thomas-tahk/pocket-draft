package engine

import (
	"reflect"
	"testing"
)

// Fixed-state fixtures for LegalMoves. Cards are built here, not dealt, so each
// test states exactly what is on the board.

func legalBasic(id, name string, retreat int, atks ...Attack) Card {
	return Card{ID: id, Name: name, Stage: Basic, HP: 60, Type: Fire, Attacks: atks, Weakness: NoEnergy, RetreatCost: retreat}
}

func legalEvo(id, name, from string) Card {
	return Card{ID: id, Name: name, Stage: Stage1, EvolvesFrom: from, HP: 90, Type: Fire, Weakness: NoEnergy}
}

// mainGame builds a game in the main phase on the given turn, P1 to act, with
// both Actives in play. The caller fills in the rest.
func mainGame(turn int) *Game {
	g := &Game{rng: newRng(1)}
	g.S = &State{Phase: PhaseMain, Turn: turn, Active: 0, Winner: -1}
	for i := range g.S.Players {
		g.S.Players[i].EnergyZone = NoEnergy
		g.S.Players[i].EnergyTypes = []Energy{Fire}
		g.S.Players[i].Active = newInPlay(legalBasic("OPP", "Opp", 1), 0)
	}
	return g
}

var (
	freeHit  = Attack{Name: "Tap", Damage: 10}
	fireBlow = Attack{Name: "Blaze", Cost: []Energy{Fire}, Damage: 40}
)

func assertMoves(t *testing.T, got, want []Event) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("LegalMoves =\n  %#v\nwant\n  %#v", got, want)
	}
}

// Turn 1: no energy, no evolving. Only the free attack, a Basic into each
// empty slot, a free retreat, End Turn and Concede are legal. The opponent may
// only concede.
func TestLegalMovesFirstTurn(t *testing.T) {
	g := mainGame(1)
	pl := &g.S.Players[0]
	pl.Active = newInPlay(legalBasic("A", "Charm", 0, freeHit, fireBlow), 0)
	pl.Bench[BenchBehind] = newInPlay(legalBasic("B", "Bulb", 1), 0)
	pl.Hand = []Card{legalBasic("C", "Squirt", 1), legalEvo("E", "Charmel", "Charm"), legalBasic("C", "Squirt", 1)}

	assertMoves(t, g.LegalMoves(0), []Event{
		PlayBasic{Player: 0, CardID: "C", Slot: BenchLeft},
		PlayBasic{Player: 0, CardID: "C", Slot: BenchRight},
		Retreat{Player: 0, BenchIndex: BenchBehind},
		UseAttack{Player: 0, Index: 0},
		EndTurn{Player: 0},
		Concede{Player: 0},
	})
	assertMoves(t, g.LegalMoves(1), []Event{Concede{Player: 1}})
}

// A later turn: energy can go on any of your Pokémon; evolution only onto the
// matching Pokémon that wasn't placed this turn; retreat needs enough energy;
// a full bench takes no more Basics.
func TestLegalMovesLaterTurn(t *testing.T) {
	g := mainGame(3)
	pl := &g.S.Players[0]
	pl.EnergyZone = Fire
	pl.Active = newInPlay(legalBasic("A", "Charm", 2, freeHit, fireBlow), 1)
	pl.Active.Energy[Fire] = 1
	pl.Bench = [BenchSlots]*InPlay{
		newInPlay(legalBasic("A", "Charm", 1), 1),
		newInPlay(legalBasic("A", "Charm", 1), 3), // played this turn
		newInPlay(legalBasic("B", "Bulb", 1), 1),
	}
	pl.Hand = []Card{legalBasic("C", "Squirt", 1), legalEvo("E", "Charmel", "Charm")}

	assertMoves(t, g.LegalMoves(0), []Event{
		AttachEnergy{Player: 0, Target: 0},
		AttachEnergy{Player: 0, Target: 1},
		AttachEnergy{Player: 0, Target: 2},
		AttachEnergy{Player: 0, Target: 3},
		Evolve{Player: 0, HandCardID: "E", Target: 0},
		Evolve{Player: 0, HandCardID: "E", Target: 1},
		UseAttack{Player: 0, Index: 0},
		UseAttack{Player: 0, Index: 1},
		EndTurn{Player: 0},
		Concede{Player: 0},
	})
}

// A pending knockout prompt: the owner may promote from an occupied slot or
// concede; the other player may only concede.
func TestLegalMovesNewActivePrompt(t *testing.T) {
	g := mainGame(4)
	pl := &g.S.Players[1]
	pl.Active = nil
	pl.Bench[BenchBehind] = newInPlay(legalBasic("B", "Bulb", 1), 1)
	pl.Bench[BenchRight] = newInPlay(legalBasic("C", "Squirt", 1), 1)
	g.S.Pending = &Prompt{Player: 1, Kind: PromptNewActive}

	assertMoves(t, g.LegalMoves(1), []Event{
		ChooseNewActive{Player: 1, BenchIndex: BenchBehind},
		ChooseNewActive{Player: 1, BenchIndex: BenchRight},
		Concede{Player: 1},
	})
	assertMoves(t, g.LegalMoves(0), []Event{Concede{Player: 0}})
}

// Setup is placed one card at a time (spec S5): each Basic to the Active or an
// empty bench slot, and Ready once there is an Active. Evolutions never appear,
// and both players may place at once.
func TestLegalMovesSetup(t *testing.T) {
	// Arrange
	g := &Game{rng: newRng(1)}
	g.S = &State{Phase: PhaseSetup, Winner: -1, Pending: &Prompt{Player: 0, Kind: PromptSetup}}
	g.S.Players[0].Hand = []Card{legalBasic("A", "Charm", 1), legalEvo("E", "Charmel", "Charm"), legalBasic("B", "Bulb", 1)}
	g.S.Players[1].Hand = []Card{legalBasic("C", "Cyn", 1)}

	// Act + Assert: before an Active, there is no Ready.
	assertMoves(t, g.LegalMoves(0), []Event{
		SetupActive{Player: 0, CardID: "A"},
		SetupActive{Player: 0, CardID: "B"},
		SetupBench{Player: 0, CardID: "A", Slot: BenchLeft},
		SetupBench{Player: 0, CardID: "A", Slot: BenchBehind},
		SetupBench{Player: 0, CardID: "A", Slot: BenchRight},
		SetupBench{Player: 0, CardID: "B", Slot: BenchLeft},
		SetupBench{Player: 0, CardID: "B", Slot: BenchBehind},
		SetupBench{Player: 0, CardID: "B", Slot: BenchRight},
		Concede{Player: 0},
	})
	if err := g.Submit(SetupActive{Player: 1, CardID: "C"}); err != nil {
		t.Fatalf("P2 placing while P1 is named by the prompt: %v", err)
	}
	assertMoves(t, g.LegalMoves(1), []Event{SetupReady{Player: 1}, Concede{Player: 1}})

	// After an Active and a bench card: only the free slots, and Ready.
	if err := g.Submit(SetupActive{Player: 0, CardID: "A"}); err != nil {
		t.Fatal(err)
	}
	if err := g.Submit(SetupBench{Player: 0, CardID: "B", Slot: BenchBehind}); err != nil {
		t.Fatal(err)
	}
	assertMoves(t, g.LegalMoves(0), []Event{SetupReady{Player: 0}, Concede{Player: 0}})
}

func TestLegalMovesGameOver(t *testing.T) {
	g := mainGame(5)
	g.S.Phase = PhaseOver
	assertMoves(t, g.LegalMoves(0), nil)
	assertMoves(t, g.LegalMoves(1), nil)
}

// An illegal move is refused with a reason a player can read, and changes
// nothing.
func TestIllegalMoveGivesReason(t *testing.T) {
	g := mainGame(1)
	pl := &g.S.Players[0]
	pl.Active = newInPlay(legalBasic("A", "Charm", 1), 0)
	pl.Hand = []Card{legalEvo("E", "Charmel", "Charm")}

	cases := []struct {
		ev   Event
		want string
	}{
		{Evolve{Player: 0, HandCardID: "E", Target: 0}, "no Pokémon may evolve on the first turn"},
		{AttachEnergy{Player: 0, Target: 0}, "no energy is available this turn"},
		{UseAttack{Player: 1, Index: 0}, "it is not P2's turn"},
		{Retreat{Player: 0, BenchIndex: 0}, "no Pokémon in that bench slot"},
		{PlayBasic{Player: 0, CardID: "E", Slot: BenchLeft}, "only Basic Pokémon may be played to the bench"},
		{ChooseNewActive{Player: 0, BenchIndex: 0}, "no prompt is pending"},
	}
	for _, c := range cases {
		err := g.Submit(c.ev)
		if err == nil || err.Error() != c.want {
			t.Errorf("Submit(%#v) = %v, want reason %q", c.ev, err, c.want)
		}
	}
	if len(g.Events) != 0 || len(pl.Hand) != 1 {
		t.Error("an illegal move changed the game")
	}
}

// Submit accepts exactly what LegalMoves lists: every listed move succeeds on
// a fresh copy of the state.
func TestEveryListedMoveIsAccepted(t *testing.T) {
	build := func() *Game {
		g := mainGame(3)
		pl := &g.S.Players[0]
		pl.EnergyZone = Fire
		pl.Active = newInPlay(legalBasic("A", "Charm", 1, freeHit, fireBlow), 1)
		pl.Active.Energy[Fire] = 1
		pl.Bench[BenchLeft] = newInPlay(legalBasic("A", "Charm", 1), 1)
		pl.Hand = []Card{legalBasic("C", "Squirt", 1), legalEvo("E", "Charmel", "Charm")}
		return g
	}
	for _, mv := range build().LegalMoves(0) {
		if err := build().Submit(mv); err != nil {
			t.Errorf("listed move %#v refused: %v", mv, err)
		}
	}
}
