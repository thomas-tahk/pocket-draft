package engine

import "errors"

// Effects that last a while ("during this turn", "during your opponent's next
// turn"): a played card leaves a Passive in State.Timed that applies like a
// Tool's or Stadium's until its last turn ends. Its PassiveCheck.Source is the
// player who played it, at no spot (Pos -1) and no Pokémon.

// ThisTurn: Passive applies for the rest of this turn (Through 0), or through
// your opponent's next turn as well (Through 1). E.g. X Speed, Clemont's
// Backpack, Beast Wall.
type ThisTurn struct {
	Through int
	Passive Passive
}

func (v ThisTurn) Apply(ctx *EffectContext) {
	ForTurn{Passive: v.Passive, Next: v.Through > 0}.Apply(ctx)
	ctx.g.narrate("%s takes effect.", ctx.Card.Name)
}

// OppActiveRetreatUp: until the end of your opponent's next turn, the
// Pokémon that is their Active now has a Retreat Cost N more while it stays
// Active (Team Rocket's Goo-zooka).
type OppActiveRetreatUp struct{ N int }

func (v OppActiveRetreatUp) Apply(ctx *EffectContext) {
	mon := ctx.defenderMon()
	ForTurn{Next: true, Passive: Passive{Hook: HookRetreatCost, Value: v.N, When: func(c PassiveCheck) bool {
		return c.Target.InPlay == mon && c.Target.Pos == 0
	}}}.Apply(ctx)
	ctx.g.narrate("%s's Retreat Cost is %d more until the end of %s's next turn.", mon.Card.Name, v.N, side(1-ctx.attacker))
}

// OppNoPoints: the card can be played only while your opponent has no points
// (Beast Wall).
type OppNoPoints struct{}

func (OppNoPoints) CanPlay(ctx *EffectContext) error {
	if ctx.opp().Points > 0 {
		return errors.New("you can use this card only if your opponent hasn't gotten any points")
	}
	return nil
}

func (OppNoPoints) Apply(*EffectContext) {}
