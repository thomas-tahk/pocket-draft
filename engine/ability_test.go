package engine

import "testing"

// drawer has "Once during your turn, if this Pokémon is in the Active Spot,
// you may draw a card." with the spot rule given by where.
func drawer(id string, where AbilitySpot, once bool) Card {
	c := legalBasic(id, "Drawer", 1)
	c.Ability = &Ability{Name: "Draw", Use: []EffectOp{DrawCards{N: 1}}, Where: where, OncePerTurn: once}
	return c
}

func abilityGame(where AbilitySpot, once bool) *Game {
	g := mainGame(3)
	pl := &g.S.Players[0]
	pl.Active = newInPlay(drawer("ACT", where, once), 0)
	pl.Bench[BenchLeft] = newInPlay(drawer("BEN", where, once), 0)
	pl.Deck = []Card{legalBasic("D1", "D1", 1), legalBasic("D2", "D2", 1), legalBasic("D3", "D3", 1)}
	return g
}

func abilityMoves(g *Game, p int) []Event {
	var out []Event
	for _, mv := range g.LegalMoves(p) {
		if _, ok := mv.(UseAbility); ok {
			out = append(out, mv)
		}
	}
	return out
}

// A4: an ability whose text says once per turn can't be used twice that turn.
func TestOncePerTurnAbilityIsNotLegalTwice(t *testing.T) {
	// Arrange
	g := abilityGame(AnySpot, true)
	use := UseAbility{Player: 0, Pos: 0}
	if err := g.Submit(use); err != nil {
		t.Fatalf("first use: %v", err)
	}

	// Act
	err := g.Submit(use)

	// Assert
	if err == nil {
		t.Error("a once-per-turn ability was used twice in one turn")
	}
	if len(g.S.Players[0].Hand) != 1 {
		t.Errorf("hand = %d cards, want the 1 drawn", len(g.S.Players[0].Hand))
	}
	assertMoves(t, abilityMoves(g, 0), []Event{UseAbility{Player: 0, Pos: 1 + BenchLeft}})
}

// A4: once per turn is per Pokémon and per turn: it is back on the owner's
// next turn.
func TestOncePerTurnAbilityReturnsNextTurn(t *testing.T) {
	// Arrange
	g := abilityGame(AnySpot, true)
	if err := g.Submit(UseAbility{Player: 0, Pos: 0}); err != nil {
		t.Fatalf("use: %v", err)
	}

	// Act
	g.S.Turn += 2

	// Assert
	if err := g.Submit(UseAbility{Player: 0, Pos: 0}); err != nil {
		t.Errorf("ability refused on the next turn: %v", err)
	}
}

// A4: an ability without "once during your turn" can be used again.
func TestRepeatableAbility(t *testing.T) {
	// Arrange
	g := abilityGame(AnySpot, false)
	use := UseAbility{Player: 0, Pos: 0}

	// Act
	err1, err2 := g.Submit(use), g.Submit(use)

	// Assert
	if err1 != nil || err2 != nil {
		t.Errorf("repeatable ability refused: %v, %v", err1, err2)
	}
}

// "if this Pokémon is in the Active Spot" / "on your Bench" limit where it
// can be used from; with neither it works from either.
func TestAbilitySpot(t *testing.T) {
	cases := []struct {
		name  string
		where AbilitySpot
		want  []Event
	}{
		{"anywhere", AnySpot, []Event{UseAbility{0, 0, Spot{}}, UseAbility{0, 1 + BenchLeft, Spot{}}}},
		{"active only", ActiveSpot, []Event{UseAbility{0, 0, Spot{}}}},
		{"bench only", BenchSpot, []Event{UseAbility{0, 1 + BenchLeft, Spot{}}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			g := abilityGame(tc.where, true)

			// Act
			got := abilityMoves(g, 0)

			// Assert
			assertMoves(t, got, tc.want)
		})
	}
}

// Only on your own turn, only with an activated ability, and a passive-only
// Ability has nothing to use.
func TestAbilityNeedsYourTurnAndAUse(t *testing.T) {
	// Arrange
	g := abilityGame(AnySpot, true)
	g.S.Players[1].Active = newInPlay(drawer("OPP", AnySpot, true), 0)
	passive := legalBasic("PAS", "Passive", 1)
	passive.Ability = &Ability{Name: "Aura", Passive: []Passive{{Hook: HookNoRetreatCost}}}
	g.S.Players[0].Bench[BenchRight] = newInPlay(passive, 0)

	// Act
	opp, mine := abilityMoves(g, 1), abilityMoves(g, 0)

	// Assert
	assertMoves(t, opp, nil)
	assertMoves(t, mine, []Event{UseAbility{0, 0, Spot{}}, UseAbility{0, 1 + BenchLeft, Spot{}}})
}

