package engine

import (
	"reflect"
	"testing"
)

// kindOf is the Action kind each applied event must be recorded as.
func kindOf(ev Event) ActionKind {
	switch ev.(type) {
	case SetupPlace:
		return ActSetup
	case PlayBasic:
		return ActPlayBasic
	case AttachEnergy:
		return ActAttachEnergy
	case Evolve:
		return ActEvolve
	case Retreat:
		return ActRetreat
	case UseAttack:
		return ActAttack
	case EndTurn:
		return ActEndTurn
	case ChooseNewActive:
		return ActPromote
	case Concede:
		return ActConcede
	}
	return ""
}

// F14/F15: every applied move emits exactly one Action, in order, with the
// player, kind and what it did.
func TestEveryAppliedMoveEmitsOneAction(t *testing.T) {
	// Arrange
	d0, d1 := twoDecks()
	g := NewGame(99, d0, d1)

	// Act
	autoPlay(t, g)

	// Assert
	if len(g.Actions) != len(g.Events) {
		t.Fatalf("%d actions for %d events", len(g.Actions), len(g.Events))
	}
	for i, ev := range g.Events {
		a := g.Actions[i]
		if a.Kind != kindOf(ev) {
			t.Errorf("action %d kind = %q, want %q for %T", i, a.Kind, kindOf(ev), ev)
		}
		if len(a.Result) == 0 {
			t.Errorf("action %d (%s) has no result", i, a.Kind)
		}
		if a.Kind == ActAttack && (a.Card.ID == "" || a.Target != (Spot{Player: 1 - a.Player, Pos: 0})) {
			t.Errorf("attack action %d: card %q target %+v", i, a.Card.ID, a.Target)
		}
	}
}

// A refused move leaves no Action behind.
func TestRefusedMoveEmitsNoAction(t *testing.T) {
	d0, d1 := twoDecks()
	g := NewGame(7, d0, d1)
	doSetup(t, g)
	before := len(g.Actions)

	if err := g.Submit(PlayBasic{Player: g.S.Active, CardID: "no-such-card", Slot: 0}); err == nil {
		t.Fatal("expected the move to be refused")
	}

	if len(g.Actions) != before {
		t.Fatalf("refused move added an action: %+v", g.Actions[before:])
	}
}

// Each kind records the card it used and the spot it touched.
func TestActionRecordsCardAndTarget(t *testing.T) {
	// Arrange: both sides set up with an Active only, then past both first turns (T8).
	base := demoEmbor()
	evo := Card{ID: "FIR-02", Name: "Blazor", Stage: Stage1, EvolvesFrom: "Embor", HP: 90, Type: Fire}
	g := NewGame(7, deckOf(base, 20), deckOf(base, 20))
	setupActiveOnly(t, g)
	endTurnAt(t, g, 2)
	p := g.S.Active
	pl := &g.S.Players[p]
	pl.Hand = append(pl.Hand, base, evo)
	pl.Active.Card = base // whichever side is up, its Active is an Embor

	steps := []struct {
		ev     Event
		card   string
		target Spot
	}{
		{PlayBasic{Player: p, CardID: base.ID, Slot: BenchRight}, base.ID, Spot{p, BenchRight + 1}},
		{AttachEnergy{Player: p, Target: 0}, "", Spot{p, 0}},
		{Evolve{Player: p, HandCardID: evo.ID, Target: 0}, evo.ID, Spot{p, 0}},
		{Retreat{Player: p, BenchIndex: BenchRight}, evo.ID, Spot{p, BenchRight + 1}},
		{EndTurn{Player: p}, "", NoSpot},
		{Concede{Player: p}, "", NoSpot},
	}
	for _, s := range steps {
		// Act
		if err := g.Submit(s.ev); err != nil {
			t.Fatalf("%T: %v", s.ev, err)
		}
		a := g.Actions[len(g.Actions)-1]

		// Assert
		if a.Player != p || a.Kind != kindOf(s.ev) {
			t.Errorf("%T: player %d kind %q", s.ev, a.Player, a.Kind)
		}
		if a.Card.ID != s.card || a.Target != s.target {
			t.Errorf("%T: card %q target %+v, want %q %+v", s.ev, a.Card.ID, a.Target, s.card, s.target)
		}
		if len(a.Result) == 0 {
			t.Errorf("%T: no result", s.ev)
		}
	}
}

// Actions are derived, so replaying the log reproduces them exactly.
func TestReplayReproducesActions(t *testing.T) {
	d0, d1 := twoDecks()
	g := NewGame(99, d0, d1)
	autoPlay(t, g)

	r, err := Replay(99, d0, d1, g.Events)
	if err != nil {
		t.Fatalf("replay failed: %v", err)
	}

	if !reflect.DeepEqual(r.Actions, g.Actions) {
		t.Fatal("replayed actions differ from the original")
	}
}
