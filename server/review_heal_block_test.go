package main

import (
	"testing"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// Claydol's Heal Block stops Tool, Stadium and Trainer healing too.
func TestHealBlockStopsLeftovers(t *testing.T) {
	g, _, a, _ := duel(t, cubone, volbeat)
	tool(t, a, "A3b-067")
	onBoard(t, g, 1-g.S.Active, 1, "A3a-031") // Claydol
	a.Damage = 30
	endTurn(t, g)
	if a.Damage != 30 {
		t.Errorf("Leftovers healed through Heal Block: %d", a.Damage)
	}
}

func TestHealBlockStopsSitrusBerry(t *testing.T) {
	g, _, a, _ := duel(t, volbeat, cubone)
	tool(t, a, "B1-218")
	onBoard(t, g, 1-g.S.Active, 1, "A3a-031")
	a.Damage = 40
	endTurn(t, g)
	if a.Damage != 40 {
		t.Errorf("Sitrus Berry healed through Heal Block: %d", a.Damage)
	}
}

func TestHealBlockStopsSoothingShore(t *testing.T) {
	g, p, a, _ := duel(t, "B4a-013", cubone)
	stadiumIn(t, g, "B4-154")
	a.Energy[engine.Water] = 1
	a.Damage = 30
	onBoard(t, g, 1-p, 1, "A3a-031")
	endTurn(t, g)
	if a.Damage != 30 {
		t.Errorf("Soothing Shore healed through Heal Block: %d", a.Damage)
	}
}

func TestHealBlockStopsPokemonCenterLady(t *testing.T) {
	useRealCards(t)
	g, p := trainerGame(t, realCard(t, "A2b-070"))
	mon := g.S.Players[p].Active
	mon.Damage = 40
	onBoard(t, g, 1-p, 1, "A3a-031")

	_ = g.Submit(engine.PlayTrainer{Player: p, CardID: "A2b-070", Target: engine.Spot{Player: p, Pos: 0}})

	if mon.Damage != 40 {
		t.Errorf("Pokémon Center Lady healed through Heal Block: %d", mon.Damage)
	}
}
