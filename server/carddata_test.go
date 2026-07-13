package main

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/thomas-tahk/pocket-draft/engine"
)

func TestParseCost(t *testing.T) {
	cases := []struct {
		in   string
		want []engine.Energy
	}{
		{"RCC", []engine.Energy{engine.Fire, engine.Colorless, engine.Colorless}},
		{"WC", []engine.Energy{engine.Water, engine.Colorless}},
		{"WWW", []engine.Energy{engine.Water, engine.Water, engine.Water}},
		{"0", nil},
		{"", nil},
	}
	for _, c := range cases {
		if got := parseCost(c.in); !reflect.DeepEqual(got, c.want) {
			t.Errorf("parseCost(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestParseDamage(t *testing.T) {
	cases := map[string]int{"70": 70, "40+": 40, "100x": 100, "": 0, "20": 20}
	for in, want := range cases {
		if got := parseDamage(in); got != want {
			t.Errorf("parseDamage(%q) = %d, want %d", in, got, want)
		}
	}
}

func TestToEngineCard(t *testing.T) {
	// Arrange: a raw card shaped like Entei ex, whose "+" suffix and effect text
	// must be dropped down to base damage.
	rc := rawCard{
		ID: "B2a-126", Name: "Entei ex", CardType: "Fire", Stage: "Basic",
		HP: 140, ExKind: "ex", Weakness: "Water", Retreat: 2,
		Attacks: []rawAttack{{Cost: "RR", Name: "Blazing Beatdown", Damage: "60+", Effect: "extra stuff"}},
	}

	// Act
	card, err := toEngineCard(rc)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !card.IsEX {
		t.Error("expected IsEX true for exKind ex")
	}
	if card.Type != engine.Fire || card.Weakness != engine.Water || card.Stage != engine.Basic {
		t.Errorf("type/weakness/stage = %v/%v/%v", card.Type, card.Weakness, card.Stage)
	}
	if len(card.Attacks) != 1 || card.Attacks[0].Damage != 60 {
		t.Fatalf("attack not modeled to base damage: %+v", card.Attacks)
	}
	if !reflect.DeepEqual(card.Attacks[0].Cost, []engine.Energy{engine.Fire, engine.Fire}) {
		t.Errorf("attack cost = %v", card.Attacks[0].Cost)
	}
}

func TestToEngineCardRejectsDragon(t *testing.T) {
	// The engine has no Dragon energy; a Dragon card must error, not mistype.
	if _, err := toEngineCard(rawCard{ID: "x", Name: "Dragonite", CardType: "Dragon", Stage: "Basic"}); err == nil {
		t.Error("expected error for unsupported Dragon type")
	}
}

func TestDeckFromIDs(t *testing.T) {
	// Arrange: 10 distinct Basic Fire cards and 10 distinct Stage-2 Fire cards, so
	// two copies of each name makes a format-legal 20-card list either way.
	rawByID = map[string]rawCard{}
	basics := make([]string, 0, 20)
	stage2 := make([]string, 0, 20)
	for i := 0; i < 10; i++ {
		b := fmt.Sprintf("B%d", i)
		s := fmt.Sprintf("S%d", i)
		rawByID[b] = rawCard{ID: b, Name: "Basic" + b, CardType: "Fire", Stage: "Basic", HP: 60,
			Attacks: []rawAttack{{Cost: "R", Name: "Hit", Damage: "20"}}}
		rawByID[s] = rawCard{ID: s, Name: "Final" + s, CardType: "Fire", Stage: "Stage 2", HP: 150,
			Attacks: []rawAttack{{Cost: "RR", Name: "Big", Damage: "90"}}}
		basics = append(basics, b, b)
		stage2 = append(stage2, s, s)
	}

	rep := func(id string, n int) []string {
		out := make([]string, n)
		for i := range out {
			out[i] = id
		}
		return out
	}

	// Happy path: exactly 20, ≤2 by name, has Basics.
	deck, err := deckFromIDs(basics)
	if err != nil {
		t.Fatalf("legal deck: unexpected error: %v", err)
	}
	if len(deck) != 20 {
		t.Fatalf("legal deck: got %d cards, want 20", len(deck))
	}

	// Each of these must be rejected.
	bad := []struct {
		name string
		ids  []string
	}{
		{"not exactly 20", rep("B0", 19)},
		{"more than 2 by name", append(rep("B0", 3), basics[3:]...)}, // 3×B0 + 17 = 20
		{"no Basic Pokémon", stage2},
		{"unknown card id", rep("NOPE", 20)},
	}
	for _, c := range bad {
		if _, err := deckFromIDs(c.ids); err == nil {
			t.Errorf("%s: expected error, got nil", c.name)
		}
	}
}
