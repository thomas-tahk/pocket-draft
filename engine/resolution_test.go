package engine

import "testing"

// recoilAndHelmet stands in for recoil plus Rocky Helmet until those card texts
// exist (milestones 3–4): damage to the attacker inside the same attack.
type recoilAndHelmet struct{ N int }

func (v recoilAndHelmet) Apply(ctx *EffectContext) { ctx.attackerMon().Damage += v.N }

// duel is a set-up game where the player to move has an ex Active whose attack
// KOs the opponent's Active and deals self Damage to itself (spec E4 example).
type duel struct {
	g        *Game
	attacker int
	defender int
}

func newDuel(t *testing.T, self int) duel {
	t.Helper()
	d0, d1 := twoDecks()
	g := NewGame(42, d0, d1)
	doSetup(t, g)
	a := g.S.Active
	ex := demoEmborEX()
	ex.Attacks[0].Damage = 100
	ex.Attacks[0].Effect = []EffectOp{recoilAndHelmet{N: self}}
	g.S.Players[a].Active = &InPlay{Card: ex, Energy: map[Energy]int{Fire: 1}}
	return duel{g: g, attacker: a, defender: 1 - a}
}

func (d duel) attack(t *testing.T) {
	t.Helper()
	if err := d.g.Submit(UseAttack{Player: d.attacker, Index: 0}); err != nil {
		t.Fatalf("attack: %v", err)
	}
}

func clearBench(pl *Player) { pl.Bench = [BenchSlots]*InPlay{} }

func assertOver(t *testing.T, g *Game, reason EndReason, winner int) {
	t.Helper()
	if g.S.Phase != PhaseOver || g.S.Reason != reason || g.S.Winner != winner {
		t.Fatalf("phase=%v reason=%v winner=%d, want over by %v, winner %d",
			g.S.Phase, g.S.Reason, g.S.Winner, reason, winner)
	}
}

// E3/E4: the attack takes the last point and its own recoil + Rocky Helmet KO
// the ex attacker in the same action — both reach 3, so it is a tie.
func TestRecoilDoubleKOAtThreeEachIsTie(t *testing.T) {
	// Arrange
	d := newDuel(t, 70)
	d.g.S.Players[d.attacker].Points = 2
	d.g.S.Players[d.defender].Points = 1

	// Act
	d.attack(t)

	// Assert
	assertOver(t, d.g, ReasonPoints, -1)
	if p0, p1 := d.g.S.Players[0].Points, d.g.S.Players[1].Points; p0 != 3 || p1 != 3 {
		t.Errorf("points %d-%d, want 3-3", p0, p1)
	}
}

// E3: a both-≥3 tie holds whatever the Benches hold — even an empty one.
func TestBothThreeTieIgnoresEmptyBench(t *testing.T) {
	// Arrange
	d := newDuel(t, 70)
	d.g.S.Players[d.attacker].Points = 2
	d.g.S.Players[d.defender].Points = 1
	clearBench(&d.g.S.Players[d.defender])

	// Act
	d.attack(t)

	// Assert
	assertOver(t, d.g, ReasonPoints, -1)
}

// E4: Poison at the next Checkup is a separate action; the attacker already won.
func TestCheckupKOAfterWinningAttackIsNotTie(t *testing.T) {
	// Arrange
	d := newDuel(t, 0)
	d.g.S.Players[d.attacker].Points = 2
	d.g.S.Players[d.defender].Points = 1
	atk := d.g.S.Players[d.attacker].Active
	atk.Poisoned = true
	atk.Damage = atk.Card.HP - 5

	// Act
	d.attack(t)

	// Assert
	assertOver(t, d.g, ReasonPoints, d.attacker)
	if got := d.g.S.Players[d.attacker].Active; got == nil {
		t.Error("attacker should still be in play: Checkup never ran")
	}
}

// E5: both short of 3, the defender has no Pokémon left → attacker wins.
func TestNoPokemonLeftLoses(t *testing.T) {
	// Arrange
	d := newDuel(t, 0)
	clearBench(&d.g.S.Players[d.defender])

	// Act
	d.attack(t)

	// Assert
	assertOver(t, d.g, ReasonNoPokemon, d.attacker)
}

// E5: both short of 3 and neither has a Pokémon left → tie.
func TestNeitherHasPokemonLeftIsTie(t *testing.T) {
	// Arrange
	d := newDuel(t, 70)
	clearBench(&d.g.S.Players[0])
	clearBench(&d.g.S.Players[1])

	// Act
	d.attack(t)

	// Assert
	assertOver(t, d.g, ReasonNoPokemon, -1)
}

// E4: a double KO short of 3 removes both, scores both, and both promote.
func TestDoubleKOShortOfThreeScoresBoth(t *testing.T) {
	// Arrange
	d := newDuel(t, 70)

	// Act
	d.attack(t)
	gone := d.g.S.Players[0].Active == nil && d.g.S.Players[1].Active == nil
	a, b := d.g.S.Players[d.attacker].Points, d.g.S.Players[d.defender].Points
	for i := 0; i < 2 && d.g.S.Pending != nil; i++ {
		p := d.g.S.Pending.Player
		slot := firstSlot(d.g.S.Players[p], true)
		if err := d.g.Submit(ChooseNewActive{Player: p, BenchIndex: slot}); err != nil {
			t.Fatalf("promote for %s: %v", side(p), err)
		}
	}

	// Assert
	if d.g.S.Phase == PhaseOver {
		t.Fatalf("game should continue, ended by %v", d.g.S.Reason)
	}
	if !gone {
		t.Error("both Actives should be removed by the attack itself, before any Checkup")
	}
	if a != 1 || b != 2 {
		t.Errorf("after the attack, attacker/defender points %d/%d, want 1/2", a, b)
	}
	for i := 0; i < 2; i++ {
		if d.g.S.Players[i].Active == nil {
			t.Errorf("%s has no Active after both promotions", side(i))
		}
	}
}
