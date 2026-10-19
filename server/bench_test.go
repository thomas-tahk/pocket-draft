package main

import (
	"testing"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// A PlayBasic move carries its bench slot through to the engine.
func TestPlayBasicCarriesSlot(t *testing.T) {
	ev, err := toEvent(moveReq{Type: "PlayBasic", Player: 1, CardID: "x", Slot: engine.BenchRight})
	if err != nil {
		t.Fatal(err)
	}
	if got := ev.(engine.PlayBasic).Slot; got != engine.BenchRight {
		t.Errorf("slot = %d, want %d", got, engine.BenchRight)
	}
}

// The view lists all three bench slots in order, with null for an empty one.
func TestBenchViewKeepsEmptySlots(t *testing.T) {
	var pl engine.Player
	pl.Bench[engine.BenchRight] = &engine.InPlay{Card: engine.Card{Name: "Pikachu", HP: 60}}

	bench := toPlayerView(pl, true).Bench
	if len(bench) != engine.BenchSlots {
		t.Fatalf("bench has %d entries, want %d", len(bench), engine.BenchSlots)
	}
	if bench[engine.BenchLeft] != nil || bench[engine.BenchBehind] != nil {
		t.Error("empty slots should be null")
	}
	if bench[engine.BenchRight] == nil || bench[engine.BenchRight].Card.Name != "Pikachu" {
		t.Error("right slot should hold Pikachu")
	}
}
