package engine

import "testing"

// Spec S4: a random player flips; the flipper goes first on heads, second on
// tails. Same seed ⇒ same flip, recorded on State for the screen, and turn 1
// belongs to the player the flip chose.
func TestOpeningFlipDecidesFirstPlayer(t *testing.T) {
	seenFlipper := map[int]bool{}
	seenHeads := map[bool]bool{}
	for seed := uint64(1); seed <= 40; seed++ {
		d0, d1 := twoDecks()
		g := NewGame(seed, d0, d1)
		flip := g.S.OpeningFlip

		want := flip.Flipper
		if !flip.Heads {
			want = 1 - flip.Flipper
		}
		if g.S.FirstPlayer != want {
			t.Fatalf("seed %d: flipper %d heads=%v but first player %d", seed, flip.Flipper, flip.Heads, g.S.FirstPlayer)
		}
		if again := NewGame(seed, d0, d1); again.S.OpeningFlip != flip {
			t.Fatalf("seed %d: flip not deterministic", seed)
		}
		doSetup(t, g)
		if g.S.Active != want || g.S.Turn != 1 {
			t.Fatalf("seed %d: turn %d active %d, want turn 1 active %d", seed, g.S.Turn, g.S.Active, want)
		}
		seenFlipper[flip.Flipper] = true
		seenHeads[flip.Heads] = true
	}
	if len(seenFlipper) != 2 || len(seenHeads) != 2 {
		t.Errorf("40 seeds should cover both flippers and both results: %v %v", seenFlipper, seenHeads)
	}
}
