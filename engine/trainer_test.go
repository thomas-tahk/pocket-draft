package engine

import "testing"

func testSupporter(id string) Card {
	return Card{ID: id, Name: "Researcher " + id, Kind: KindSupporter, Trainer: []EffectOp{DrawCards{N: 2}}}
}

func testItem(id string) Card {
	return Card{ID: id, Name: "Gadget " + id, Kind: KindItem, Trainer: []EffectOp{DrawCards{N: 1}}}
}

// mainPhaseWith returns a game past setup whose player on turn holds extra.
func mainPhaseWith(t *testing.T, extra ...Card) (*Game, int) {
	t.Helper()
	d0, d1 := twoDecks()
	g := NewGame(42, d0, d1)
	doSetup(t, g)
	p := g.S.Active
	g.S.Players[p].Hand = append(g.S.Players[p].Hand, extra...)
	return g, p
}

func isLegal(g *Game, p int, ev Event) bool {
	for _, m := range g.LegalMoves(p) {
		if pt, ok := m.(PlayTrainer); ok && pt == ev {
			return true
		}
	}
	return false
}

// A1: one Supporter per turn, from turn 1; the played card goes to the discard.
func TestSecondSupporterNotLegal(t *testing.T) {
	g, p := mainPhaseWith(t, testSupporter("S1"), testSupporter("S2"))
	first := PlayTrainer{Player: p, CardID: "S1", Target: NoSpot}

	if !isLegal(g, p, first) {
		t.Fatalf("first Supporter not legal on turn %d", g.S.Turn)
	}
	if err := g.Submit(first); err != nil {
		t.Fatalf("play first Supporter: %v", err)
	}

	second := PlayTrainer{Player: p, CardID: "S2", Target: NoSpot}
	if isLegal(g, p, second) || g.Submit(second) == nil {
		t.Error("second Supporter in one turn was allowed")
	}
	if d := g.S.Players[p].Discard; len(d) != 1 || d[0].ID != "S1" {
		t.Errorf("discard = %v, want the played Supporter", d)
	}
}

// A1: the Supporter limit resets on the player's next turn.
func TestSupporterAllowedAgainNextTurn(t *testing.T) {
	g, p := mainPhaseWith(t, testSupporter("S1"), testSupporter("S2"))
	mustSubmit(t, g, PlayTrainer{Player: p, CardID: "S1", Target: NoSpot})
	mustSubmit(t, g, EndTurn{Player: p})
	mustSubmit(t, g, EndTurn{Player: 1 - p})

	if !isLegal(g, p, PlayTrainer{Player: p, CardID: "S2", Target: NoSpot}) {
		t.Error("Supporter not legal on the player's next turn")
	}
}

// A2, A3: Items are unlimited, and each goes to the discard pile.
func TestItemsUnlimited(t *testing.T) {
	g, p := mainPhaseWith(t, testItem("I1"), testItem("I2"), testSupporter("S1"))
	mustSubmit(t, g, PlayTrainer{Player: p, CardID: "S1", Target: NoSpot})
	mustSubmit(t, g, PlayTrainer{Player: p, CardID: "I1", Target: NoSpot})
	mustSubmit(t, g, PlayTrainer{Player: p, CardID: "I2", Target: NoSpot})

	if n := len(g.S.Players[p].Discard); n != 3 {
		t.Errorf("discard has %d cards, want 3", n)
	}
}

// A Trainer whose text isn't built has no move, and says why.
func TestUnbuiltTrainerNotPlayable(t *testing.T) {
	g, p := mainPhaseWith(t, Card{ID: "U1", Name: "Unbuilt", Kind: KindItem})

	if isLegal(g, p, PlayTrainer{Player: p, CardID: "U1", Target: NoSpot}) {
		t.Error("a Trainer with no built text was legal")
	}
	if g.WhyCantPlay(p, "U1") == nil {
		t.Error("WhyCantPlay gave no reason for an unbuilt Trainer")
	}
}

// A Trainer is not a Basic Pokémon: it can't be placed at setup or benched.
func TestTrainerIsNotABasic(t *testing.T) {
	g, p := mainPhaseWith(t, testItem("I1"))
	for _, m := range g.LegalMoves(p) {
		if pb, ok := m.(PlayBasic); ok && pb.CardID == "I1" {
			t.Fatal("a Trainer was offered as a Basic for the Bench")
		}
	}
}

func mustSubmit(t *testing.T, g *Game, ev Event) {
	t.Helper()
	if err := g.Submit(ev); err != nil {
		t.Fatalf("submit %#v: %v", ev, err)
	}
}
