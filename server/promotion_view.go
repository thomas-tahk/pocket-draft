package main

import (
	"fmt"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// hidePromotion keeps the opponent's new Active secret from viewer while
// viewer still owes their own after a double knockout (spec E6). The
// opponent's side is shown as it was before the pick — Active empty, the
// Pokémon still in its bench slot — and the narration line and history entry
// that name it say only that a new Active was chosen.
func hidePromotion(g *engine.Game, viewer int, v *gameView) {
	h, hidden := g.HiddenPromotion()
	if !hidden || h.Player == viewer {
		return
	}
	pv := &v.Players[h.Player]
	pv.Bench[h.BenchIndex], pv.Active = pv.Active, nil

	named := g.Narration[h.Narration]
	vague := fmt.Sprintf("P%d chose a new Active.", h.Player+1)
	v.Narration = append([]string(nil), v.Narration...)
	v.Narration[h.Narration] = vague

	a := &v.Actions[h.Action]
	a.Result = replaceLine(a.Result, named, vague)
	if a.Kind == string(engine.ActPromote) {
		a.Card, a.Target = nil, spotView{Player: -1, Pos: -1}
	}
}

// replaceLine returns a copy of lines with every from replaced by to.
func replaceLine(lines []string, from, to string) []string {
	out := append([]string(nil), lines...)
	for i, l := range out {
		if l == from {
			out[i] = to
		}
	}
	return out
}
