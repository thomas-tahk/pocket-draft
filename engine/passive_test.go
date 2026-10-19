package engine

import (
	"reflect"
	"testing"
)

// Test passives modelled on real texts. The real cards are wired in server/.

// fightingCoach: Lucario's "Attacks used by your [F] Pokémon do +20 damage to
// your opponent's Active Pokémon."
var fightingCoach = Passive{Hook: HookDamageDealt, Value: 20, When: func(c PassiveCheck) bool {
	return c.Target.Player == c.Source.Player && c.Target.Card.Type == Fighting
}}

// arenaOfAntiquity: "Attacks used by each [F] Pokémon in play (both yours and
// your opponent's) do +20 damage to the opponent's Active Pokémon ex."
var arenaOfAntiquity = Passive{Hook: HookDamageDealt, Value: 20, When: func(c PassiveCheck) bool {
	return c.Target.Card.Type == Fighting && c.Foe != nil && c.Foe.Card.IsEX
}}

func fighter(id, name string) Card {
	return Card{ID: id, Name: name, Stage: Basic, HP: 100, Type: Fighting, Weakness: NoEnergy,
		Attacks: []Attack{{Name: "Punch", Damage: 30}}}
}

// lucarioGame: P1's Fighting Active faces an ex, a Lucario with Fighting
// Coach sits on P1's Bench, and Arena of Antiquity is the Stadium.
func lucarioGame() *Game {
	g := mainGame(3)
	g.S.Players[0].Active = newInPlay(fighter("ATK", "Hitmonchan"), 0)
	lucario := fighter("LUC", "Lucario")
	lucario.Ability = &Ability{Name: "Fighting Coach", Passive: []Passive{fightingCoach}}
	g.S.Players[0].Bench[BenchBehind] = newInPlay(lucario, 0)
	foe := legalBasic("EX", "Foe ex", 1)
	foe.HP, foe.IsEX = 200, true
	g.S.Players[1].Active = newInPlay(foe, 0)
	g.S.Stadium = &Card{ID: "STA", Name: "Arena of Antiquity", Passive: []Passive{arenaOfAntiquity}}
	return g
}

// L6: a Stadium and a Benched Lucario show as two modifiers, each with its
// source and value.
func TestModifiersListEachSourceAndValue(t *testing.T) {
	// Arrange
	g := lucarioGame()

	// Act
	got := g.Modifiers(0, 0)

	// Assert
	want := []Modifier{
		{Source: "Lucario", Hook: HookDamageDealt, Value: 20},
		{Source: "Arena of Antiquity", Hook: HookDamageDealt, Value: 20},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Modifiers = %+v, want %+v", got, want)
	}
}

// T11: both damage modifiers land on the attack.
func TestDamageDealtModifiersAddUp(t *testing.T) {
	// Arrange
	g := lucarioGame()

	// Act
	if err := g.Submit(UseAttack{Player: 0, Index: 0}); err != nil {
		t.Fatalf("attack: %v", err)
	}

	// Assert
	if got := g.S.Players[1].Active.Damage; got != 70 {
		t.Errorf("damage = %d, want 30 + 20 + 20", got)
	}
}

// T11: "This Pokémon takes −20 damage from attacks" is applied after
// Weakness, and never drives damage below 0.
func TestDamageTakenReducesAfterWeakness(t *testing.T) {
	cases := []struct {
		name   string
		reduce int
		want   int
	}{
		{"reduced", -20, 30}, // 30 +20 weakness −20
		{"floored", -100, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			g := mainGame(3)
			g.S.Players[0].Active = newInPlay(fighter("ATK", "Hitmonchan"), 0)
			def := legalBasic("DEF", "Tank", 1)
			def.Weakness = Fighting
			def.Ability = &Ability{Name: "Thick", Passive: []Passive{{Hook: HookDamageTaken, Value: tc.reduce, When: self}}}
			g.S.Players[1].Active = newInPlay(def, 0)

			// Act
			if err := g.Submit(UseAttack{Player: 0, Index: 0}); err != nil {
				t.Fatalf("attack: %v", err)
			}

			// Assert
			if got := g.S.Players[1].Active.Damage; got != tc.want {
				t.Errorf("damage = %d, want %d", got, tc.want)
			}
		})
	}
}

// self is "This Pokémon …": the rule is about the Pokémon printing it.
func self(c PassiveCheck) bool { return c.Target == c.Source }

// T11: Bounded Field turns Weakness from +20 into ×2.
func TestWeaknessDouble(t *testing.T) {
	// Arrange
	g := mainGame(3)
	g.S.Players[0].Active = newInPlay(fighter("ATK", "Hitmonchan"), 0)
	def := legalBasic("DEF", "Weak", 1)
	def.HP, def.Weakness = 100, Fighting
	g.S.Players[1].Active = newInPlay(def, 0)
	g.S.Stadium = &Card{Name: "Bounded Field", Passive: []Passive{{Hook: HookWeaknessDouble}}}

	// Act
	if err := g.Submit(UseAttack{Player: 0, Index: 0}); err != nil {
		t.Fatalf("attack: %v", err)
	}

	// Assert
	if got := g.S.Players[1].Active.Damage; got != 60 {
		t.Errorf("damage = %d, want 30 ×2", got)
	}
}

