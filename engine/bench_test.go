package engine

import "testing"

// setupActiveOnly answers both setup prompts with an Active and an empty bench.
func setupActiveOnly(t *testing.T, g *Game) {
	t.Helper()
	for g.S.Pending != nil && g.S.Pending.Kind == PromptSetup {
		p := g.S.Pending.Player
		id, _ := firstBasicID(g.S.Players[p].Hand)
		if err := g.Submit(SetupPlace{Player: p, ActiveCardID: id}); err != nil {
			t.Fatalf("setup failed: %v", err)
		}
	}
}

// T7: placing a Basic into a slot keeps that slot; a full slot is not a legal target.
func TestBenchPlacementKeepsSlot(t *testing.T) {
	d0, d1 := twoDecks()
	g := NewGame(7, d0, d1)
	setupActiveOnly(t, g)
	p := g.S.Active

	id, _ := firstBasicID(g.S.Players[p].Hand)
	if err := g.Submit(PlayBasic{Player: p, CardID: id, Slot: BenchRight}); err != nil {
		t.Fatalf("placing into an empty slot should be legal: %v", err)
	}
	bench := g.S.Players[p].Bench
	if bench[BenchRight] == nil || bench[BenchRight].Card.ID != id {
		t.Fatalf("right slot should hold %s, got %+v", id, bench[BenchRight])
	}
	if bench[BenchLeft] != nil || bench[BenchBehind] != nil {
		t.Fatal("the other slots should still be empty")
	}

	id2, _ := firstBasicID(g.S.Players[p].Hand)
	if err := g.Submit(PlayBasic{Player: p, CardID: id2, Slot: BenchRight}); err == nil {
		t.Error("placing into a full slot should be refused")
	}
	if err := g.Submit(PlayBasic{Player: p, CardID: id2, Slot: 3}); err == nil {
		t.Error("placing into a slot that does not exist should be refused")
	}
	if err := g.Submit(PlayBasic{Player: p, CardID: id2, Slot: BenchLeft}); err != nil {
		t.Fatalf("placing into the empty left slot should be legal: %v", err)
	}
	if got := g.S.Players[p].Bench[BenchRight].Card.ID; got != id {
		t.Errorf("right slot changed to %s after placing on the left", got)
	}
}

// Setup places bench cards positionally: index i of BenchCardIDs is slot i, "" leaves it empty.
func TestSetupBenchKeepsSlots(t *testing.T) {
	d0, d1 := twoDecks()
	g := NewGame(7, d0, d1)
	p := g.S.Pending.Player
	hand := g.S.Players[p].Hand
	activeID, benchID := hand[0].ID, hand[1].ID
	if err := g.Submit(SetupPlace{Player: p, ActiveCardID: activeID, BenchCardIDs: []string{"", "", benchID}}); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	bench := g.S.Players[p].Bench
	if bench[BenchRight] == nil || bench[BenchRight].Card.ID != benchID {
		t.Fatalf("right slot should hold %s", benchID)
	}
	if bench[BenchLeft] != nil || bench[BenchBehind] != nil {
		t.Fatal("left and behind should be empty")
	}
}

// Promoting from a slot empties that slot and leaves the others where they were.
func TestPromoteLeavesOtherSlotsInPlace(t *testing.T) {
	d0, d1 := twoDecks()
	g := NewGame(7, d0, d1)
	pl := &g.S.Players[0]
	g.S.Phase = PhaseMain
	pl.Active = nil
	pl.Bench[BenchLeft] = newInPlay(pl.Deck[0], 0)
	rightID := pl.Deck[1].ID
	pl.Bench[BenchRight] = newInPlay(pl.Deck[1], 0)
	g.S.Pending = &Prompt{Player: 0, Kind: PromptNewActive}

	if err := g.Submit(ChooseNewActive{Player: 0, BenchIndex: BenchBehind}); err == nil {
		t.Error("promoting from an empty slot should be refused")
	}
	if err := g.Submit(ChooseNewActive{Player: 0, BenchIndex: BenchLeft}); err != nil {
		t.Fatalf("promote: %v", err)
	}
	if pl.Bench[BenchLeft] != nil {
		t.Error("left slot should be empty after promoting from it")
	}
	if pl.Bench[BenchRight] == nil || pl.Bench[BenchRight].Card.ID != rightID {
		t.Error("right slot should be untouched")
	}
}
