package engine

import "testing"

// conditionGame is mainGame(3) with P1's Active holding a free attack and a
// Pokémon behind it on the bench, P1 to act.
func conditionGame() (*Game, *InPlay) {
	g := mainGame(3)
	pl := &g.S.Players[0]
	pl.Active = newInPlay(legalBasic("ME", "Me", 0, freeHit), 0)
	pl.Bench[BenchBehind] = newInPlay(legalBasic("B", "Bencher", 1), 0)
	pl.Hand = []Card{legalBasic("H", "Handy", 1)}
	pl.EnergyZone = Fire
	return g, pl.Active
}

// seedFor returns a game rng whose first flip lands heads (true) or tails.
func seedFor(heads bool) *rng {
	for s := uint64(1); ; s++ {
		if (newRng(s).intn(2) == 0) == heads {
			return newRng(s)
		}
	}
}

func hasMove[E Event](moves []Event) bool {
	for _, m := range moves {
		if _, ok := m.(E); ok {
			return true
		}
	}
	return false
}

// C5: Asleep or Paralyzed can't attack or retreat, but cards and energy still play.
func TestAsleepAndParalyzedBlockAttackAndRetreat(t *testing.T) {
	for _, s := range []Status{Asleep, Paralyzed} {
		// Arrange
		g, me := conditionGame()
		me.Inflict(s)

		// Act
		moves := g.LegalMoves(0)

		// Assert
		if hasMove[UseAttack](moves) || hasMove[Retreat](moves) {
			t.Errorf("%v: attack or retreat still listed: %v", s, moves)
		}
		if !hasMove[AttachEnergy](moves) || !hasMove[PlayBasic](moves) {
			t.Errorf("%v: energy and Basics should still play: %v", s, moves)
		}
		if err := g.Submit(UseAttack{Player: 0, Index: 0}); err == nil {
			t.Errorf("%v: attack accepted", s)
		}
	}
}

// C8: a Confused Pokémon may attack and retreat.
func TestConfusedMayAttackAndRetreat(t *testing.T) {
	g, me := conditionGame()
	me.Inflict(Confused)

	moves := g.LegalMoves(0)

	if !hasMove[UseAttack](moves) || !hasMove[Retreat](moves) {
		t.Errorf("Confused should still attack and retreat: %v", moves)
	}
}

// C8: tails, the attack fails with no self-damage, the condition stays and the turn ends.
func TestConfusedTailsAttackFails(t *testing.T) {
	// Arrange
	g, me := conditionGame()
	me.Inflict(Confused)
	g.rng = seedFor(false)

	// Act
	if err := g.Submit(UseAttack{Player: 0, Index: 0}); err != nil {
		t.Fatalf("attack: %v", err)
	}

	// Assert
	if d := g.S.Players[1].Active.Damage; d != 0 {
		t.Errorf("failed attack dealt %d", d)
	}
	if me.Damage != 0 || !me.Has(Confused) {
		t.Errorf("self-damage %d, still confused %v; want 0, true", me.Damage, me.Has(Confused))
	}
	if g.S.Active != 1 {
		t.Error("a failed attack still ends the turn")
	}
}

// C8: heads, the attack goes ahead.
func TestConfusedHeadsAttackHits(t *testing.T) {
	g, me := conditionGame()
	me.Inflict(Confused)
	g.rng = seedFor(true)

	if err := g.Submit(UseAttack{Player: 0, Index: 0}); err != nil {
		t.Fatalf("attack: %v", err)
	}

	if d := g.S.Players[1].Active.Damage; d != freeHit.Damage {
		t.Errorf("damage %d, want %d", d, freeHit.Damage)
	}
}

// C3, C4: Poison and Burn stack with one of Asleep/Paralyzed/Confused; a new
// one of those replaces the old, so a fourth never stacks.
func TestConditionsStackToThreeAtMost(t *testing.T) {
	// Arrange
	_, me := conditionGame()

	// Act
	for _, s := range []Status{Poison, Burn, Asleep, Paralyzed, Confused} {
		me.Inflict(s)
	}

	// Assert
	got := me.Conditions()
	want := []Status{Poison, Burn, Confused}
	if len(got) != len(want) {
		t.Fatalf("conditions %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("conditions %v, want %v", got, want)
		}
	}
}

// C7: Paralysis wears off at the Checkup after its owner's next turn.
func TestParalysisWearsOffAfterOwnersNextTurn(t *testing.T) {
	// Arrange: P1 paralyzes P2's Active on P1's turn.
	g, _ := conditionGame()
	opp := g.S.Players[1].Active
	opp.Inflict(Paralyzed)

	// Act + Assert
	endTurn(t, g) // P1's Checkup: still paralyzed
	if !opp.Has(Paralyzed) {
		t.Fatal("Paralysis cleared before its owner's turn")
	}
	if hasMove[UseAttack](g.LegalMoves(1)) {
		t.Error("paralyzed P2 can attack on its turn")
	}
	endTurn(t, g) // P2's Checkup: wears off
	if opp.Has(Paralyzed) {
		t.Error("Paralysis should wear off after its owner's turn")
	}
}

// The attack's action records the damage it dealt, for the screen's hit number.
func TestAttackActionRecordsDamage(t *testing.T) {
	// Arrange
	g, _ := conditionGame()

	// Act
	if err := g.Submit(UseAttack{Player: 0, Index: 0}); err != nil {
		t.Fatalf("attack: %v", err)
	}

	// Assert
	if got := g.Actions[len(g.Actions)-1].Damage; got != freeHit.Damage {
		t.Errorf("action damage %d, want %d", got, freeHit.Damage)
	}
}
