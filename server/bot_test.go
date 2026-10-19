package main

import (
	"reflect"
	"testing"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// botGameSteps bounds one bot-vs-bot game; the engine's turn cap ends a game
// far sooner, so reaching it means the bot stalled.
const botGameSteps = 5000

// 200 seeded bot-vs-bot games all end with a result: every bot move is
// accepted, every prompt (setup and promotion included) is answered, and no
// game ends by the bot conceding.
func TestBotFinishesSeededGames(t *testing.T) {
	useRealCards(t)
	for seed := uint64(1); seed <= 200; seed++ {
		startBotGame(t, seed)
		playBotGame(t, seed)
	}
}

func playBotGame(t *testing.T, seed uint64) {
	t.Helper()
	for step := 0; step < botGameSteps; step++ {
		if game.S.Phase == engine.PhaseOver {
			if game.S.Reason == engine.ReasonConcede {
				t.Errorf("seed %d: the bot conceded", seed)
			}
			return
		}
		ev, ok := botChoice(game)
		if !ok {
			t.Errorf("seed %d: bot found no move at step %d (pending %+v)", seed, step, game.S.Pending)
			return
		}
		if !listed(ev) {
			t.Errorf("seed %d: bot chose %#v, which is not in LegalMoves", seed, ev)
			return
		}
		if err := game.Submit(ev); err != nil {
			t.Errorf("seed %d: engine refused bot move %#v: %v", seed, ev, err)
			return
		}
	}
	t.Errorf("seed %d: game stalled after %d bot moves", seed, botGameSteps)
}

// startBotGame alternates the demo matchup with two drafted decks, so evolution
// lines and mixed types reach the bot too.
func startBotGame(t *testing.T, seed uint64) {
	t.Helper()
	opponent, err := draftBotDeck(seed)
	if err != nil {
		t.Fatalf("seed %d: %v", seed, err)
	}
	you := fireDeck
	if seed%2 == 0 {
		if you, err = draftBotDeck(seed + 1_000_000); err != nil {
			t.Fatalf("seed %d: %v", seed, err)
		}
	}
	game = engine.NewGame(seed, you, opponent)
}

// listed reports whether ev is one of its player's LegalMoves and not Concede.
func listed(ev engine.Event) bool {
	if _, quit := ev.(engine.Concede); quit {
		return false
	}
	for p := 0; p < 2; p++ {
		for _, m := range game.LegalMoves(p) {
			if reflect.DeepEqual(m, ev) {
				return true
			}
		}
	}
	return false
}

// useRealCards loads the scraped card data and the demo deck, as the server
// does at startup, and restores the globals afterwards.
func useRealCards(t *testing.T) {
	t.Helper()
	savedRaw, savedImages, savedDeck, savedGame := rawByID, cardImages, fireDeck, game
	t.Cleanup(func() { rawByID, cardImages, fireDeck, game = savedRaw, savedImages, savedDeck, savedGame })
	rawByID, cardImages = map[string]rawCard{}, map[string]string{}
	if err := loadCards(cardsPath); err != nil {
		t.Fatal(err)
	}
	deck, err := fireDeckPreset()
	if err != nil {
		t.Fatal(err)
	}
	fireDeck = deck
}
