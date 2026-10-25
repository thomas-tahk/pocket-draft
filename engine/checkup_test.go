package engine

import (
	"strings"
	"testing"
)

// lineIndex is the index of the first line containing sub, or -1.
func lineIndex(lines []string, sub string) int {
	for i, l := range lines {
		if strings.Contains(l, sub) {
			return i
		}
	}
	return -1
}

// assertInOrder fails unless each of subs appears in lines, in order.
func assertInOrder(t *testing.T, lines []string, subs ...string) {
	t.Helper()
	last := -1
	for _, s := range subs {
		i := lineIndex(lines, s)
		if i <= last {
			t.Fatalf("want %q after line %d in %q", s, last, lines)
		}
		last = i
	}
}

// C1, C2, C6: a Checkup with Poison, Burn and Sleep resolves them in that
// order: 10, then 20 and a flip for Burn, then a flip for Sleep.
func TestCheckupPoisonBurnAsleepInOrder(t *testing.T) {
	// Arrange
	g, me := conditionGame()
	me.Card.HP = 200
	for _, s := range []Status{Poison, Burn, Asleep} {
		me.Inflict(s)
	}

	// Act
	endTurn(t, g)

	// Assert
	act := g.Actions[len(g.Actions)-1]
	assertInOrder(t, act.Result, "took 10 from poison", "took 20 from burn", "Asleep")
	if me.Damage != 30 {
		t.Errorf("damage %d, want 30", me.Damage)
	}
	if len(act.Flips) != 2 || !act.Flips[0].Waits || !act.Flips[1].Waits {
		t.Fatalf("flips %+v, want Burn's then Sleep's, each its own effect", act.Flips)
	}
	if asleep := act.Flips[1].Heads; me.Has(Asleep) == asleep {
		t.Errorf("sleep flip heads=%v but still asleep=%v; heads cures", asleep, me.Has(Asleep))
	}
}

// C1, C7: with Paralysis in place of Sleep, it wears off last, after Poison and Burn.
func TestCheckupPoisonBurnParalyzedInOrder(t *testing.T) {
	// Arrange: P1's Active was paralyzed before P1's turn began.
	g, me := conditionGame()
	me.Card.HP = 200
	for _, s := range []Status{Poison, Burn, Paralyzed} {
		me.Inflict(s)
	}
	me.ownTurnParalyzed = true

	// Act
	endTurn(t, g)

	// Assert
	act := g.Actions[len(g.Actions)-1]
	assertInOrder(t, act.Result, "took 10 from poison", "took 20 from burn", "no longer Paralyzed")
}

// C6: Sleep cures on heads and stays on tails.
func TestSleepFlipAtCheckup(t *testing.T) {
	for _, heads := range []bool{true, false} {
		g, me := conditionGame()
		me.Inflict(Asleep)
		g.rng = seedFor(heads)

		endTurn(t, g)

		if me.Has(Asleep) == heads {
			t.Errorf("heads=%v: still asleep=%v", heads, me.Has(Asleep))
		}
	}
}

// C1: the Checkup checks both Actives.
func TestCheckupChecksBothActives(t *testing.T) {
	g, me := conditionGame()
	me.Inflict(Poison)
	g.S.Players[1].Active.Inflict(Poison)

	endTurn(t, g)

	if me.Damage != 10 || g.S.Players[1].Active.Damage != 10 {
		t.Errorf("damage %d and %d, want 10 each", me.Damage, g.S.Players[1].Active.Damage)
	}
}
