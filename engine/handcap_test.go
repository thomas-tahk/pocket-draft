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
