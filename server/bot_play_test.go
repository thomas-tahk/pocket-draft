package main

import (
	"testing"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// seatPlays is what one seat did in a finished bot-vs-bot game.
type seatPlays struct {
	benched, evolved, trainers bool
}

func playsBySeat() [2]seatPlays {
	var out [2]seatPlays
	for _, ev := range game.Events {
		switch e := ev.(type) {
		case engine.SetupBench:
			out[e.Player].benched = true
		case engine.PlayBasic:
			out[e.Player].benched = true
		case engine.Evolve:
			out[e.Player].evolved = true
		case engine.PlayTrainer:
			out[e.Player].trainers = true
		}
	}
	return out
}

func hasCard(deck []engine.Card, want func(engine.Card) bool) bool {
	for _, c := range deck {
		if want(c) {
			return true
		}
	}
	return false
}

func isEvolution(c engine.Card) bool { return c.EvolvesFrom != "" }

func TestBotPlaysARealGame(t *testing.T) {
	useRealCards(t)
	evolvers, evolved := 0, 0
	for seed := uint64(1); seed <= botSeeds; seed++ {
		botVsBot(t, seed)
		if game.S.Reason == engine.ReasonTurnCap {
			t.Errorf("seed %d: game hit the turn cap on turn %d", seed, game.S.Turn)
		}
		var decks [2][]engine.Card
		decks[0], _ = draftBotDeck(seed)
		decks[1], _ = draftBotDeck(seed + 1_000_000)
		plays := playsBySeat()
		for seat, deck := range decks {
			if !plays[seat].benched {
				t.Errorf("seed %d seat %d: never benched a Pokémon", seed, seat)
			}
			if hasCard(game.S.Players[seat].Hand, engine.Card.IsTrainer) && !plays[seat].trainers {
				t.Errorf("seed %d seat %d: ended holding a Trainer, played none", seed, seat)
			}
			if hasCard(deck, isEvolution) {
				evolvers++
				if plays[seat].evolved {
					evolved++
				}
			}
		}
	}
	if evolved*2 < evolvers {
		t.Errorf("bot evolved in %d of %d games with an evolution, want at least half", evolved, evolvers)
	}
}
