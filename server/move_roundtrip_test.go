package main

import (
	"testing"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// A legal Stadium use reaches the state view as a move request. Before, the
// view panicked on it, which crashed the server once a bot played a Stadium.
func TestUseStadiumRoundTripsThroughTheMoveRequest(t *testing.T) {
	ev := engine.UseStadium{Player: 1, CardID: "A1-001"}

	got, err := toEvent(fromEvent(ev))

	if err != nil || got != ev {
		t.Fatalf("round trip gave %#v, %v; want %#v", got, err, ev)
	}
}
