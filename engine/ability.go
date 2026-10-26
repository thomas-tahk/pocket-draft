package engine

import (
	"errors"
	"fmt"
)

// Ability is a Pokémon's Ability (spec A4). Passive is its always-on part
// (rule hooks, passive.go); Use is the part a player activates, as one
// UseAbility move.
//
// Adding an Ability text is one Ability value keyed by its printed wording in
// server's abilityTexts (registered from server/texts_<plan-id>.go, e.g.
// texts_4_2.go): Use is the effect, built from the same EffectOp verbs
// as attacks (effects.go); OncePerTurn is "Once during your turn"; Where is
// "if this Pokémon is in the Active Spot" / "on your Bench"; When is any other
// condition in the text. A new kind of effect is one new EffectOp. Inside
// Apply, ctx.self() is the Pokémon using the Ability and ctx.Card its card.
// An op may also be Gated (why it can't be used now) or Targeted (the spots
// it is aimed at, read back from ctx.Target), exactly as for Trainers.
type Ability struct {
	Name        string
	Passive     []Passive
	Use         []EffectOp              // nil: nothing to activate
	OncePerTurn bool                    // "Once during your turn"
	Where       AbilitySpot             // where it may be used from
	EvolvesAs   string                  // "can evolve into any Pokémon that evolves from <name>": also takes evolutions of that name (Eevee ex)
	When        func(AbilityCheck) bool // nil means always

	// Triggered parts, run by the engine instead of a player's move (4.4):
	AtCheckup    []EffectOp     // at the end of a turn's Checkup, if Where and When hold
	CheckupOf    CheckupTurn    // whose Checkup AtCheckup runs at
	WhenDamaged  []EffectOp     // this Pokémon, in the Active Spot, was damaged by an attack; ops see the attacker as the opponent's Active
	OnZoneAttach *AttachTrigger // an Energy from the Energy Zone is attached to this Pokémon
}

// CheckupTurn is whose turn's Checkup an Ability's AtCheckup runs at.
type CheckupTurn uint8

const (
	EveryCheckup CheckupTurn = iota // "at the end of each turn"
	OwnCheckup                      // "at the end of your turn"
	OppCheckup                      // "at the end of your opponent's turn"
)

// AttachTrigger is "whenever you attach a Type Energy from your Energy Zone to
// this Pokémon, ...".
type AttachTrigger struct {
	Type Energy
	Ops  []EffectOp
}

// AbilitySpot is where an activated Ability may be used from.
type AbilitySpot uint8

const (
	AnySpot    AbilitySpot = iota // Active or Bench
	ActiveSpot                    // "if this Pokémon is in the Active Spot"
	BenchSpot                     // "if this Pokémon is on your Bench"
)

// AbilityCheck is what an Ability's When is asked about: the Pokémon using it.
type AbilityCheck struct {
	Self Holder
	Game *Game // the game, read only, for texts that look at the board
}

// UseAbility uses the activated Ability of player p's Pokémon at Pos (0 the
// Active, n bench slot n-1). It does not end the turn.
type UseAbility struct {
	Player int
	Pos    int
	Target Spot // the spot its effect is aimed at; the zero Spot for an Ability with no target
}

func (UseAbility) event() {}

func listUseAbility(g *Game, p int) []UseAbility {
	var out []UseAbility
	for _, pos := range targets() {
		if a := g.abilityOf(g.holder(p, pos)); a != nil && a.Use != nil {
			for _, t := range abilityTargets(a, g.abilityCtx(p, pos, Spot{})) {
				out = append(out, UseAbility{Player: p, Pos: pos, Target: t})
			}
		}
	}
	return out
}

func (g *Game) abilityCtx(p, pos int, target Spot) *EffectContext {
	return &EffectContext{g: g, attacker: p, pos: pos, Card: targetMon(&g.S.Players[p], pos).Card, Target: target}
}

// abilityTargets is the union of the Ability's ops' targets, or just the
// zero Spot when no op takes one.
func abilityTargets(a *Ability, ctx *EffectContext) []Spot {
	var out []Spot
	targeted := false
	for _, op := range a.Use {
		if t, ok := op.(Targeted); ok {
			targeted = true
			out = append(out, t.Targets(ctx)...)
		}
	}
	if !targeted {
		return []Spot{{}}
	}
	return out
}

// abilityOf is the Ability of the Pokémon at h, or nil when it has none or a
// "no Abilities" passive (HookNoAbilities) turns it off.
func (g *Game) abilityOf(h Holder) *Ability {
	if h.InPlay == nil || h.Card.Ability == nil || g.noAbilitiesIn(g.allPassiveSources(), h) {
		return nil
	}
	return h.Card.Ability
}

func (g *Game) checkUseAbility(e UseAbility) error {
	if err := g.onTurn(e.Player); err != nil {
		return err
	}
	at := g.holder(e.Player, e.Pos)
	a := g.abilityOf(at)
	if a == nil || a.Use == nil {
		return errors.New("that Pokémon has no Ability to use")
	}
	if err := abilitySpotOK(a, e.Pos); err != nil {
		return err
	}
	if a.OncePerTurn && at.AbilityUsedTurn == g.S.Turn {
		return fmt.Errorf("%s was already used this turn", a.Name)
	}
	if a.When != nil && !a.When(AbilityCheck{Self: at, Game: g}) {
		return fmt.Errorf("%s can't be used right now", a.Name)
	}
	return g.abilityOpsAllow(a, e)
}

// abilityOpsAllow asks each op's gate, then checks the target is one offered.
func (g *Game) abilityOpsAllow(a *Ability, e UseAbility) error {
	ctx := g.abilityCtx(e.Player, e.Pos, Spot{})
	for _, op := range a.Use {
		if gate, ok := op.(Gated); ok {
			if err := gate.CanPlay(ctx); err != nil {
				return err
			}
		}
	}
	for _, t := range abilityTargets(a, ctx) {
		if t == e.Target {
			return nil
		}
	}
	return fmt.Errorf("%s can't be aimed there", a.Name)
}

func abilitySpotOK(a *Ability, pos int) error {
	switch {
	case a.Where == ActiveSpot && pos != 0:
		return fmt.Errorf("%s works only in the Active Spot", a.Name)
	case a.Where == BenchSpot && pos == 0:
		return fmt.Errorf("%s works only on the Bench", a.Name)
	}
	return nil
}

func (g *Game) applyUseAbility(e UseAbility) {
	at := g.holder(e.Player, e.Pos)
	a := at.Card.Ability
	at.AbilityUsedTurn = g.S.Turn
	g.narrate("%s's %s used %s.", side(e.Player), at.Card.Name, a.Name)
	ctx := g.abilityCtx(e.Player, e.Pos, e.Target)
	for _, op := range a.Use {
		op.Apply(ctx)
	}
	g.resolveAction()
}
