package main

import (
	"testing"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// C4/C5: the view says which of Asleep, Paralyzed, Confused a Pokémon has.
func TestInPlayViewShowsSpecialCondition(t *testing.T) {
	mon := &engine.InPlay{Card: engine.Card{Name: "X", HP: 60}, Energy: map[engine.Energy]int{}}
	mon.Inflict(engine.Paralyzed)

	v := toInPlayView(mon)

	if !v.Paralyzed || v.Asleep || v.Confused {
		t.Errorf("view asleep=%v paralyzed=%v confused=%v, want only paralyzed", v.Asleep, v.Paralyzed, v.Confused)
	}
}
