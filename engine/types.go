package engine

// Energy is one of Pocket's energy types. Colorless is the "any" type used in
// attack costs; NoEnergy is a sentinel meaning "none / not set".
type Energy uint8

const (
	Colorless Energy = iota
	Grass
	Fire
	Water
	Lightning
	Psychic
	Fighting
	Darkness
	Metal
	Dragon // a Pokémon type only: Dragon has no Energy Zone energy (see energyTypesOf)

	NoEnergy Energy = 255
)

func (e Energy) String() string {
	switch e {
	case Colorless:
		return "Colorless"
	case Grass:
		return "Grass"
	case Fire:
		return "Fire"
	case Water:
		return "Water"
	case Lightning:
		return "Lightning"
	case Psychic:
		return "Psychic"
	case Fighting:
		return "Fighting"
	case Darkness:
		return "Darkness"
	case Metal:
		return "Metal"
	case Dragon:
		return "Dragon"
	default:
		return "None"
	}
}

// energyOrder is a fixed iteration order over the real energy types. Used
// anywhere we must consume energy deterministically (replay depends on it —
// Go map iteration order is random).
var energyOrder = []Energy{Colorless, Grass, Fire, Water, Lightning, Psychic, Fighting, Darkness, Metal}

type Stage uint8

const (
	Basic Stage = iota
	Stage1
	Stage2
)

// Attack is a single attack printed on a card: its energy cost and base damage.
// Vanilla cards have only this — no special text.
type Attack struct {
	Name   string
	Cost   []Energy // Colorless entries are paid by any leftover energy
	Damage int
}

// Card is the static, printed definition of a Pokémon. The engine never
// enforces deck-construction rules (copy limits, 20-card decks) — those belong
// to the deckbuild layer. A deck is just an ordered []Card handed to the engine.
type Card struct {
	ID          string
	Name        string
	Stage       Stage
	EvolvesFrom string // pre-evolution's Name; "" for a Basic
	HP          int
	Type        Energy
	Attacks     []Attack
	Weakness    Energy // NoEnergy if the card has no weakness
	RetreatCost int
	IsEX        bool
}

// canPay reports whether the energy attached to a Pokémon (have) covers an
// attack cost. Typed costs must be matched by that exact type; Colorless costs
// are paid by any leftover energy.
func canPay(have map[Energy]int, cost []Energy) bool {
	rem := make(map[Energy]int, len(have))
	for k, v := range have {
		rem[k] = v
	}
	colorless := 0
	for _, c := range cost {
		if c == Colorless {
			colorless++
			continue
		}
		if rem[c] > 0 {
			rem[c]--
		} else {
			return false
		}
	}
	leftover := 0
	for _, n := range rem {
		leftover += n
	}
	return leftover >= colorless
}

// energyTypesOf derives a deck's registered Energy Zone types from the types of
// the Pokémon in it. Colorless-only decks fall back to Colorless.
func energyTypesOf(deck []Card) []Energy {
	seen := map[Energy]bool{}
	var out []Energy
	for _, c := range deck {
		// Colorless and Dragon have no Energy Zone energy of their own; their
		// attacks are paid by the deck's other types.
		if c.Type != Colorless && c.Type != Dragon && !seen[c.Type] {
			seen[c.Type] = true
			out = append(out, c.Type)
		}
	}
	if len(out) == 0 {
		out = []Energy{Colorless}
	}
	return out
}
