package main

import (
	"reflect"
	"testing"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// The view lists the viewer's legal moves, and only the viewer's, in the same
// shape /api/move takes, so the screen and bot can post a listed move back
// unchanged.
func TestViewListsLegalMovesAsPostableMoves(t *testing.T) {
	fire, lightning := engine.DemoDecks()
	game = engine.NewGame(1, fire, lightning)

	for step := 0; step < 6; step++ {
		for viewer := 0; viewer < 2; viewer++ {
			v := toGameView(game, viewer)
			want := game.LegalMoves(viewer)
			got := v.Players[viewer].LegalMoves
			if len(got) != len(want) {
				t.Fatalf("viewer %d: view lists %d moves, engine %d", viewer, len(got), len(want))
			}
			for i, mv := range got {
				ev, err := toEvent(mv)
				if err != nil || !reflect.DeepEqual(ev, want[i]) {
					t.Fatalf("viewer %d move %d: %+v round-trips to %#v (%v), want %#v", viewer, i, mv, ev, err, want[i])
				}
			}
			if other := v.Players[1-viewer].LegalMoves; len(other) != 0 {
				t.Fatalf("viewer %d sees the opponent's moves: %+v", viewer, other)
			}
		}

		actor := game.S.Active
		if game.S.Pending != nil {
			actor = game.S.Pending.Player
		}
		listed := toGameView(game, actor).Players[actor].LegalMoves
		// Answer prompts with the first option; in the main phase take End
		// Turn (second to last, before Concede) so the game moves on.
		pick := listed[0]
		if game.S.Pending == nil {
			pick = listed[len(listed)-2]
		}
		if resp := applyMove(pick); !resp.OK {
			t.Fatalf("posting listed move %+v refused: %s", pick, resp.Error)
		}
	}
}
