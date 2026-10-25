package engine

import "testing"

func testFossil(id string) Card {
	return Card{ID: id, Name: "Fossil " + id, Kind: KindItem, Fossil: true, HP: 40, Type: Colorless, Weakness: NoEnergy}
}

// S3 done-gate: an opening hand whose only "Basic" is a Fossil is redrawn.
func TestFossilIsNotTheOpeningBasic(t *testing.T) {
	deck := deckOf(testFossil("F1"), 19)
	deck = append(deck, demoEmbor())
	for seed := uint64(1); seed <= 20; seed++ {
		g := NewGame(seed, deck, deck)
		for p := 0; p < 2; p++ {
			if !hasBasic(g.S.Players[p].Hand) {
				t.Fatalf("seed %d: player %d opened with no Basic: %v", seed, p, g.S.Players[p].Hand)
			}
		}
	}
}

func hasBasic(hand []Card) bool {
	for _, c := range hand {
		if c.IsBasicPokemon() {
			return true
		}
	}
	return false
}

// A5: a Fossil plays to an empty Bench slot as a 40 HP Pokémon.
func TestFossilPlaysToEmptyBench(t *testing.T) {
	g, p := mainPhaseWith(t, testFossil("F1"))
	clearBench(&g.S.Players[p])
	slot := Spot{p, 1 + BenchBehind}
	if isLegal(g, p, PlayTrainer{Player: p, CardID: "F1", Target: active(p)}) {
		t.Error("a Fossil was legal on the occupied Active spot")
	}

	mustSubmit(t, g, PlayTrainer{Player: p, CardID: "F1", Target: slot})

	mon := g.S.Players[p].Bench[BenchBehind]
	if mon == nil || mon.Card.ID != "F1" || mon.remainingHP() != 40 {
		t.Fatalf("bench slot holds %+v, want the 40 HP Fossil", mon)
	}
	if len(g.S.Players[p].Discard) != 0 {
		t.Error("a played Fossil went to the discard")
	}
}

// A5: a Fossil in the Active Spot can't retreat.
func TestFossilCantRetreat(t *testing.T) {
	g, p := mainPhaseWith(t)
	pl := &g.S.Players[p]
	clearBench(&g.S.Players[p])
	pl.Bench[0] = pl.Active
	pl.Active = newInPlay(testFossil("F1"), 0)

	if g.Submit(Retreat{Player: p, BenchIndex: 0}) == nil {
		t.Error("a Fossil retreated")
	}
}

// A5 done-gate: discarding a Fossil from play scores nothing for the opponent.
func TestDiscardFossilScoresNothing(t *testing.T) {
	g, p := mainPhaseWith(t, testFossil("F1"))
	clearBench(&g.S.Players[p])
	mustSubmit(t, g, PlayTrainer{Player: p, CardID: "F1", Target: Spot{p, 1}})

	mustSubmit(t, g, DiscardFossil{Player: p, Pos: 1})

	if g.S.Players[p].Bench[0] != nil {
		t.Error("Fossil still on the Bench")
	}
	if d := g.S.Players[p].Discard; len(d) != 1 || d[0].ID != "F1" {
		t.Errorf("discard = %v, want the Fossil", d)
	}
	if pts := g.S.Players[1-p].Points; pts != 0 {
		t.Errorf("opponent scored %d for a discarded Fossil", pts)
	}
}

// A5: only on your own turn, and only a Fossil.
func TestDiscardFossilOnlyOwnFossil(t *testing.T) {
	g, p := mainPhaseWith(t)
	if g.Submit(DiscardFossil{Player: p, Pos: 0}) == nil {
		t.Error("a real Pokémon was discarded as a Fossil")
	}
	o := 1 - p
	clearBench(&g.S.Players[o])
	g.S.Players[o].Bench[0] = newInPlay(testFossil("F1"), 0)
	if g.Submit(DiscardFossil{Player: o, Pos: 1}) == nil {
		t.Error("a Fossil was discarded on the opponent's turn")
	}
}

// Discarding the Active Fossil asks for a new Active, and the turn goes on.
func TestDiscardActiveFossilPromotes(t *testing.T) {
	g, p := mainPhaseWith(t)
	pl := &g.S.Players[p]
	clearBench(&g.S.Players[p])
	pl.Bench[0] = pl.Active
	pl.Active = newInPlay(testFossil("F1"), 0)

	mustSubmit(t, g, DiscardFossil{Player: p, Pos: 0})

	if pl.Active == nil || pl.Active.Card.Fossil {
		t.Fatalf("Active = %+v, want the benched Pokémon promoted", pl.Active)
	}
	if g.S.Active != p || g.S.Phase != PhaseMain || g.S.Pending != nil {
		t.Errorf("turn didn't go on: active %d phase %v pending %v", g.S.Active, g.S.Phase, g.S.Pending)
	}
}
