package engine

// Rule hooks (spec T11, C9, L6): always-on text that changes a rule while its
// card is in play — a passive Ability, a Tool, a Stadium.
//
// A card carries its passives as data (Ability.Passive for a Pokémon,
// Card.Passive for a Tool or Stadium). At each rule the engine asks every
// passive in play whose Hook names that rule; a passive applies when its When
// says so. Adding a text is one Passive value; adding a rule is one Hook
// constant and one call to sumHook or hasHook where the rule is worked out.

// Hook names the rule a passive changes.
type Hook uint8

const (
	HookDamageDealt    Hook = iota // Value is added to the damage an attack does (Target attacks, Foe defends)
	HookDamageTaken                // Value is added to attack damage Target takes, after Weakness (Foe attacks); negative reduces
	HookWeaknessDouble             // Weakness on Target's attack is ×2 instead of +20 (Foe defends)
	HookRetreatCost                // Value is added to Target's Retreat Cost (never below 0)
	HookNoRetreatCost              // Target has no Retreat Cost
	HookNoConditions               // Target can't be affected by Special Conditions
	HookEvolveEarly                // Target may evolve on its owner's first turn or the turn it was played
)

// Passive is one always-on rule change printed on a card.
type Passive struct {
	Hook  Hook
	Value int                     // the amount, for the hooks that take one
	When  func(PassiveCheck) bool // nil means always
}

// Holder is a Pokémon in play and its spot. A Stadium's holder is NoSpot with
// no Pokémon.
type Holder struct {
	Spot
	*InPlay
}

// PassiveCheck is what When is asked about: where the card printing the
// passive sits (a Tool sits with the Pokémon it is attached to), the Pokémon
// the rule is being worked out for, and, for the damage hooks, the other
// Pokémon in the attack.
type PassiveCheck struct {
	Source Holder
	Target Holder
	Foe    *InPlay
}

// Modifier is one passive in effect on a Pokémon, for the screen (spec L6):
// the name of the card it comes from, the rule and the amount.
type Modifier struct {
	Source string
	Hook   Hook
	Value  int
}

type passiveSource struct {
	at   Holder
	card Card
	list []Passive
}

// passiveSources is every card in play with passives: each Pokémon's Ability,
// both players in turn, Active first; then the Stadium.
func (g *Game) passiveSources() []passiveSource {
	var out []passiveSource
	for p := 0; p < 2; p++ {
		for _, pos := range targets() {
			at := g.holder(p, pos)
			if at.InPlay == nil {
				continue
			}
			if a := at.Card.Ability; a != nil {
				out = append(out, passiveSource{at, at.Card, a.Passive})
			}
			if tool := at.Tool; tool != nil {
				out = append(out, passiveSource{at, *tool, tool.Passive})
			}
		}
	}
	if st := g.S.Stadium; st != nil {
		out = append(out, passiveSource{Holder{NoSpot, nil}, *st, st.Passive})
	}
	return out
}

func (g *Game) holder(p, pos int) Holder {
	return Holder{Spot{p, pos}, targetMon(&g.S.Players[p], pos)}
}

// modifiers is every passive on hook h in effect for target.
func (g *Game) modifiers(h Hook, target Holder, foe *InPlay) []Modifier {
	var out []Modifier
	for _, src := range g.passiveSources() {
		for _, ps := range src.list {
			if ps.Hook != h {
				continue
			}
			if ps.When == nil || ps.When(PassiveCheck{src.at, target, foe}) {
				out = append(out, Modifier{src.card.Name, h, ps.Value})
			}
		}
	}
	return out
}

func (g *Game) sumHook(h Hook, target Holder, foe *InPlay) int {
	n := 0
	for _, m := range g.modifiers(h, target, foe) {
		n += m.Value
	}
	return n
}

func (g *Game) hasHook(h Hook, target Holder, foe *InPlay) bool {
	return len(g.modifiers(h, target, foe)) > 0
}

// Modifiers lists every passive in effect on player p's Pokémon at pos (0 the
// Active, n bench slot n-1), in hook order. Damage hooks are worked out
// against the opponent's Active.
func (g *Game) Modifiers(p, pos int) []Modifier {
	target := g.holder(p, pos)
	if target.InPlay == nil {
		return nil
	}
	foe := g.S.Players[1-p].Active
	var out []Modifier
	for h := HookDamageDealt; h <= HookEvolveEarly; h++ {
		out = append(out, g.modifiers(h, target, foe)...)
	}
	return out
}

// RetreatCost is what player p's Active costs to retreat right now.
func (g *Game) RetreatCost(p int) int {
	at := g.holder(p, 0)
	if at.InPlay == nil || g.hasHook(HookNoRetreatCost, at, nil) {
		return 0
	}
	return max(0, at.Card.RetreatCost+g.sumHook(HookRetreatCost, at, nil))
}

// attackDamage is the damage an attack's running total does to the defending
// Active (spec T11): attacker's modifiers, then Weakness, then the
// defender's. An attack that does no damage gets none of them.
func (g *Game) attackDamage(attacker int, base int) (dmg int, weakness string) {
	if base <= 0 {
		return 0, ""
	}
	att, def := g.holder(attacker, 0), g.holder(1-attacker, 0)
	dmg = base + g.sumHook(HookDamageDealt, att, def.InPlay)
	if w := def.Card.Weakness; w != NoEnergy && w == att.Card.Type {
		if g.hasHook(HookWeaknessDouble, att, def.InPlay) {
			dmg, weakness = dmg*2, "×2 weakness"
		} else {
			dmg, weakness = dmg+20, "+20 weakness"
		}
	}
	return max(0, dmg+g.sumHook(HookDamageTaken, def, att.InPlay)), weakness
}

// canBeAffected reports whether a Special Condition may be put on player p's
// Pokémon at pos (spec C9).
func (g *Game) canBeAffected(p, pos int) bool {
	return !g.hasHook(HookNoConditions, g.holder(p, pos), nil)
}
