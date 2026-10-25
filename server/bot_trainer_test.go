package main

import (
	"testing"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// trainerBoard is a game past setup whose current player's Active has one
// [R] attack, an [R] Energy Zone, no Energy attached and no damage.
func trainerBoard(t *testing.T) (*engine.Player, int) {
	t.Helper()
	useRealCards(t)
	fire, lightning := engine.DemoDecks()
	game = engine.NewGame(1, fire, lightning)
	for game.S.Phase == engine.PhaseSetup {
		if !botMove() {
			t.Fatal("bot made no setup move")
		}
	}
	p := game.S.Active
	pl := &game.S.Players[p]
	pl.EnergyTypes = []engine.Energy{engine.Fire}
	pl.Active.Card.Attacks = []engine.Attack{{Name: "Hit", Cost: []engine.Energy{engine.Fire}, Damage: 30}}
	pl.Active.Energy = map[engine.Energy]int{}
	return pl, p
}

func rankOf(p int, id string, at engine.Spot) (int, bool) {
	return moveRank(game, engine.PlayTrainer{Player: p, CardID: id, Target: at})
}

func TestBotHealsTheMostDamagedPokemon(t *testing.T) {
	// Arrange
	pl, p := trainerBoard(t)
	pl.Active.Damage = 20
	pl.Bench[0] = &engine.InPlay{Card: pl.Active.Card, Damage: 60, Energy: map[engine.Energy]int{}}

	// Act
	active, okA := rankOf(p, "P-A-001", engine.Spot{Player: p, Pos: 0})
	bench, okB := rankOf(p, "P-A-001", engine.Spot{Player: p, Pos: 1})

	// Assert
	if !okA || !okB || bench <= active {
		t.Errorf("Potion: active %d/%v, bench %d/%v; want both ok, bench higher", active, okA, bench, okB)
	}
}

func TestBotSkipsCardsWithNothingToDo(t *testing.T) {
	cases := []struct {
		name, id string
		set      func(pl *engine.Player)
		want     bool
	}{
		{"X Speed, Active can attack", "P-A-002", func(pl *engine.Player) {}, false},
		{"X Speed, Active stuck", "P-A-002", func(pl *engine.Player) { pl.EnergyTypes = []engine.Energy{engine.Water} }, true},
		{"Giovanni, attack unpaid", "A4b-334", func(pl *engine.Player) {}, false},
		{"Giovanni, attack paid", "A4b-334", func(pl *engine.Player) { pl.Active.Energy[engine.Fire] = 1 }, true},
		{"Iono, big hand", "B2a-089", func(pl *engine.Player) { pl.Hand = make([]engine.Card, 5) }, false},
		{"Iono, small hand", "B2a-089", func(pl *engine.Player) { pl.Hand = make([]engine.Card, 2) }, true},
		{"Professor's Research", "A4b-373", func(pl *engine.Player) { pl.Hand = make([]engine.Card, 5) }, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// Arrange
			pl, p := trainerBoard(t)
			c.set(pl)

			// Act
			_, ok := rankOf(p, c.id, engine.NoSpot)

			// Assert
			if ok != c.want {
				t.Errorf("played = %v, want %v", ok, c.want)
			}
		})
	}
}
