package engine

import "testing"

// koDefender sets up a duel whose attack KOs the defender's Active, with the
// defender's bench holding one Pokémon per given slot.
func koDefender(t *testing.T, slots ...int) duel {
	t.Helper()
	d := newDuel(t, 0)
	pl := &d.g.S.Players[d.defender]
	clearBench(pl)
	for i, s := range slots {
		pl.Bench[s] = newInPlay(legalBasic(string(rune('A'+i)), "Mon", 1), 0)
	}
	return d
}

// F8: with one Pokémon on the bench the new Active is chosen for the player;
// no prompt reaches them and play goes on to the next turn.
func TestPromptWithOneOptionResolvesItself(t *testing.T) {
	// Arrange
	d := koDefender(t, BenchRight)
	only := d.g.S.Players[d.defender].Bench[BenchRight]

	// Act
	d.attack(t)

	// Assert
	if d.g.S.Pending != nil {
		t.Fatalf("a one-option prompt reached the player: %+v", d.g.S.Pending)
	}
	pl := d.g.S.Players[d.defender]
	if pl.Active != only || pl.Bench[BenchRight] != nil {
		t.Error("the only benched Pokémon should have been promoted from its slot")
	}
	if d.g.S.Active != d.defender {
		t.Errorf("turn should have passed to %s", side(d.defender))
	}
}

// F8: with two choices the prompt waits for the right player and lists exactly
// the choices, which are the moves that player may answer it with.
func TestPromptWithTwoOptionsReachesThePlayer(t *testing.T) {
	// Arrange
	d := koDefender(t, BenchLeft, BenchRight)

	// Act
	d.attack(t)

	// Assert
	pr := d.g.S.Pending
	if pr == nil || pr.Player != d.defender || pr.Kind != PromptNewActive {
		t.Fatalf("want a new-Active prompt for %s, got %+v", side(d.defender), pr)
	}
	assertMoves(t, pr.Options, []Event{
		ChooseNewActive{Player: d.defender, BenchIndex: BenchLeft},
		ChooseNewActive{Player: d.defender, BenchIndex: BenchRight},
	})
}

// A choice the engine makes by itself is derived during the fold, so it is
// never logged as an event (ADR-0005) — here after a poison knockout.
func TestAutoChoiceIsNotLogged(t *testing.T) {
	// Arrange
	d0, d1 := twoDecks()
	g := NewGame(42, d0, d1)
	doSetup(t, g)
	opp := &g.S.Players[1-g.S.Active]
	clearBench(opp)
	opp.Bench[BenchBehind] = newInPlay(legalBasic("Z", "Mon", 1), 0)
	opp.Active.Damage = opp.Active.Card.HP - 5
	opp.Active.Poisoned = true

	// Act
	if err := g.Submit(EndTurn{Player: g.S.Active}); err != nil {
		t.Fatalf("end turn: %v", err)
	}

	// Assert
	if g.S.Pending != nil || opp.Active == nil || opp.Active.Card.ID != "Z" {
		t.Fatalf("poison KO with one benched Pokémon should promote it, pending %+v", g.S.Pending)
	}
	if n := len(g.Events); n != 3 {
		t.Errorf("the automatic choice should not be logged as an event; %d events", n)
	}
}