// An ability's condition from its text (When) is checked before it is legal.
func TestAbilityWhen(t *testing.T) {
	// Arrange
	g := abilityGame(AnySpot, true)
	g.S.Players[0].Active.Card.Ability.When = func(c AbilityCheck) bool { return c.Self.Damage > 0 }

	// Act
	got := abilityMoves(g, 0)

	// Assert
	assertMoves(t, got, []Event{UseAbility{0, 1 + BenchLeft, Spot{}}})
}

// Using an ability is recorded and replays.
func TestAbilityReplays(t *testing.T) {
	// Arrange
	d0, d1 := twoDecks()
	d0[0] = drawer("DRW", AnySpot, true)
	for i := range d0 {
		d0[i] = drawer(d0[i].ID, AnySpot, true)
	}
	g := NewGame(7, d0, d1)
	doSetup(t, g)
	p := g.S.Active
	if p != 0 {
		endTurn(t, g)
	}
	if err := g.Submit(UseAbility{Player: 0, Pos: 0}); err != nil {
		t.Fatalf("use: %v", err)
	}

	// Act
	r, err := Replay(7, d0, d1, g.Events)

	// Assert
	if err != nil {
		t.Fatalf("replay: %v", err)
	}
	if len(r.S.Players[0].Hand) != len(g.S.Players[0].Hand) {
		t.Error("replay diverged")
	}
	if a := g.Actions[len(g.Actions)-1]; a.Kind != ActAbility || a.Target != (Spot{0, 0}) {
		t.Errorf("action = %+v, want an ability action on P1's Active", a)
	}
}

// alchemy: Alolan Muk's "Basic Pokémon in play (both yours and your
// opponent's) have no Abilities."
var alchemy = Passive{Hook: HookNoAbilities, When: func(c PassiveCheck) bool { return c.Target.Card.Stage == Basic }}

func alchemyGame() *Game {
	g := abilityGame(AnySpot, true)
	muk := legalEvo("MUK", "Alolan Muk", "Grimer")
	muk.Ability = &Ability{Name: "Power of Alchemy", Passive: []Passive{alchemy}}
	g.S.Players[1].Bench[BenchLeft] = newInPlay(muk, 0)
	return g
}

// "have no Abilities": a Basic's activated Ability can't be used.
func TestNoAbilitiesBlocksUse(t *testing.T) {
	// Arrange
	g := alchemyGame()

	// Act
	moves, err := abilityMoves(g, 0), g.Submit(UseAbility{Player: 0, Pos: 0})

	// Assert
	assertMoves(t, moves, nil)
	if err == nil {
		t.Error("a Basic used its Ability under Power of Alchemy")
	}
}

// "have no Abilities": a Basic's passive stops; a Stage 1's keeps working.
func TestNoAbilitiesStopsPassives(t *testing.T) {
	// Arrange
	g := alchemyGame()
	mon := g.S.Players[0].Active
	mon.Card.Ability = &Ability{Name: "Float", Passive: []Passive{{Hook: HookNoRetreatCost, When: PassiveCheck.Self}}}
	mon.Card.RetreatCost = 2

	// Act
	basic := g.RetreatCost(0)
	mon.Card.Stage = Stage1
	stage1 := g.RetreatCost(0)

	// Assert
	if basic != 2 || stage1 != 0 {
		t.Errorf("retreat cost Basic %d, Stage 1 %d; want 2 and 0", basic, stage1)
	}
}

// A no-Abilities passive can read the board: its When is handed the game.
func TestNoAbilitiesPassiveSeesTheGame(t *testing.T) {
	g := alchemyGame()
	muk := g.S.Players[1].Bench[BenchLeft]
	muk.Card.Ability.Passive = []Passive{{Hook: HookNoAbilities, When: func(c PassiveCheck) bool { return c.Game.S.Turn > 0 }}}

	if err := g.Submit(UseAbility{Player: 0, Pos: 0}); err == nil {
		t.Error("a Basic used its Ability while a no-Abilities passive applied")
	}
}
