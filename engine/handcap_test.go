package engine

import "testing"

// T3: at the start of a turn with 10 cards in hand, no card is drawn and it
// stays on top of the deck.
func TestHandAt10BlocksTurnDraw(t *testing.T) {
	d0, d1 := twoDecks()
	g := NewGame(42, d0, d1)
	doSetup(t, g)

	next := &g.S.Players[1-g.S.Active]
	for len(next.Hand) < 10 {
		next.Hand = append(next.Hand, next.Deck[0])
		next.Deck = next.Deck[1:]
	}
	top, deckLen := next.Deck[0], len(next.Deck)

	if err := g.Submit(EndTurn{Player: g.S.Active}); err != nil {
		t.Fatalf("end turn: %v", err)
	}

	if len(next.Hand) != 10 {
		t.Errorf("hand = %d cards, want 10", len(next.Hand))
	}
	if len(next.Deck) != deckLen || next.Deck[0].ID != top.ID {
		t.Errorf("deck = %d cards, top %q; want %d, top %q", len(next.Deck), next.Deck[0].ID, deckLen, top.ID)
	}
}

// T3: a draw effect (Munchlax's Hungrily Draw) at 10 cards in hand draws
// nothing; the card stays on top of the deck.
func TestHandAt10BlocksEffectDraw(t *testing.T) {
	g := newEffectTestGame(1)
	pl := &g.S.Players[0]
	pl.Hand = make([]Card, 10)
	pl.Deck = []Card{{ID: "X"}, {ID: "Y"}}

	DrawCards{N: 1}.Apply(&EffectContext{g: g, attacker: 0})

	if len(pl.Hand) != 10 {
		t.Errorf("hand = %d cards, want 10", len(pl.Hand))
	}
	if len(pl.Deck) != 2 || pl.Deck[0].ID != "X" {
		t.Errorf("deck = %d cards, want 2 with X on top", len(pl.Deck))
	}
}
