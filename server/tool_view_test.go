package main

import (
	"testing"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// The view shows an attached Tool and the Stadium in play (spec F13).
func TestViewShowsToolAndStadium(t *testing.T) {
	useRealCards(t)
	g, p := trainerGame(t, engine.Card{ID: "T"})
	tool := engine.Card{ID: "TL", Name: "Giant Cape", Kind: engine.KindTool}
	stadium := engine.Card{ID: "ST", Name: "Training Area", Kind: engine.KindStadium}
	g.S.Players[p].Active.Tool = &tool
	g.S.Stadium = &stadium

	v := toGameView(g, p)

	if at := v.Players[p].Active.Tool; at == nil || at.ID != "TL" {
		t.Errorf("active tool = %+v, want TL", at)
	}
	if v.Stadium == nil || v.Stadium.ID != "ST" {
		t.Errorf("stadium = %+v, want ST", v.Stadium)
	}
}
