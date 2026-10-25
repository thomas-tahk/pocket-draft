package engine

import "testing"

func testStadium(id, name string) Card {
	return Card{ID: id, Name: name, Kind: KindStadium, Passive: []Passive{{Hook: HookRetreatCost, Value: 1}}}
}

func playStadium(p int, id string) PlayTrainer {
	return PlayTrainer{Player: p, CardID: id, Target: NoSpot}
}

// A6: a played Stadium is in play for both players, and its text applies.
func TestStadiumInPlay(t *testing.T) {
	g, p := mainPhaseWith(t, testStadium("S1", "Trail"))
	before := g.RetreatCost(1 - p)

	mustSubmit(t, g, playStadium(p, "S1"))

	if g.S.Stadium == nil || g.S.Stadium.ID != "S1" {
		t.Fatalf("Stadium = %v, want S1", g.S.Stadium)
	}
	if len(g.S.Players[p].Discard) != 0 {
		t.Error("a played Stadium went to the discard")
	}
	if got := g.RetreatCost(1 - p); got != before+1 {
		t.Errorf("opponent's retreat cost %d, want %d", got, before+1)
	}
}

// A6 done-gate: a Stadium with the same name as the one in play isn't legal.
func TestSameNameStadiumNotLegal(t *testing.T) {
	g, p := mainPhaseWith(t, testStadium("S1", "Trail"))
	mustSubmit(t, g, playStadium(p, "S1"))
	mustSubmit(t, g, EndTurn{Player: p})
	o := 1 - p
	g.S.Players[o].Hand = append(g.S.Players[o].Hand, testStadium("S2", "Trail"))

	if isLegal(g, o, playStadium(o, "S2")) || g.Submit(playStadium(o, "S2")) == nil {
		t.Error("a same-name Stadium was played")
	}
}

// A6: a new Stadium replaces the old, which goes to its owner's discard.
func TestStadiumReplaces(t *testing.T) {
	g, p := mainPhaseWith(t, testStadium("S1", "Trail"))
	mustSubmit(t, g, playStadium(p, "S1"))
	mustSubmit(t, g, EndTurn{Player: p})
	o := 1 - p
	g.S.Players[o].Hand = append(g.S.Players[o].Hand, testStadium("S2", "Field"))

	mustSubmit(t, g, playStadium(o, "S2"))

	if g.S.Stadium.ID != "S2" {
		t.Errorf("Stadium = %s, want S2", g.S.Stadium.ID)
	}
	if d := g.S.Players[p].Discard; len(d) != 1 || d[0].ID != "S1" {
		t.Errorf("first player's discard = %v, want the replaced Stadium", d)
	}
}

// A6: one Stadium per turn per player.
func TestOneStadiumPerTurn(t *testing.T) {
	g, p := mainPhaseWith(t, testStadium("S1", "Trail"), testStadium("S2", "Field"))
	mustSubmit(t, g, playStadium(p, "S1"))

	if isLegal(g, p, playStadium(p, "S2")) {
		t.Error("a second Stadium in one turn was legal")
	}
	mustSubmit(t, g, EndTurn{Player: p})
	mustSubmit(t, g, EndTurn{Player: 1 - p})
	if !isLegal(g, p, playStadium(p, "S2")) {
		t.Error("Stadium not legal again on the player's next turn")
	}
}

// Field Blower may discard the Stadium in play.
func TestDiscardStadium(t *testing.T) {
	blower := Card{ID: "B1", Name: "Blower", Kind: KindItem, Trainer: []EffectOp{DiscardToolOrStadium{}}}
	g, p := mainPhaseWith(t, testStadium("S1", "Trail"), blower)
	mustSubmit(t, g, playStadium(p, "S1"))

	mustSubmit(t, g, PlayTrainer{Player: p, CardID: "B1", Target: StadiumSpot})

	if g.S.Stadium != nil {
		t.Error("Stadium still in play")
	}
	if n := len(g.S.Players[p].Discard); n != 2 {
		t.Errorf("discard has %d cards, want the Blower and the Stadium", n)
	}
}
