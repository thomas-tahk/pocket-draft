package main

import "testing"

func TestTallyTypes(t *testing.T) {
	picks := []draftCard{
		{rawCard: rawCard{CardType: "Fire"}},
		{rawCard: rawCard{CardType: "Fire"}},
		{rawCard: rawCard{CardType: "Trainer"}},
		{rawCard: rawCard{CardType: "Colorless"}},
		{rawCard: rawCard{CardType: "Water"}},
	}
	tally := tallyTypes(picks)
	if tally["Fire"] != 2 {
		t.Errorf("tally[Fire] = %d, want 2", tally["Fire"])
	}
	if tally["Water"] != 1 {
		t.Errorf("tally[Water] = %d, want 1", tally["Water"])
	}
	if _, ok := tally["Trainer"]; ok {
		t.Errorf("Trainers must not be tallied by type")
	}
	if _, ok := tally["Colorless"]; ok {
		t.Errorf("Colorless must not be tallied by type")
	}
}

func TestWeightOf(t *testing.T) {
	tally := typeTally{"Fire": 5}
	trainer := draftCard{rawCard: rawCard{CardType: "Trainer"}}
	colorless := draftCard{rawCard: rawCard{CardType: "Colorless"}}
	fire := draftCard{rawCard: rawCard{CardType: "Fire"}}

	if w := weightOf(trainer, tally); w != 1 {
		t.Errorf("weightOf(Trainer) = %v, want 1", w)
	}
	if w := weightOf(colorless, tally); w != 1 {
		t.Errorf("weightOf(Colorless) = %v, want 1", w)
	}
	// TYPE_WEIGHT_COEF is 0 (a dead no-op term, ported as-is) so even a heavy
	// tally must not move the weight off 1.
	if w := weightOf(fire, tally); w != 1 {
		t.Errorf("weightOf(Fire, heavy tally) = %v, want 1 (TYPE_WEIGHT_COEF=0)", w)
	}
}
