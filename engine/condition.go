package engine

// Status is a special condition (spec C1–C8). Poison and Burn are flags of
// their own; Asleep, Paralyzed and Confused share one slot, so a Pokémon holds
// at most three at once (C3, C4).
type Status uint8

const (
	NoStatus Status = iota
	Burn
	Poison
	Asleep
	Paralyzed
	Confused
)

func (s Status) String() string {
	switch s {
	case Burn:
		return "Burned"
	case Poison:
		return "Poisoned"
	case Asleep:
		return "Asleep"
	case Paralyzed:
		return "Paralyzed"
	case Confused:
		return "Confused"
	}
	return "none"
}

// Inflict gives p condition s. Asleep, Paralyzed and Confused replace each
// other (C4); Poison and Burn stack with them (C3).
func (p *InPlay) Inflict(s Status) {
	switch s {
	case Poison:
		p.Poisoned = true
	case Burn:
		p.Burned = true
	case Asleep, Paralyzed, Confused:
		p.special = s
		p.ownTurnParalyzed = false
	}
}

// Has reports whether p has condition s.
func (p *InPlay) Has(s Status) bool {
	switch s {
	case Poison:
		return p.Poisoned
	case Burn:
		return p.Burned
	case NoStatus:
		return false
	}
	return p.special == s
}

// Cure removes condition s from p.
func (p *InPlay) Cure(s Status) {
	switch s {
	case Poison:
		p.Poisoned = false
	case Burn:
		p.Burned = false
	default:
		if p.special == s {
			p.special = NoStatus
		}
	}
}

// CureAll removes every condition from p.
func (p *InPlay) CureAll() {
	p.Poisoned, p.Burned, p.special = false, false, NoStatus
}

// Conditions lists p's conditions in Checkup order (C1): Poison, Burn, then
// Asleep, Paralyzed or Confused. Empty when it has none.
func (p *InPlay) Conditions() []Status {
	out := []Status{}
	for _, s := range []Status{Poison, Burn, Asleep, Paralyzed, Confused} {
		if p.Has(s) {
			out = append(out, s)
		}
	}
	return out
}

// cantAct reports whether p's condition stops it attacking and retreating (C5).
func (p *InPlay) cantAct() bool { return p.special == Asleep || p.special == Paralyzed }
