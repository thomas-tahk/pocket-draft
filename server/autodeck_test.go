package main

import (
	"encoding/json"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// names lists the distinct card names in a player's hand and deck.
func names(p engine.Player) map[string]bool {
	seen := map[string]bool{}
	for _, c := range append(append([]engine.Card{}, p.Hand...), p.Deck...) {
		seen[c.Name] = true
	}
	return seen
}

// A game started without your deck deals you a drafted deck, not the old
// three-card fire stand-in, and it differs by seed.
func TestStartGameWithoutDeckDealsADraftedDeck(t *testing.T) {
	useRealCards(t)
	fire := map[string]bool{}
	for _, c := range fireDeck {
		fire[c.Name] = true
	}

	decks := map[string]bool{}
	for _, seed := range []uint64{1, 2, 3} {
		s := seed
		if _, err := startGame(newReq{Seed: &s}); err != nil {
			t.Fatalf("seed %d: %v", seed, err)
		}
		got := names(game.S.Players[0])
		if len(got) <= len(fire) {
			t.Errorf("seed %d: only %d distinct names %v", seed, len(got), got)
		}
		key := []string{}
		for n := range got {
			key = append(key, n)
		}
		sort.Strings(key)
		decks[strings.Join(key, ",")] = true
	}
	if len(decks) < 2 {
		t.Errorf("three seeds dealt the same deck")
	}
}

// POST /api/autodeck returns a legal drafted deck that /api/new accepts.
func TestAutoDeckIsPlayable(t *testing.T) {
	useRealCards(t)
	rec := httptest.NewRecorder()
	handleAutoDeck(rec, httptest.NewRequest("POST", "/api/autodeck", strings.NewReader(`{"seed":5}`)))
	var got autoDeck
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if len(got.Deck) != draftDeckSize || len(got.Picks) != totalDraftPacks {
		t.Fatalf("got %d deck cards, %d picks", len(got.Deck), len(got.Picks))
	}
	if _, err := startGame(newReq{You: got.Deck}); err != nil {
		t.Fatalf("auto deck rejected: %v", err)
	}
}
