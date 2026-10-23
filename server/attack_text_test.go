package main

import "testing"

// An attack's printed effect text and damage reach the browser with its card,
// exactly as printed; an attack with no text sends none.
func TestAttackViewCarriesPrintedText(t *testing.T) {
	useRealCards(t)
	cases := map[string]string{
		"B4-020":  "Take a [R] Energy from your Energy Zone and attach it to this Pokémon.", // Ponyta, Stoke
		"B4a-001": "",                                                                       // Volbeat, Tackle
	}
	for id, want := range cases {
		card, err := toEngineCard(rawByID[id])
		if err != nil {
			t.Fatal(err)
		}
		if got := toCardView(card).Attacks[0].Effect; got != want {
			t.Errorf("%s: effect = %q, want %q", id, got, want)
		}
	}
}

func TestAttackViewCarriesPrintedDamage(t *testing.T) {
	useRealCards(t)
	card, err := toEngineCard(rawByID["A4b-076"]) // Heatran, Ragin' Mad Strike 40+
	if err != nil {
		t.Fatal(err)
	}
	if got := toCardView(card).Attacks[0].Printed; got != "40+" {
		t.Errorf("printed = %q, want 40+", got)
	}
}
