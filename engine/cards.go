package engine

// Demo vanilla cards used by tests and the first playable slice. These are not
// real Pocket cards — they exist to exercise the core engine before the real
// card data and its "playable" filter (ADR-0002) are wired in.

func demoEmbor() Card {
	return Card{
		ID: "FIR-01", Name: "Embor", Stage: Basic, HP: 50, Type: Fire,
		Attacks:     []Attack{{Name: "Ember", Cost: []Energy{Fire}, Damage: 30}},
		Weakness:    Water,
		RetreatCost: 1,
	}
}

func demoVolt() Card {
	return Card{
		ID: "LTG-01", Name: "Volt", Stage: Basic, HP: 50, Type: Lightning,
		Attacks:     []Attack{{Name: "Spark", Cost: []Energy{Lightning}, Damage: 30}},
		Weakness:    Fighting,
		RetreatCost: 1,
	}
}

func demoEmborEX() Card {
	return Card{
		ID: "FIR-EX", Name: "Embor ex", Stage: Basic, HP: 70, Type: Fire,
		Attacks:     []Attack{{Name: "Inferno", Cost: []Energy{Fire}, Damage: 40}},
		Weakness:    Water,
		RetreatCost: 2,
		IsEX:        true,
	}
}

// deckOf builds an n-card deck of a single card. The engine does not enforce
// deck-construction rules, so this is fine for exercising play.
func deckOf(c Card, n int) []Card {
	d := make([]Card, n)
	for i := range d {
		d[i] = c
	}
	return d
}

// DemoDecks returns two ready-to-play 20-card demo decks for the first playable
// slice: a Fire deck (Embor + Embor ex) versus a Lightning deck (Volt). Each is
// single-type so the Energy Zone always feeds the printed attack cost — games
// progress reliably without real card data wired in yet (ADR-0002).
func DemoDecks() (fire, lightning []Card) {
	fire = append(deckOf(demoEmbor(), 14), deckOf(demoEmborEX(), 6)...)
	lightning = deckOf(demoVolt(), 20)
	return fire, lightning
}
