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
	Dragon // a Pokémon type only: Dragon has no Energy Zone energy (see ValidateEnergyTypes)

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

// Attack is a single attack printed on a card: its energy cost, base damage,
// and optional effect. Vanilla cards leave Effect nil — today's base-damage
// path, unchanged.
type Attack struct {
	Name   string
	Cost   []Energy // Colorless entries are paid by any leftover energy
	Damage int
	Effect []EffectOp // nil = vanilla; runs before weakness in applyAttack
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
	IsMegaEX    bool // a Mega ex; IsEX is also true

	Ability *Ability  // a Pokémon's Ability; nil when it has none (ability.go)
	Passive []Passive // a Tool's or Stadium's always-on text (passive.go)
	// Triggers is a Tool's or Stadium's text that fires on an event, and
	// StadiumUse a Stadium's "once during each player's turn" text
	// (texts_3_13.go).
	Triggers   []Trigger
	StadiumUse []EffectOp
	// LendAttacks is a Tool that lets its Pokémon use attacks it isn't printed
	// with: given the Pokémon's card, the extra ones (Memory Light).
	LendAttacks func(Card) []Attack

	// Kind is Pokémon (the zero value) or one of the Trainer kinds. Stage,
	// HP and the rest are meaningless on a Trainer.
	Kind CardKind
	// Trainer is what a Trainer card does when played, run in order (see
	// trainer.go). Nil means its text isn't built yet, so it can't be played.
	Trainer []EffectOp
	// Fossil marks an Item played as a Basic with this card's HP (fossil.go).
	Fossil bool
}

// CardKind is what sort of card this is: a Pokémon or a Trainer of some kind.
type CardKind uint8

const (
	KindPokemon CardKind = iota
	KindItem
	KindSupporter
	KindTool
	KindStadium
)

// IsTrainer reports whether c is any kind of Trainer card.
func (c Card) IsTrainer() bool { return c.Kind != KindPokemon }

// TextBuilt reports whether a Trainer's text is built: ops to run when played,
// or always-on passives (a Tool or Stadium). An unbuilt Trainer can't be played.
func (c Card) TextBuilt() bool {
	return c.Trainer != nil || c.Passive != nil || c.Triggers != nil || c.StadiumUse != nil || c.LendAttacks != nil || c.Fossil
}

// IsBasicPokemon reports whether c is a Basic Pokémon: what setup, the
// opening hand and the Bench accept (a Trainer is never one).
func (c Card) IsBasicPokemon() bool { return c.Kind == KindPokemon && c.Stage == Basic }

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
