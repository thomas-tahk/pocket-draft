package engine

import "testing"

// twoBenched gives a player Pokémon in bench slots 0 and 2 only, so a
// promotion has two answers and reaches the player as a prompt.
func twoBenched(pl *Player, c Card) {
	clearBench(pl)
	pl.Bench[0] = &InPlay{Card: c, Energy: map[Energy]int{}}
	pl.Bench[2] = &InPlay{Card: c, Energy: map[Energy]int{}}
}

// duelAs is newDuel with the attack on attacker's turn, so the order rule is
// checked from both seats rather than whichever seat the seed starts.
func duelAs(t *testing.T, self, attacker int) duel {
	t.Helper()
	d := newDuel(t, self)
	if d.attacker == attacker {
		return d
	}
	ex := d.g.S.Players[d.attacker].Active
	if err := d.g.Submit(EndTurn{Player: d.attacker}); err != nil {
		t.Fatalf("end turn: %v", err)
	}
	d.g.S.Players[attacker].Active = ex
	return duel{g: d.g, attacker: attacker, defender: 1 - attacker}
}

func promoteSlot(t *testing.T, g *Game, p, slot int) {
	t.Helper()
	if err := g.Submit(ChooseNewActive{Player: p, BenchIndex: slot}); err != nil {
		t.Fatalf("promote for %s: %v", side(p), err)
	}
}

func assertPending(t *testing.T, g *Game, p int) {
	t.Helper()
	if g.S.Pending == nil || g.S.Pending.Kind != PromptNewActive || g.S.Pending.Player != p {
		t.Fatalf("expected a new-Active prompt for %s, got %+v", side(p), g.S.Pending)
	}
}

// E6: after a double KO from an attack, the defender (whose turn is next)
// promotes first; the attacker cannot see that pick until it has placed its own.
func TestDoubleKOAttackNextPlayerPromotesFirstHidden(t *testing.T) {
	for attacker := 0; attacker < 2; attacker++ {
		doubleKOAttackHidden(t, attacker)
	}
}

func doubleKOAttackHidden(t *testing.T, attacker int) {
	// Arrange
	d := duelAs(t, 70, attacker)
	twoBenched(&d.g.S.Players[d.attacker], demoEmbor())
	twoBenched(&d.g.S.Players[d.defender], demoVolt())

	// Act
	d.attack(t)
	assertPending(t, d.g, d.defender)
	if _, hidden := d.g.HiddenPromotion(); hidden {
		t.Fatal("nothing is hidden before anyone has promoted")
	}
	promoteSlot(t, d.g, d.defender, 2)

	// Assert
	assertPending(t, d.g, d.attacker)
	hp, hidden := d.g.HiddenPromotion()
	if !hidden || hp.Player != d.defender || hp.BenchIndex != 2 {
		t.Fatalf("hidden promotion = %+v, %v; want %s from bench slot 2", hp, hidden, side(d.defender))
	}
	want := side(d.defender) + " promoted " + demoVolt().Name + " to Active."
	if got := d.g.Narration[hp.Narration]; got != want {
		t.Errorf("hidden narration line = %q, want %q", got, want)
	}
	if hp.Action != len(d.g.Actions)-1 {
		t.Errorf("hidden action = %d, want the promote action %d", hp.Action, len(d.g.Actions)-1)
	}
	promoteSlot(t, d.g, d.attacker, 0)
	if _, hidden := d.g.HiddenPromotion(); hidden {
		t.Error("the first pick must be shown once the second player has placed")
	}
}

// E6: a double KO at the Checkup happens before the turn passes, so the
// opponent of the player whose turn is ending promotes first.
func TestDoubleKOCheckupNextPlayerPromotesFirst(t *testing.T) {
	for attacker := 0; attacker < 2; attacker++ {
		doubleKOCheckup(t, attacker)
	}
}

func doubleKOCheckup(t *testing.T, attacker int) {
	// Arrange
	d := duelAs(t, 0, attacker)
	twoBenched(&d.g.S.Players[d.attacker], demoEmbor())
	twoBenched(&d.g.S.Players[d.defender], demoVolt())
	for p := 0; p < 2; p++ {
		mon := d.g.S.Players[p].Active
		mon.Poisoned = true
		mon.Damage = mon.Card.HP - 10
	}

	// Act
	if err := d.g.Submit(EndTurn{Player: d.attacker}); err != nil {
		t.Fatalf("end turn: %v", err)
	}

	// Assert
	assertPending(t, d.g, d.defender)
}

// E6: when the first player has only one choice it is made for them, and it
// is still hidden while the other player picks.
func TestDoubleKOAutoPickStillHidden(t *testing.T) {
	// Arrange
	d := newDuel(t, 70)
	twoBenched(&d.g.S.Players[d.attacker], demoEmbor())
	clearBench(&d.g.S.Players[d.defender])
	d.g.S.Players[d.defender].Bench[1] = &InPlay{Card: demoVolt(), Energy: map[Energy]int{}}

	// Act
	d.attack(t)

	// Assert
	assertPending(t, d.g, d.attacker)
	hp, hidden := d.g.HiddenPromotion()
	if !hidden || hp.Player != d.defender || hp.BenchIndex != 1 || hp.Action != len(d.g.Actions)-1 {
		t.Fatalf("hidden promotion = %+v, %v; want %s from slot 1 inside the attack action", hp, hidden, side(d.defender))
	}
}
