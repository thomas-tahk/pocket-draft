package engine

import "testing"

// F17: picking up a card with nowhere to go flashes one readable reason. The
// engine gives it, so the screen never re-derives a rule.

func reasonOf(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func TestWhyCantPlayGivesOneReasonPerHandCard(t *testing.T) {
	// Arrange
	build := func(turn int) (*Game, *Player) {
		g := mainGame(turn)
		pl := &g.S.Players[0]
		pl.Active = newInPlay(legalBasic("A", "Charm", 1), 1)
		return g, pl
	}
	cases := []struct {
		name  string
		setup func() *Game
		card  string
		want  string
	}{
		{"evolution on the first turn", func() *Game {
			g, pl := build(1)
			pl.Hand = []Card{legalEvo("E", "Charmel", "Charm")}
			return g
		}, "E", "no Pokémon may evolve on the first turn"},
		{"evolution with nothing to evolve from", func() *Game {
			g, pl := build(3)
			pl.Hand = []Card{legalEvo("E", "Ivy", "Bulb")}
			return g
		}, "E", "no Bulb in play to evolve into Ivy"},
		{"evolution onto a Pokémon played this turn", func() *Game {
			g, pl := build(3)
			pl.Bench[BenchLeft] = newInPlay(legalBasic("B", "Bulb", 1), 3)
			pl.Hand = []Card{legalEvo("E", "Ivy", "Bulb")}
			return g
		}, "E", "a Pokémon cannot evolve the turn it was played"},
		{"legal evolution", func() *Game {
			g, pl := build(3)
			pl.Hand = []Card{legalEvo("E", "Charmel", "Charm")}
			return g
		}, "E", ""},
		{"Basic with a full Bench", func() *Game {
			g, pl := build(3)
			for s := range pl.Bench {
				pl.Bench[s] = newInPlay(legalBasic("B", "Bulb", 1), 1)
			}
			pl.Hand = []Card{legalBasic("C", "Squirt", 1)}
			return g
		}, "C", "your Bench is full"},
		{"Basic with room", func() *Game {
			g, pl := build(3)
			pl.Hand = []Card{legalBasic("C", "Squirt", 1)}
			return g
		}, "C", ""},
		{"Basic at setup", func() *Game {
			g, pl := build(0)
			g.S.Phase, g.S.Pending, pl.Active = PhaseSetup, &Prompt{Kind: PromptSetup}, nil
			pl.Hand = []Card{legalBasic("C", "Squirt", 1)}
			return g
		}, "C", ""},
		{"not your turn", func() *Game {
			g, pl := build(3)
			pl.Hand = []Card{legalBasic("C", "Squirt", 1)}
			g.S.Active = 1
			return g
		}, "C", "it is not P1's turn"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// Act
			got := reasonOf(c.setup().WhyCantPlay(0, c.card))

			// Assert
			if got != c.want {
				t.Errorf("WhyCantPlay = %q, want %q", got, c.want)
			}
		})
	}
}

func TestWhyCantAttachGivesTheZoneReason(t *testing.T) {
	// Arrange
	g := mainGame(3)
	pl := &g.S.Players[0]
	pl.EnergyZone = Fire

	// Act, Assert
	if got := reasonOf(g.WhyCantAttach(0)); got != "" {
		t.Errorf("energy available: WhyCantAttach = %q, want none", got)
	}
	pl.EnergyUsed = true
	if got := reasonOf(g.WhyCantAttach(0)); got != "energy was already attached this turn" {
		t.Errorf("energy used: WhyCantAttach = %q", got)
	}
}
