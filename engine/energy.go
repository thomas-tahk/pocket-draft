package engine

import (
	"errors"
	"sort"
)

// MaxEnergyTypes is how many energy types a deck may register (spec S2).
const MaxEnergyTypes = 3

// Deck is what a player brings to a game: its cards plus the energy types
// chosen at the final deck build (spec S2). An empty Energy means "use
// DefaultEnergyTypes(Cards)".
type Deck struct {
	Cards  []Card
	Energy []Energy
}

// energyTypes is the deck's Energy Zone types: the chosen ones, else the default.
func (d Deck) energyTypes() []Energy {
	if len(d.Energy) > 0 {
		return append([]Energy(nil), d.Energy...)
	}
	return DefaultEnergyTypes(d.Cards)
}

// DefaultEnergyTypes is a deck's default energy choice (spec S2): the types its
// attacks' costs need, most-needed first, capped at MaxEnergyTypes. Only typed
// cost entries count, so a single-type deck defaults to mono-energy. A deck
// whose attacks need no typed energy falls back to Colorless.
func DefaultEnergyTypes(deck []Card) []Energy {
	need := map[Energy]int{}
	for _, c := range deck {
		for _, a := range c.Attacks {
			for _, e := range a.Cost {
				if e != Colorless {
					need[e]++
				}
			}
		}
	}
	var out []Energy
	for _, e := range energyOrder { // fixed order keeps ties deterministic
		if need[e] > 0 {
			out = append(out, e)
		}
	}
	if len(out) == 0 {
		return []Energy{Colorless}
	}
	sort.SliceStable(out, func(i, j int) bool { return need[out[i]] > need[out[j]] })
	return out[:min(len(out), MaxEnergyTypes)]
}

// ValidateEnergyTypes checks a player's explicit choice: 1 to MaxEnergyTypes
// distinct real energy types (Grass…Metal; there is no Colorless or Dragon
// Energy Zone energy).
func ValidateEnergyTypes(types []Energy) error {
	if len(types) == 0 || len(types) > MaxEnergyTypes {
		return errors.New("choose 1 to 3 energy types")
	}
	seen := map[Energy]bool{}
	for _, e := range types {
		if e < Grass || e > Metal {
			return errors.New("not a choosable energy type: " + e.String())
		}
		if seen[e] {
			return errors.New("energy type chosen twice: " + e.String())
		}
		seen[e] = true
	}
	return nil
}
