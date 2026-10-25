package main

import (
	"testing"

	"github.com/thomas-tahk/pocket-draft/engine"
)

const botSeeds = 50

// botVsBot plays a finished game between two drafted bot decks, one per seat.
func botVsBot(t *testing.T, seed uint64) {
	t.Helper()
	a, err := draftBotDeck(seed)
	if err != nil {
		t.Fatalf("seed %d: %v", seed, err)
	}
	b, err := draftBotDeck(seed + 1_000_000)
	if err != nil {
		t.Fatalf("seed %d: %v", seed, err)
	}
	game = engine.NewGame(seed, a, b)
	playBotGame(t, seed)
}

func attackTypes(deck []engine.Card) map[engine.Energy]bool {
	types := map[engine.Energy]bool{}
	for _, c := range deck {
		for _, a := range c.Attacks {
			for _, e := range a.Cost {
				if e != engine.Colorless {
					types[e] = true
				}
			}
		}
	}
	return types
}

// A finished deck holds no Trainer its drafter did not pick (#107: no free
// universals), for the bot and for the auto-draft deck alike.
func TestBotDeckHasNoUnpickedTrainers(t *testing.T) {
	useRealCards(t)
	for seed := uint64(1); seed <= botSeeds; seed++ {
		picks, deck, err := draftBotIDs(seed)
		if err != nil {
			t.Fatalf("seed %d: %v", seed, err)
		}
		picked := map[string]bool{}
		for _, id := range picks {
			picked[rawByID[id].Name] = true
		}
		for _, id := range deck {
			rc := rawByID[id]
			if rc.CardType == "Trainer" && !picked[rc.Name] {
				t.Errorf("seed %d: deck has unpicked Trainer %s", seed, rc.Name)
			}
		}
	}
}

func TestBotDeckIsPlayable(t *testing.T) {
	useRealCards(t)
	for seed := uint64(1); seed <= botSeeds; seed++ {
		deck, err := draftBotDeck(seed)
		if err != nil {
			t.Fatalf("seed %d: %v", seed, err)
		}
		names := map[string]bool{}
		for _, c := range deck {
			names[c.Name] = true
		}
		for _, c := range deck {
			if c.EvolvesFrom != "" && !names[c.EvolvesFrom] {
				t.Errorf("seed %d: %s has no %s in the deck", seed, c.Name, c.EvolvesFrom)
			}
		}
		if n := len(attackTypes(deck)); n > 2 {
			t.Errorf("seed %d: deck attacks need %d energy types, want at most 2", seed, n)
		}
	}
}

func TestBotAttacksInNinetyPercentOfGames(t *testing.T) {
	useRealCards(t)
	attacked, seats := 0, 0
	for seed := uint64(1); seed <= botSeeds; seed++ {
		botVsBot(t, seed)
		hit := [2]bool{}
		for _, ev := range game.Events {
			if a, ok := ev.(engine.UseAttack); ok {
				hit[a.Player] = true
			}
		}
		for _, h := range hit {
			seats++
			if h {
				attacked++
			}
		}
	}
	if attacked*10 < seats*9 {
		t.Errorf("bot attacked in %d of %d games, want at least 90%%", attacked, seats)
	}
}
