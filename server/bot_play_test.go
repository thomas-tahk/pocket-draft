package main

import (
	"strings"
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

// isPlayableTrainer is a Trainer the bot could play from hand: not a Fossil,
// which is only ever played as a 40-HP Basic.
func isPlayableTrainer(c engine.Card) bool {
	return c.IsTrainer() && !strings.HasPrefix(rawByID[c.ID].TrainerText, "Play this card as if it were a 40-HP Basic")
}

func isEvolution(c engine.Card) bool { return c.EvolvesFrom != "" }

func TestBotPlaysARealGame(t *testing.T) {
	useRealCards(t)
	evolvers, evolved, seats, benchedSeats := 0, 0, 0, 0
	holders, holdersPlayed := 0, 0
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
			seats++
			if plays[seat].benched {
				benchedSeats++
			}
			if hasCard(deck, isPlayableTrainer) {
				holders++
				if plays[seat].trainers {
					holdersPlayed++
				}
			}
			if hasCard(deck, isEvolution) {
				evolvers++
				if plays[seat].evolved {
					evolved++
				}
			}
		}
	}
	t.Logf("seats %d, benched %d; Trainer holders %d, played %d; evolvers %d, evolved %d", seats, benchedSeats, holders, holdersPlayed, evolvers, evolved)
	if benchedSeats != seats {
		t.Errorf("bot benched a Pokémon in %d of %d seats, want every seat", benchedSeats, seats)
	}
	// Most seats hold a non-Fossil Trainer now (#110); a clear majority of
	// them should play one. The count guard stops the check passing with few
	// Trainer decks.
	if holders*2 < seats || holdersPlayed*10 < holders*7 {
		t.Errorf("bot played a Trainer in %d of %d seats whose deck had one (of %d seats), want 70%% of at least half the seats", holdersPlayed, holders, seats)
	}
	if evolved*2 < evolvers {
		t.Errorf("bot evolved in %d of %d games with an evolution, want at least half", evolved, evolvers)
	}
}
