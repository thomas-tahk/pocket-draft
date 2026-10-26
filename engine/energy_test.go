package engine

import (
	"reflect"
	"testing"
)

func withCost(c Card, id string, cost ...Energy) Card {
	c.ID = id
	c.Attacks = []Attack{{Name: "Hit", Cost: cost, Damage: 10}}
	return c
}

// Spec S2: the default is the types the deck's attacks need, most-needed first,
// at most 3. Colorless costs need no type; a Dragon's typed cost counts.
func TestDefaultEnergyTypesForSampleDecks(t *testing.T) {
	fire, lightning := DemoDecks()
	dragon := Card{ID: "DRG", Name: "Dragonite", Type: Dragon, HP: 140}
	colorless := Card{ID: "CLR", Name: "Pidgey", Type: Colorless, HP: 50}
	cases := []struct {
		name string
		deck []Card
		want []Energy
	}{
		{"mono fire demo", fire, []Energy{Fire}},
		{"mono lightning demo", lightning, []Energy{Lightning}},
		{"fire pokemon with colorless-only attacks need nothing typed",
			deckOf(withCost(demoEmbor(), "a", Colorless, Colorless), 20), []Energy{Colorless}},
		{"dragon's typed cost counts, by demand",
			append(deckOf(withCost(dragon, "d", Water, Water, Lightning, Lightning), 4),
				deckOf(withCost(demoVolt(), "v", Lightning), 4)...),
			[]Energy{Lightning, Water}},
		{"capped at the 3 most-needed types",
			append(append(append(
				deckOf(withCost(colorless, "g", Grass), 1),
				deckOf(withCost(colorless, "w", Water), 2)...),
				deckOf(withCost(colorless, "r", Fire), 3)...),
				deckOf(withCost(colorless, "p", Psychic), 4)...),
			[]Energy{Psychic, Fire, Water}},
	}
	for _, c := range cases {
		if got := DefaultEnergyTypes(c.deck); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

// Spec S2: each turn the Energy Zone rolls only among the deck's chosen types,
// not the types its Pokémon happen to be.
func TestZoneRollsOnlyChosenTypes(t *testing.T) {
	d0, d1 := twoDecks()
	chosen := []Energy{Water, Grass}
	g := NewGameFromDecks(7, [2]Deck{{Cards: d0, Energy: chosen}, {Cards: d1}})
	doSetup(t, g)
	seen := map[Energy]bool{}
	for g.S.Turn < 40 && g.S.Phase != PhaseOver {
		p := g.S.Active
		zone := g.S.Players[p].EnergyZone
		if p == 0 && zone != NoEnergy {
			if zone != Water && zone != Grass {
				t.Fatalf("turn %d: zone rolled %v, chosen %v", g.S.Turn, zone, chosen)
			}
			seen[zone] = true
		}
		if p == 1 && zone != NoEnergy && zone != Lightning {
			t.Fatalf("turn %d: default deck rolled %v, want Lightning", g.S.Turn, zone)
		}
		if err := g.Submit(EndTurn{Player: p}); err != nil {
			t.Fatalf("end turn: %v", err)
		}
	}
	if !seen[Water] || !seen[Grass] {
		t.Errorf("20 turns should roll both chosen types, saw %v", seen)
	}
}

func TestValidateEnergyTypes(t *testing.T) {
	ok := [][]Energy{{Fire}, {Fire, Water}, {Grass, Water, Metal}}
	bad := [][]Energy{nil, {Fire, Water, Grass, Metal}, {Fire, Fire}, {Dragon}, {Colorless}, {NoEnergy}}
	for _, ts := range ok {
		if err := ValidateEnergyTypes(ts); err != nil {
			t.Errorf("%v rejected: %v", ts, err)
		}
	}
	for _, ts := range bad {
		if err := ValidateEnergyTypes(ts); err == nil {
			t.Errorf("%v accepted", ts)
		}
	}
}

// #127: the deckbuild picker's default (src/game/energy.ts defaultZoneTypes) mirrors this;
// e2e/energy-default.spec.ts checks the two agree on real cards. Same deck as there:
// two Druddigon (Fire, Water, Colorless) plus Psychic attackers. A Dragon's off-type
// costs count, so Fire and Water are not lost to Psychic.
func TestDefaultEnergyTypesDruddigonWithPsychic(t *testing.T) {
	dragon := Card{ID: "A1a-056", Name: "Druddigon", Type: Dragon, HP: 100}
	psychic := Card{ID: "PSY", Name: "Psy", Type: Psychic, HP: 60}
	deck := append(deckOf(withCost(dragon, "A1a-056", Fire, Water, Colorless), 2), deckOf(withCost(psychic, "PSY", Psychic), 2)...)
	if got, want := DefaultEnergyTypes(deck), []Energy{Fire, Water, Psychic}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
