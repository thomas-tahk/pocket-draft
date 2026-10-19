package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// checkupDoubleKO sets up a game where both Actives are knocked out by poison
// at the Checkup, each with two benched Pokémon (slots 0 and 2), and ends the
// turn. The player whose turn is next owes the first new Active.
func checkupDoubleKO(t *testing.T) *engine.Game {
	t.Helper()
	g := engine.NewGame(7, viewTestDeck("mine"), viewTestDeck("theirs"))
	doSetup(t, g)
	for p := 0; p < 2; p++ {
		pl := &g.S.Players[p]
		pl.Active.Poisoned = true
		pl.Active.Damage = pl.Active.Card.HP - 10
		for _, slot := range []int{0, 2} {
			c := engine.Card{ID: "bench-" + seatLetter(p) + string(rune('0'+slot)), Name: "Hidey" + seatLetter(p), Stage: engine.Basic, HP: 60, Type: engine.Fire}
			pl.Bench[slot] = &engine.InPlay{Card: c, Energy: map[engine.Energy]int{}}
		}
	}
	if err := g.Submit(engine.EndTurn{Player: g.S.Active}); err != nil {
		t.Fatal(err)
	}
	return g
}

func seatLetter(p int) string { return string(rune('A' + p)) }

// E6: while the second player is still choosing, the first player's pick is
// hidden from them on the board, in the narration and in the action history;
// once both have placed it shows.
func TestViewHidesFirstPromotionUntilSecondPlaced(t *testing.T) {
	// Arrange
	g := checkupDoubleKO(t)
	first := 1 - g.S.Active
	picked := g.S.Players[first].Bench[2].Card.ID
	if err := g.Submit(engine.ChooseNewActive{Player: first, BenchIndex: 2}); err != nil {
		t.Fatal(err)
	}

	// Act
	hidden := toGameView(g, 1-first)
	own := toGameView(g, first)
	if err := g.Submit(engine.ChooseNewActive{Player: 1 - first, BenchIndex: 0}); err != nil {
		t.Fatal(err)
	}
	shown := toGameView(g, 1-first)

	// Assert
	opp := hidden.Players[first]
	if opp.Active != nil || opp.Bench[2] == nil || opp.Bench[2].Card.ID != picked {
		t.Errorf("board gives the pick away: active=%+v bench[2]=%+v", opp.Active, opp.Bench[2])
	}
	last := hidden.Actions[len(hidden.Actions)-1]
	if last.Card != nil || last.Target.Pos != -1 {
		t.Errorf("history gives the pick away: %+v", last)
	}
	raw, _ := json.Marshal(struct {
		N []string
		A []actionView
	}{hidden.Narration, hidden.Actions})
	if strings.Contains(string(raw), "Hidey"+seatLetter(first)) || strings.Contains(string(raw), picked) {
		t.Errorf("narration or history names the pick: %s", raw)
	}
	if own.Players[first].Active == nil || own.Players[first].Active.Card.ID != picked {
		t.Errorf("the picker's own view should show its pick: %+v", own.Players[first].Active)
	}
	if a := shown.Players[first].Active; a == nil || a.Card.ID != picked || shown.Players[first].Bench[2] != nil {
		t.Errorf("after both placed the pick should show: %+v", shown.Players[first])
	}
	if !strings.Contains(strings.Join(shown.Narration, "\n"), "promoted Hidey"+seatLetter(first)) {
		t.Errorf("after both placed the narration should name the pick: %v", shown.Narration)
	}
}
