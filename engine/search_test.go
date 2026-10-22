package engine

import "testing"

func testBall(id string) Card {
	return Card{ID: id, Name: "Ball " + id, Kind: KindItem, Trainer: []EffectOp{SearchDeck{Match: Card.IsBasicPokemon}}}
}

// searchDeck is a deck of Items with three distinct Basics among them.
func searchDeck() []Card {
	deck := deckOf(testItem("X"), 17)
	for _, id := range []string{"B1", "B2", "B3"} {
		b := demoEmbor()
		b.ID = id
		deck = append(deck, b)
	}
	return deck
}

// A7 done-gate: the searched card is random among the matching cards.
func TestSearchIsRandomAmongMatches(t *testing.T) {
	seen := map[string]bool{}
	for seed := 0; seed < 30; seed++ {
		g, p := mainPhaseWith(t, testBall("P1"))
		pl := &g.S.Players[p]
		pl.Deck = searchDeck()
		g.rng = newRng(uint64(seed))
		hand := len(pl.Hand)

		mustSubmit(t, g, PlayTrainer{Player: p, CardID: "P1", Target: NoSpot})

		got := pl.Hand[len(pl.Hand)-1]
		if len(pl.Hand) != hand || !got.IsBasicPokemon() || len(pl.Deck) != 19 {
			t.Fatalf("seed %d: hand %d→%d, took %s, deck %d", seed, hand, len(pl.Hand), got.ID, len(pl.Deck))
		}
		seen[got.ID] = true
	}
	if len(seen) != 3 {
		t.Errorf("searched cards over 30 seeds = %v, want all three Basics", seen)
	}
}

// T3 done-gate: with 10 cards in hand a search can't be played.
func TestSearchBlockedAtTen(t *testing.T) {
	g, p := mainPhaseWith(t, testBall("P1"))
	pl := &g.S.Players[p]
	for len(pl.Hand) < handMax {
		pl.Hand = append(pl.Hand, testItem("X"))
	}

	if isLegal(g, p, PlayTrainer{Player: p, CardID: "P1", Target: NoSpot}) {
		t.Error("search legal with 10 cards in hand")
	}
}

// A search with no match in the deck plays and puts nothing in hand.
func TestSearchWithNoMatch(t *testing.T) {
	g, p := mainPhaseWith(t, testBall("P1"))
	pl := &g.S.Players[p]
	pl.Deck = deckOf(testItem("X"), 5)
	hand := len(pl.Hand)

	mustSubmit(t, g, PlayTrainer{Player: p, CardID: "P1", Target: NoSpot})

	if len(pl.Hand) != hand-1 || len(pl.Deck) != 5 {
		t.Errorf("hand %d→%d, deck %d; want the Ball gone and the deck untouched", hand, len(pl.Hand), len(pl.Deck))
	}
}
