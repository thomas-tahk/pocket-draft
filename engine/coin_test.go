package engine

import (
	"reflect"
	"testing"
)

// flipAttack is a free attack whose effect flips three coins.
var flipAttack = Attack{Name: "Triple Spin", Damage: 0, Effect: []EffectOp{DamagePerHeads{Coins: 3, Per: 10}}}

func waits(flips []CoinFlip) []bool {
	out := []bool{}
	for _, f := range flips {
		out = append(out, f.Waits)
	}
	return out
}

// F11: an effect with 3 flips records 3 results on its Action, made by the
// attacker; only the first waits for the flipper's tap.
func TestEffectFlipsRecordedOnlyFirstWaits(t *testing.T) {
	// Arrange
	g := mainGame(3)
	g.S.Players[0].Active = newInPlay(legalBasic("ME", "Me", 1, flipAttack), 0)
	g.S.Players[1].Active.Card.HP = 200

	// Act
	if err := g.Submit(UseAttack{Player: 0, Index: 0}); err != nil {
		t.Fatalf("attack: %v", err)
	}

	// Assert
	flips := g.Actions[len(g.Actions)-1].Flips
	if len(flips) != 3 {
		t.Fatalf("got %d flips, want 3", len(flips))
	}
	if got := waits(flips); !reflect.DeepEqual(got, []bool{true, false, false}) {
		t.Errorf("waits = %v, want only the first", got)
	}
	heads := 0
	for _, f := range flips {
		if f.Flipper != 0 {
			t.Errorf("flipper %d, want the attacker", f.Flipper)
		}
		if f.Heads {
			heads++
		}
	}
	if d := g.S.Players[1].Active.Damage; d != 10*heads {
		t.Errorf("damage %d does not match %d recorded heads", d, heads)
	}
}

// F11: against the bot, the bot's flips run on their own.
func TestBotFlipsNeverWait(t *testing.T) {
	g := mainGame(4)
	g.S.Active = 1
	g.S.Players[1].Active = newInPlay(legalBasic("BOT", "Bot", 1, flipAttack), 0)
	g.S.Players[0].Active.Card.HP = 200

	if err := g.Submit(UseAttack{Player: 1, Index: 0}); err != nil {
		t.Fatalf("attack: %v", err)
	}

	if got := waits(g.Actions[len(g.Actions)-1].Flips); !reflect.DeepEqual(got, []bool{false, false, false}) {
		t.Errorf("bot waits = %v, want none", got)
	}
}

// F11: every effect's first flip waits; a move with no flips records none.
func TestEachEffectsFirstFlipWaits(t *testing.T) {
	g := mainGame(3)
	g.S.Players[0].Active = newInPlay(legalBasic("ME", "Me", 1, flipAttack), 0)
	g.S.Players[0].Active.Inflict(Confused)
	g.rng = seedFor(true) // the Confusion flip lands heads, so the attack goes on

	attackFlips := func() []CoinFlip {
		if err := g.Submit(UseAttack{Player: 0, Index: 0}); err != nil {
			t.Fatalf("attack: %v", err)
		}
		return g.Actions[len(g.Actions)-1].Flips
	}

	if got := waits(attackFlips()); !reflect.DeepEqual(got, []bool{true, true, false, false}) {
		t.Errorf("waits = %v, want Confusion's flip and the attack's first", got)
	}
	if err := g.Submit(EndTurn{Player: 1}); err != nil {
		t.Fatalf("end turn: %v", err)
	}
	if f := g.Actions[len(g.Actions)-1].Flips; len(f) != 0 {
		t.Errorf("end turn recorded %v, want no flips", f)
	}
}