// An attack that does no damage stays at 0: damage modifiers need damage.
func TestModifiersSkipAttacksWithoutDamage(t *testing.T) {
	// Arrange
	g := lucarioGame()
	g.S.Players[0].Active.Card.Attacks[0].Damage = 0

	// Act
	if err := g.Submit(UseAttack{Player: 0, Index: 0}); err != nil {
		t.Fatalf("attack: %v", err)
	}

	// Assert
	if got := g.S.Players[1].Active.Damage; got != 0 {
		t.Errorf("damage = %d, want 0", got)
	}
}

// Retreat cost: a Benched Pokémon makes the Active's cost 1 less; "no Retreat
// Cost" wins over any amount; cost never goes below 0.
func TestRetreatCostHooks(t *testing.T) {
	cheaper := Passive{Hook: HookRetreatCost, Value: -1, When: func(c PassiveCheck) bool {
		return c.Target.Player == c.Source.Player && c.Target.Pos == 0
	}}
	cases := []struct {
		name    string
		passive []Passive
		cost    int
		want    int
	}{
		{"none", nil, 2, 2},
		{"one less", []Passive{cheaper}, 2, 1},
		{"floored", []Passive{cheaper}, 0, 0},
		{"no cost", []Passive{{Hook: HookNoRetreatCost}}, 3, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			g := mainGame(3)
			g.S.Players[0].Active = newInPlay(legalBasic("A", "Heavy", tc.cost), 0)
			helper := legalBasic("B", "Helper", 1)
			if tc.passive != nil {
				helper.Ability = &Ability{Name: "Help", Passive: tc.passive}
			}
			g.S.Players[0].Bench[BenchLeft] = newInPlay(helper, 0)

			// Act
			got := g.RetreatCost(0)

			// Assert
			if got != tc.want {
				t.Errorf("RetreatCost = %d, want %d", got, tc.want)
			}
		})
	}
}

// A lowered retreat cost is what retreating actually charges.
func TestRetreatChargesTheModifiedCost(t *testing.T) {
	// Arrange
	g := mainGame(3)
	g.S.Players[0].Active = newInPlay(legalBasic("A", "Heavy", 1), 0)
	g.S.Players[0].Bench[BenchLeft] = newInPlay(legalBasic("B", "Bencher", 1), 0)
	g.S.Stadium = &Card{Name: "Free Ride", Passive: []Passive{{Hook: HookNoRetreatCost}}}

	// Act
	if err := g.Submit(Retreat{Player: 0, BenchIndex: BenchLeft}); err != nil {
		t.Fatalf("retreat with no energy: %v", err)
	}

	// Assert
	if g.S.Players[0].Active.Card.ID != "B" {
		t.Error("the free retreat did not swap")
	}
}

// C9: a Pokémon that "can't be affected by any Special Conditions" isn't.
func TestNoConditionsBlocksStatus(t *testing.T) {
	// Arrange
	g := mainGame(3)
	g.S.Players[0].Active = newInPlay(Card{Name: "Burner", HP: 60, Type: Fire, Weakness: NoEnergy,
		Attacks: []Attack{{Name: "Singe", Effect: []EffectOp{ApplyStatus{Status: Burn}}}}}, 0)
	def := legalBasic("DEF", "Ogerpon", 1)
	def.Ability = &Ability{Name: "Guard", Passive: []Passive{{Hook: HookNoConditions, When: self}}}
	g.S.Players[1].Active = newInPlay(def, 0)

	// Act
	if err := g.Submit(UseAttack{Player: 0, Index: 0}); err != nil {
		t.Fatalf("attack: %v", err)
	}

	// Assert
	if g.S.Players[1].Active.Burned {
		t.Error("a Pokémon immune to conditions was Burned")
	}
}

// T8: a passive can let a Pokémon evolve on its first turn or the turn it
// was played.
func TestEvolveEarlyHook(t *testing.T) {
	// Arrange
	g := mainGame(1)
	pl := &g.S.Players[0]
	baby := legalBasic("A", "Eevee", 1)
	baby.Ability = &Ability{Name: "Boost", Passive: []Passive{{Hook: HookEvolveEarly, When: self}}}
	pl.Active = newInPlay(baby, 1)
	pl.Hand = []Card{legalEvo("E", "Vaporeon", "Eevee")}

	// Act
	err := g.Submit(Evolve{Player: 0, HandCardID: "E", Target: 0})

	// Assert
	if err != nil {
		t.Errorf("evolve refused: %v", err)
	}
}
