package main

import (
	"testing"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// koBench has the attacker's Cubone knock out a Benched Volbeat that has no HP
// left, holding Tool benchTool, while the Active holds activeTool ("" for none).
func koBench(t *testing.T, benchTool, activeTool string) (*engine.Game, int) {
	t.Helper()
	g, p, _, d := duel(t, cubone, pikachu)
	b := onBoard(t, g, 1-p, 1, volbeat)
	b.Damage = b.Card.HP
	if benchTool != "" {
		tool(t, b, benchTool)
	}
	if activeTool != "" {
		tool(t, d, activeTool)
	}
	useAttack(t, g, p, 0)
	return g, p
}

func handHasName(pl engine.Player, name string) bool {
	for _, c := range pl.Hand {
		if c.Name == name {
			return true
		}
	}
	return false
}

// A Benched Pokémon's own Rescue Scarf fires when an attack knocks it out.
func TestBenchedRescueScarfFires(t *testing.T) {
	g, p := koBench(t, "A4-155", "")
	if !handHasName(g.S.Players[1-p], "Volbeat") {
		t.Error("Benched Volbeat with Rescue Scarf did not return to hand")
	}
}

// The Active's Rescue Scarf does not save a Benched Pokémon.
func TestActiveRescueScarfDoesNotSaveBench(t *testing.T) {
	g, p := koBench(t, "", "A4-155")
	if handHasName(g.S.Players[1-p], "Volbeat") {
		t.Error("Volbeat went to hand on the Active's Rescue Scarf")
	}
}
