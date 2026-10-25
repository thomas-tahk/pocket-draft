package main

import (
	"testing"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// F11: each action carries its coin flips for the screen; only the first of
// an effect waits for the tap.
func TestActionViewCarriesFlips(t *testing.T) {
	// Arrange
	spinner := engine.Card{ID: "SPIN", Name: "Spinner", Stage: engine.Basic, HP: 200, Type: engine.Grass, Weakness: engine.NoEnergy,
		Attacks: []engine.Attack{{Name: "Triple Spin", Effect: []engine.EffectOp{engine.DamagePerHeads{Coins: 3, Per: 10}}}}}
	g := engine.NewGame(1, deckOfCard(spinner, 20), deckOfCard(spinner, 20))
	doSetup(t, g)
	p := g.S.Active

	// Act
	if err := g.Submit(engine.UseAttack{Player: p, Index: 0}); err != nil {
		t.Fatalf("attack: %v", err)
	}
	v := toGameView(g, 0)

	// Assert
	flips := v.Actions[len(v.Actions)-1].Flips
	if len(flips) != 3 {
		t.Fatalf("got %d flips, want 3", len(flips))
	}
	wantFirst := p == 0 // P2 is the bot, whose flips never wait
	if flips[0].Waits != wantFirst || flips[1].Waits || flips[2].Waits || flips[0].Flipper != p {
		t.Errorf("flips = %+v, want flipper %d and only the first waiting=%v", flips, p, wantFirst)
	}
}
