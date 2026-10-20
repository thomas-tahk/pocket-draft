package main

// Ability texts, part 1 of 3 (issue #77, plan 4.4). Each entry is one printed
// wording, keyed exactly as scraped; every printing with that wording plays it
// under its own name (abilityTexts in carddata.go, engine/ability.go). Passive
// wordings are rule hooks (engine/passive.go); activated ones are EffectOps;
// the rest are triggers the engine runs (engine/texts_4_4.go).

import (
	"slices"

	"github.com/thomas-tahk/pocket-draft/engine"
)

func init() {
	for text, a := range texts44 {
		abilityTexts[text] = a
	}
}

// --- conditions shared by the wordings ---

// isSelf: the passive is worked out for the Pokémon printing it.
func isSelf(c engine.PassiveCheck) bool { return c.Target.Spot == c.Source.Spot }

// selfAnd narrows isSelf.
func selfAnd(cond func(engine.PassiveCheck) bool) passiveWhen {
	return func(c engine.PassiveCheck) bool { return isSelf(c) && cond(c) }
}

func monsOf(g *engine.Game, p int) []*engine.InPlay {
	pl := &g.S.Players[p]
	var out []*engine.InPlay
	if pl.Active != nil {
		out = append(out, pl.Active)
	}
	for _, b := range pl.Bench {
		if b != nil {
			out = append(out, b)
		}
	}
	return out
}

func countNamed(g *engine.Game, p int, names ...string) int {
	n := 0
	for _, m := range monsOf(g, p) {
		if slices.Contains(names, m.Card.Name) {
			n++
		}
	}
	return n
}

// arceusLink: "If you have Arceus or Arceus ex in play".
func arceusLink(c engine.PassiveCheck) bool {
	return countNamed(c.Game, c.Source.Player, "Arceus", "Arceus ex") > 0
}

func arceusInPlay(c engine.AbilityCheck) bool {
	return countNamed(c.Game, c.Self.Player, "Arceus", "Arceus ex") > 0
}

// anotherNamed: "If you have another Beldum in play" (this one is counted too).
func anotherNamed(name string) passiveWhen {
	return selfAnd(func(c engine.PassiveCheck) bool { return countNamed(c.Game, c.Source.Player, name) >= 2 })
}

// takes: "This Pokémon takes −n damage from attacks".
func takes(n int, cond func(engine.PassiveCheck) bool) engine.Ability {
	return engine.Ability{Passive: []engine.Passive{{Hook: engine.HookDamageTaken, Value: -n, When: selfAnd(cond)}}}
}

func always(engine.PassiveCheck) bool { return true }

func foeOfType(types ...engine.Energy) func(engine.PassiveCheck) bool {
	return func(c engine.PassiveCheck) bool { return c.Foe != nil && slices.Contains(types, c.Foe.Card.Type) }
}

// evolvedNow: the Pokémon evolved this turn (a Pokémon put into play as a
// Basic is never one that evolved).
func evolvedNow(c engine.AbilityCheck) bool {
	return c.Self.PlacedTurn == c.Game.S.Turn && c.Self.Card.Stage != engine.Basic
}

func foeAsleep(c engine.AbilityCheck) bool {
	foe := c.Game.S.Players[1-c.Self.Player].Active
	return foe != nil && foe.Has(engine.Asleep)
}

func activeIsUltraBeast(c engine.AbilityCheck) bool {
	a := c.Game.S.Players[c.Self.Player].Active
	return a != nil && a.Card.IsUltraBeast()
}

// once is an activated Ability usable once during your turn.
func once(where engine.AbilitySpot, when func(engine.AbilityCheck) bool, ops ...engine.EffectOp) engine.Ability {
	return engine.Ability{Use: ops, OncePerTurn: true, Where: where, When: when}
}

