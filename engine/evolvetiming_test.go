package engine

import "testing"

// T8: the second player's first turn is game turn 2. Their Active has been in
// play since setup, yet it still may not evolve.
func TestSecondPlayerCannotEvolveOnFirstTurn(t *testing.T) {
	g := mainGame(2)
	g.S.Active = 1
	pl := &g.S.Players[1]
	pl.Active = newInPlay(legalBasic("A", "Charm", 1), 0)
	pl.Hand = []Card{legalEvo("E", "Charmel", "Charm")}

	for _, mv := range g.LegalMoves(1) {
		if _, ok := mv.(Evolve); ok {
			t.Errorf("evolution listed on P2's first turn: %#v", mv)
		}
	}
	if err := g.Submit(Evolve{Player: 1, HandCardID: "E", Target: 0}); err == nil {
		t.Error("P2 evolved on their first turn")
	}
}

// T8: a Basic played this turn may not evolve, and neither may a Pokémon that
// evolved this turn. Both are fine again on the owner's next turn.
func TestPlayedOrEvolvedThisTurnCannotEvolve(t *testing.T) {
	g := mainGame(3)
	pl := &g.S.Players[0]
	pl.Active = newInPlay(legalBasic("A", "Charm", 1), 0)
	stage2 := Card{ID: "Z", Name: "Charz", Stage: Stage2, EvolvesFrom: "Charmel", HP: 150, Type: Fire, Weakness: NoEnergy}
	pl.Hand = []Card{legalBasic("A", "Charm", 1), legalEvo("E", "Charmel", "Charm"), stage2}

	if err := g.Submit(PlayBasic{Player: 0, CardID: "A", Slot: BenchLeft}); err != nil {
		t.Fatalf("play basic: %v", err)
	}
	if err := g.Submit(Evolve{Player: 0, HandCardID: "E", Target: 1}); err == nil {
		t.Error("a Basic played this turn evolved")
	}
	if err := g.Submit(Evolve{Player: 0, HandCardID: "E", Target: 0}); err != nil {
		t.Fatalf("evolve the Active: %v", err)
	}
	if err := g.Submit(Evolve{Player: 0, HandCardID: "Z", Target: 0}); err == nil {
		t.Error("a Pokémon evolved twice in one turn")
	}
	for _, mv := range g.LegalMoves(0) {
		if _, ok := mv.(Evolve); ok {
			t.Errorf("evolution listed for a Pokémon placed or evolved this turn: %#v", mv)
		}
	}
}

// From each player's second turn on, a Pokémon in play since setup evolves.
func TestEvolveAllowedAfterOwnFirstTurn(t *testing.T) {
	for _, turn := range []int{3, 4} {
		g := mainGame(turn)
		g.S.Active = (turn - 1) % 2
		pl := &g.S.Players[g.S.Active]
		pl.Active = newInPlay(legalBasic("A", "Charm", 1), 0)
		pl.Hand = []Card{legalEvo("E", "Charmel", "Charm")}
		if err := g.Submit(Evolve{Player: g.S.Active, HandCardID: "E", Target: 0}); err != nil {
			t.Errorf("turn %d: evolve refused: %v", turn, err)
		}
	}
}
