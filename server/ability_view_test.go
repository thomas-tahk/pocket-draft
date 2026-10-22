package main

import "testing"

// A Pokémon's card view carries its Ability's name and printed text, so the
// screen can list it; a card without one carries none.
func TestCardViewCarriesAbility(t *testing.T) {
	useRealCards(t)
	c, _ := toEngineCard(rawByID["A3a-062"]) // Celesteela, Ultra Thrusters
	v := toCardView(c)
	want := rawByID["A3a-062"].Ability
	if v.Ability == nil || v.Ability.Name != want.Name || v.Ability.Effect != want.Effect {
		t.Fatalf("Ability view = %+v, want %+v", v.Ability, want)
	}
	plain, _ := toEngineCard(rawByID["A1-211"])
	if toCardView(plain).Ability != nil {
		t.Fatal("a card with no Ability should carry none")
	}
}