var texts44 = map[string]engine.Ability{
	// --- always-on ---
	"This Pokémon can't be affected by any Special Conditions.": {Passive: []engine.Passive{{Hook: engine.HookNoConditions, When: isSelf}}},
	"If you have Arceus or Arceus ex in play, attacks used by this Pokémon cost 1 less [C] Energy.": {Passive: []engine.Passive{{Hook: engine.HookColorlessCost, Value: -1,
		When: selfAnd(arceusLink)}}},
	"If you have Arceus or Arceus ex in play, attacks used by this Pokémon do +30 damage to your opponent's Active Pokémon.": {Passive: []engine.Passive{{Hook: engine.HookDamageDealt, Value: 30,
		When: selfAnd(arceusLink)}}},
	"Attacks used by your [P] Pokémon and [M] Pokémon do +30 damage to your opponent's Active Pokémon.": {Passive: []engine.Passive{{Hook: engine.HookDamageDealt, Value: 30, When: func(c engine.PassiveCheck) bool {
		return c.Target.Player == c.Source.Player && (c.Target.Card.Type == engine.Psychic || c.Target.Card.Type == engine.Metal)
	}}}},
	"If a Stadium is in play, this Pokémon has no Retreat Cost.": {Passive: []engine.Passive{{Hook: engine.HookNoRetreatCost,
		When: selfAnd(func(c engine.PassiveCheck) bool { return c.Game.S.Stadium != nil })}}},
	"Your opponent's Active Pokémon's Retreat Cost is 1 more.": {Passive: []engine.Passive{{Hook: engine.HookRetreatCost, Value: 1, When: func(c engine.PassiveCheck) bool {
		return c.Target.Player != c.Source.Player && c.Target.Pos == 0
	}}}},
	"This Pokémon takes −30 damage from attacks.":                                                 takes(30, always),
	"This Pokémon takes −30 damage from attacks from [R] or [W] Pokémon.":                         takes(30, foeOfType(engine.Fire, engine.Water)),
	"This Pokémon takes -10 damage from attacks.":                                                 takes(10, always),
	"This Pokémon takes −10 damage from attacks.":                                                 takes(10, always),
	"This Pokémon takes −20 damage from attacks.":                                                 takes(20, always), // also Fur Coat: same wording
	"If this Pokémon has full HP, it takes −40 damage from attacks from your opponent's Pokémon.": takes(40, func(c engine.PassiveCheck) bool { return c.Target.Damage == 0 }),
	"If you have another Beldum in play, this Pokémon's Retreat Cost is 2 less.":                  {Passive: []engine.Passive{{Hook: engine.HookRetreatCost, Value: -2, When: anotherNamed("Beldum")}}},
	"As long as this Pokémon is on your Bench, your Active [D] Pokémon's Retreat Cost is 1 less.": {Passive: []engine.Passive{{Hook: engine.HookRetreatCost, Value: -1, When: func(c engine.PassiveCheck) bool {
		return c.Source.Pos > 0 && c.Target.Player == c.Source.Player && c.Target.Pos == 0 && c.Target.Card.Type == engine.Darkness
	}}}},
	"If you have another Falinks in play, this Pokémon's attacks do +20 damage to your opponent's Active Pokémon, and this Pokémon takes −20 damage from attacks from your opponent's Pokémon.": {Passive: []engine.Passive{
		{Hook: engine.HookDamageDealt, Value: 20, When: anotherNamed("Falinks")},
		{Hook: engine.HookDamageTaken, Value: -20, When: anotherNamed("Falinks")}}},
	"As long as this Pokémon is in the Active Spot, it can evolve during your first turn or the turn you play it.": {Passive: []engine.Passive{{Hook: engine.HookEvolveEarly,
		When: selfAnd(func(c engine.PassiveCheck) bool { return c.Source.Pos == 0 })}}},
	"Pokémon (both yours and your opponent's) can't be healed.": {Passive: []engine.Passive{{Hook: engine.HookNoHeal}}},
	"Each of your Pokémon that has any [P] Energy attached recovers from all Special Conditions and can't be affected by any Special Conditions.": {Passive: []engine.Passive{{Hook: engine.HookNoConditions, When: func(c engine.PassiveCheck) bool {
		return c.Target.Player == c.Source.Player && c.Target.Energy[engine.Psychic] > 0
	}}}},

	// --- once during your turn ---
	"Once during your turn, you may put a random Pokémon Tool card from your deck into your hand.":                                                                once(engine.AnySpot, nil, engine.SearchDeck{Match: ofKind(engine.KindTool)}),
	"Once during your turn, you may heal 60 damage from 1 of your Pokémon ex that has any Energy attached. If you do, discard a random Energy from that Pokémon.": once(engine.AnySpot, nil, engine.HealExDiscardEnergy{N: 60}),
	"Once during your turn, you may take a [W] Energy from your Energy Zone and attach it to the [W] Pokémon in the Active Spot.":                                 once(engine.AnySpot, nil, engine.AttachZoneToActive{Type: engine.Water}),
	"Once during your turn, you may remove a random Special Condition from your Active Pokémon.":                                                                  once(engine.AnySpot, nil, engine.CureOwnRandomCondition{}),
	"Once during your turn, if this Pokémon is on your Bench, you may move 30 damage that your Active Pokémon has on it to this Pokémon.":                         once(engine.BenchSpot, nil, engine.MoveDamageHere{N: 30}),
	"Once during your turn, you may heal 20 damage from each of your Pokémon.":                                                                                    once(engine.AnySpot, nil, engine.HealEach{Amount: 20}),
	"Once during your turn, you may switch your Active Ultra Beast with 1 of your Benched Ultra Beasts.":                                                          once(engine.AnySpot, activeIsUltraBeast, engine.SwitchUltraBeast{}),
	"Once during your turn, you may discard the top card of your opponent's deck.":                                                                                once(engine.AnySpot, nil, engine.DiscardTopOppDeck{}),
	"Once during your turn, if you have Arceus or Arceus ex in play, you may do 30 damage to your opponent's Active Pokémon.":                                     once(engine.AnySpot, arceusInPlay, engine.DamageOpp{N: 30}),
	"Once during your turn, if this Pokémon is on your Bench, you may attach an Energy from your discard pile to your Active [N] Pokémon.":                        once(engine.BenchSpot, nil, engine.AttachDiscardToActive{Type: engine.Dragon}),
	"Once during your turn, you may use this Ability. 1 Special Condition from among Burned, Confused, and Poisoned is chosen at random, and your opponent's Active Pokémon is now affected by that Special Condition. Any Special Conditions already affecting that Pokémon will not be chosen.": once(engine.AnySpot, nil, engine.RandomOppCondition{}),
	"Once during your turn, if this Pokémon is in the Active Spot, you may heal 30 damage from 1 of your Pokémon.":                               once(engine.ActiveSpot, nil, engine.HealOne{Amount: 30}),
	"Once during your turn, you may attach a [R] Energy from your discard pile to this Pokémon. If you do, do 20 damage to this Pokémon.":        once(engine.AnySpot, nil, engine.Combust{Type: engine.Fire, Damage: 20}),
	"As often as you like during your turn, you may choose 1 of your Pokémon that has damage on it, and move all of its damage to this Pokémon.": {Use: []engine.EffectOp{engine.MoveAllDamageHere{}}},

	// --- once during your turn, when you play this Pokémon to evolve ---
	"Once during your turn, when you play this Pokémon from your hand to evolve 1 of your Pokémon, you may take a [R] Energy from your Energy Zone and attach it to your Active [R] Pokémon.": once(engine.AnySpot, evolvedNow, engine.AttachZoneToActive{Type: engine.Fire}),
	"Once during your turn, when you play this Pokémon from your hand to evolve 1 of your Pokémon, you may discard a random Energy from your opponent's Active Pokémon.":                      once(engine.AnySpot, evolvedNow, engine.DiscardOppEnergy{Type: engine.NoEnergy}),
	"Once during your turn, when you play this Pokémon from your hand to evolve 1 of your Pokémon, you may put a Supporter card from your discard pile into your hand.":                       once(engine.AnySpot, evolvedNow, engine.RandomFromDiscard{Match: ofKind(engine.KindSupporter)}),
	"Once during your turn, when you play this Pokémon from your hand to evolve 1 of your Pokémon, you may do 20 damage to your opponent's Active Pokémon.":                                   once(engine.AnySpot, evolvedNow, engine.DamageOpp{N: 20}),
	"Once during your turn, when you play this Pokémon from your hand to evolve 1 of your Pokémon, you may put 2 random Pokémon Tool cards from your discard pile into your hand.":            once(engine.AnySpot, evolvedNow, engine.RandomsFromDiscard{N: 2, Match: ofKind(engine.KindTool)}),

	// --- run by the engine ---
	"At the end of your opponent's turn, if this Pokémon is in the Active Spot, put a random card from your deck that evolves from this Pokémon onto this Pokémon to evolve it.": {Where: engine.ActiveSpot, CheckupOf: engine.OppCheckup, AtCheckup: []engine.EffectOp{engine.EvolveRandomFromDeck{}}},
	"At the end of your turn, if this Pokémon is in the Active Spot, draw a card.":                                                                                               {Where: engine.ActiveSpot, CheckupOf: engine.OwnCheckup, AtCheckup: []engine.EffectOp{engine.DrawCards{N: 1}}},
	"At the end of each turn, if your opponent's Active Pokémon is Asleep, do 20 damage to that Pokémon.":                                                                        {AtCheckup: []engine.EffectOp{engine.DamageOpp{N: 20}}, When: foeAsleep},
	"If this Pokémon is in the Active Spot and is damaged by an attack from your opponent's Pokémon, the Attacking Pokémon is now Poisoned.":                                     {Where: engine.ActiveSpot, WhenDamaged: []engine.EffectOp{engine.ApplyStatus{Status: engine.Poison}}},
	"If this Pokémon is in the Active Spot and is damaged by an attack from your opponent's Pokémon, do 20 damage to the Attacking Pokémon.": { // also Steel Spikes: same wording
		Where: engine.ActiveSpot, WhenDamaged: []engine.EffectOp{engine.DamageOpp{N: 20}}},
	"Whenever you attach a [D] Energy from your Energy Zone to this Pokémon, do 20 damage to your opponent's Active Pokémon.": {OnZoneAttach: &engine.AttachTrigger{Type: engine.Darkness, Ops: []engine.EffectOp{engine.DamageOpp{N: 20}}}},
	"Whenever you attach a [P] Energy from your Energy Zone to this Pokémon, heal 20 damage from this Pokémon.":               {OnZoneAttach: &engine.AttachTrigger{Type: engine.Psychic, Ops: []engine.EffectOp{engine.HealSelf{N: 20}}}},
}
