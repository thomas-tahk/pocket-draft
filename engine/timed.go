package engine

// Timed passives: a rule change a played card puts in effect "during this
// turn" or "during your opponent's next turn" (Giovanni, Blue). It is a
// Passive like a Tool's (passive.go) whose source is the card's player, not a
// spot: When reads that player as Source.Player. It lasts while the turn
// counter is at most Until.

// TimedPassive is one passive in effect until the end of turn Until.
type TimedPassive struct {
	Card    Card // the card that put it in effect, named on the screen
	Owner   int  // the player who played it
	Until   int  // the last turn it applies on
	Passive Passive
}

// timedSources is every timed passive still in effect, as passive sources.
func (g *Game) timedSources() []passiveSource {
	var out []passiveSource
	for _, t := range g.S.Timed {
		if g.S.Turn <= t.Until {
			at := Holder{Spot{t.Owner, -1}, nil}
			out = append(out, passiveSource{at, t.Card, []Passive{t.Passive}, false})
		}
	}
	return out
}

// ForTurn puts Passive in effect for the rest of this turn, or with Next
// through the opponent's next turn as well.
type ForTurn struct {
	Passive Passive
	Next    bool
}

func (v ForTurn) Apply(ctx *EffectContext) {
	s := ctx.g.S
	until := s.Turn
	if v.Next {
		until++
	}
	live := s.Timed[:0:0]
	for _, t := range s.Timed {
		if s.Turn <= t.Until {
			live = append(live, t)
		}
	}
	s.Timed = append(live, TimedPassive{ctx.Card, ctx.attacker, until, v.Passive})
}

// attackCost is attack i of player p's Active's cost after HookColorlessCost:
// each point of reduction drops one [C] (never below none).
func (g *Game) attackCost(p, i int) []Energy {
	at := g.holder(p, 0)
	cost := at.Attacks()[i].Cost
	cut := -g.sumHook(HookColorlessCost, at, nil)
	if cut <= 0 {
		return cost
	}
	out := make([]Energy, 0, len(cost))
	for _, e := range cost {
		if e == Colorless && cut > 0 {
			cut--
			continue
		}
		out = append(out, e)
	}
	return out
}

// surviveKO leaves the defending Active at 10 HP when attack damage would
// knock it out and HookSurviveKO protects it (Hala).
func (g *Game) surviveKO(attacker int) {
	def := g.holder(1-attacker, 0)
	if g.hpLeft(1-attacker, 0) > 0 || !g.hasHook(HookSurviveKO, def, g.S.Players[attacker].Active) {
		return
	}
	def.Damage = g.maxHP(def) - 10
	g.narrate("%s's %s was not Knocked Out: its HP is now 10.", side(1-attacker), def.Card.Name)
}
