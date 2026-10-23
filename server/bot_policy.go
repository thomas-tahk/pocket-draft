package main

import "github.com/thomas-tahk/pocket-draft/engine"

// The bot's policy: simple rules, no lookahead. It exists as a practice
// opponent (docs/VISION.md), so it only has to play a real game.

// Rank bands, highest first. Every move that does not end the turn ranks above
// attacking, so the bot builds its board and then swings.
const (
	rankRetreat   = 90 // minus Energy still missing on the Pokémon it brings in
	rankEvolve    = 80
	rankBench     = 70
	rankAbility   = 60
	rankTrainer   = 50
	rankStadium   = 48
	rankAttachFor = 44 // a Benched Pokémon that still needs Energy; minus what it lacks
	rankAttachAct = 46 // the Active, while its attack is unpaid
	rankAttachAny = 30 // Energy has nowhere better to go
	rankAttack    = 20 // plus damage/10, capped
	rankEndTurn   = 10
	rankPromote   = 100 // minus Energy still missing; the bot picks who steps up
)

// unreachable is what missingEnergy reports for a Pokémon that can never pay
// for an attack with the Energy its owner's zone makes.
const unreachable = 99

// moveRank is the bot's policy. At setup: place the Active, fill the bench,
// press Ready. In play: retreat a stuck Active, evolve, bench, use Abilities
// and Trainers, attach Energy, attack, otherwise pass. Prompt answers are the
// only moves listed during a prompt, so they are taken whatever their kind.
// Concede is never ok, nor are moves the bot has no policy for.
func moveRank(g *engine.Game, ev engine.Event) (int, bool) {
	switch e := ev.(type) {
	case engine.Concede, engine.SetupReturn, engine.SetupMove, engine.DiscardFossil:
		return 0, false
	case engine.SetupActive:
		return 3, true
	case engine.SetupBench:
		return 2, true
	case engine.SetupReady:
		return 1, true
	case engine.Evolve:
		return rankEvolve, true
	case engine.PlayBasic:
		return rankBench, true
	case engine.PlayTrainer:
		return trainerRank(g, e)
	case engine.UseStadium:
		return rankStadium, true
	case engine.UseAbility:
		return rankAbility, !abilityUsedThisTurn(g, e)
	case engine.AttachEnergy:
		return attachRank(g, e)
	case engine.Retreat:
		return retreatRank(g, e)
	case engine.ChooseNewActive:
		return rankPromote - missingEnergy(g, e.Player, benchMon(g, e.Player, e.BenchIndex)), true
	case engine.UseAttack:
		return attackRank(g, e), true
	case engine.EndTurn:
		return rankEndTurn, true
	}
	return 0, true
}

func benchMon(g *engine.Game, player, slot int) *engine.InPlay {
	return g.S.Players[player].Bench[slot]
}

// target is the Pokémon an Active/Bench target number names (0 Active, n slot n-1).
func target(g *engine.Game, player, n int) *engine.InPlay {
	if n == 0 {
		return g.S.Players[player].Active
	}
	return benchMon(g, player, n-1)
}

// attachRank puts Energy on the Active while its attack is unpaid, else on the
// Benched Pokémon closest to attacking, else back on the Active.
func attachRank(g *engine.Game, e engine.AttachEnergy) (int, bool) {
	missing := missingEnergy(g, e.Player, target(g, e.Player, e.Target))
	needs := missing > 0 && missing != unreachable
	switch {
	case e.Target == 0 && needs:
		return rankAttachAct, true
	case e.Target != 0 && needs:
		return rankAttachFor - missing, true
	}
	return rankAttachAny, true
}

// retreatRank retreats only an Active that can never attack, into a Benched
// Pokémon that can. The engine lists a retreat only when it can be paid.
func retreatRank(g *engine.Game, e engine.Retreat) (int, bool) {
	stuck := missingEnergy(g, e.Player, g.S.Players[e.Player].Active) == unreachable
	missing := missingEnergy(g, e.Player, benchMon(g, e.Player, e.BenchIndex))
	if !stuck || missing == unreachable {
		return 0, false
	}
	return rankRetreat - missing, true
}

// attackRank prefers the bigger hit. An index past the printed attacks is one
// a Tool lent the Pokémon.
func attackRank(g *engine.Game, e engine.UseAttack) int {
	attacks := g.S.Players[e.Player].Active.Card.Attacks
	if e.Index >= len(attacks) {
		return rankAttack
	}
	return rankAttack + min(attacks[e.Index].Damage/10, 9)
}

// missingEnergy is how many Energy the Pokémon still needs for the attack
// closest to paid, counting only attacks its owner's Energy Zone can ever
// pay: unreachable when none can.
func missingEnergy(g *engine.Game, player int, mon *engine.InPlay) int {
	if mon == nil {
		return unreachable
	}
	zone := g.S.Players[player].EnergyTypes
	best := unreachable
	for _, a := range mon.Card.Attacks {
		if m, ok := missingFor(a.Cost, mon.Energy, zone); ok {
			best = min(best, m)
		}
	}
	return best
}

// missingFor is the Energy still lacking for one cost; false when a typed
// entry is neither attached nor in the zone.
func missingFor(cost []engine.Energy, have map[engine.Energy]int, zone []engine.Energy) (int, bool) {
	left := map[engine.Energy]int{}
	spare := 0
	for t, n := range have {
		left[t], spare = n, spare+n
	}
	missing := 0
	for _, c := range cost {
		switch {
		case c == engine.Colorless:
		case left[c] > 0:
			left[c]--
			spare--
		case inZone(zone, c):
			missing++
		default:
			return 0, false
		}
	}
	for _, c := range cost {
		if c != engine.Colorless {
			continue
		}
		if spare > 0 {
			spare--
		} else {
			missing++
		}
	}
	return missing, true
}

func inZone(zone []engine.Energy, e engine.Energy) bool {
	for _, z := range zone {
		if z == e {
			return true
		}
	}
	return false
}

// abilityUsedThisTurn is true once this Pokémon's Ability was used since the
// turn began, so the bot never repeats it. The turn began at the last EndTurn
// or attack.
func abilityUsedThisTurn(g *engine.Game, e engine.UseAbility) bool {
	for i := len(g.Events) - 1; i >= 0; i-- {
		switch prev := g.Events[i].(type) {
		case engine.EndTurn, engine.UseAttack:
			return false
		case engine.UseAbility:
			if prev.Player == e.Player && prev.Pos == e.Pos {
				return true
			}
		}
	}
	return false
}
