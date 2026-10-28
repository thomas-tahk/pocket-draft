package main

import "github.com/thomas-tahk/pocket-draft/engine"

// Ability texts, part 2 (issue #78): each printed wording maps to the engine
// Ability that plays it (engine/texts_4_5.go).

func init() {
	for wording, ability := range abilityTexts45 {
		abilityTexts[wording] = ability
	}
}

func passive(hook engine.Hook, value int, when func(engine.PassiveCheck) bool) engine.Ability {
	return engine.Ability{Passive: []engine.Passive{{Hook: hook, Value: value, When: when}}}
}

// benchedNow: the Basic was put onto the Bench from the hand this turn and is
// still there.
func benchedNow(c engine.AbilityCheck) bool {
	return c.Self.Pos > 0 && c.Self.Card.Stage == engine.Basic && c.Self.PlacedTurn == c.Game.S.Turn
}

func hasEnergy(m *engine.InPlay) bool {
	for _, n := range m.Energy {
		if n > 0 {
			return true
		}
	}
	return false
}

// itself: the passive's Pokémon is the one the rule is worked out for.
func itself(c engine.PassiveCheck) bool { return c.Source.Spot == c.Target.Spot }

// yourActiveIs: the rule applies to the Active of the player holding the Ability.
func yourActive(c engine.PassiveCheck) bool {
	return c.Target.Player == c.Source.Player && c.Target.Pos == 0
}

func yourPokemon(ok func(engine.Card) bool) func(engine.PassiveCheck) bool {
	return func(c engine.PassiveCheck) bool { return c.Target.Player == c.Source.Player && ok(c.Target.Card) }
}

func typeIs(t engine.Energy) func(engine.Card) bool {
	return func(c engine.Card) bool { return c.Type == t }
}

var abilityTexts45 = map[string]engine.Ability{
	"Once during your turn, you may take 1 [P] Energy from your Energy Zone and attach it to the [P] Pokémon in the Active Spot.":                                      once(engine.AnySpot, nil, engine.AttachZoneToActive{Type: engine.Psychic}),
	"Once during your turn, you may take a [P] Energy from your Energy Zone and attach it to this Pokémon. If you use this Ability, your turn ends.":                   once(engine.AnySpot, nil, engine.ZoneToSelf{Energy: engine.Psychic, N: 1}, engine.EndTurnNow{}),
	"Once during your turn, you may take a [L] Energy from your Energy Zone and attach it to this Pokémon.":                                                            once(engine.AnySpot, nil, engine.ZoneToSelf{Energy: engine.Lightning, N: 1}),
	"Once during your turn, you may take a [W] Energy from your Energy Zone and attach it to this Pokémon.":                                                            once(engine.AnySpot, nil, engine.ZoneToSelf{Energy: engine.Water, N: 1}),
	"Once during your turn, you may take 2 [D] Energy from your Energy Zone and attach it to this Pokémon. If you do, do 30 damage to this Pokémon.":                   once(engine.AnySpot, nil, engine.ZoneToSelf{Energy: engine.Darkness, N: 2}, engine.DamageSelf{N: 30}),
	"Once during your turn, if this Pokémon is in the Active Spot, you may take a [G] Energy from your Energy Zone and attach it to 1 of your [G] Pokémon.":            once(engine.ActiveSpot, nil, engine.AttachFromZone{Which: engine.OfType(engine.Grass), Type: engine.Grass, N: 1}),
	"Once during your turn, you may flip a coin. If heads, your opponent's Active Pokémon is now Poisoned.":                                                            once(engine.AnySpot, nil, engine.OnHeads{Op: engine.ApplyStatus{Status: engine.Poison}}),
	"Once during your turn, you may flip a coin. If heads, your opponent's Active Pokémon is now Asleep.":                                                              once(engine.AnySpot, nil, engine.OnHeads{Op: engine.ApplyStatus{Status: engine.Asleep}}),
	"Once during your turn, if this Pokémon is in the Active Spot, you may make your opponent's Active Pokémon Confused.":                                              once(engine.ActiveSpot, nil, engine.ApplyStatus{Status: engine.Confused}),
	"Once during your turn, you may do 20 damage to 1 of your opponent's Pokémon.":                                                                                     once(engine.AnySpot, nil, engine.DamageOneOpp{N: 20}),
	"Once during your turn, you may switch your Active [W] Pokémon with 1 of your Benched Pokémon.":                                                                    once(engine.AnySpot, nil, engine.SwitchOwnActive{Fits: typeIs(engine.Water)}),
	"Once during your turn, you may heal 20 damage from your Active Pokémon.":                                                                                          once(engine.AnySpot, nil, engine.HealActive{N: 20}),
	"Once during your turn, you may heal 30 damage from each of your [W] Pokémon.":                                                                                     once(engine.AnySpot, nil, engine.HealEach{Amount: 30, Which: engine.OfType(engine.Water)}),
	"Once during your turn, you may move all [P] Energy from 1 of your Benched [P] Pokémon to your Active Pokémon.":                                                    once(engine.AnySpot, nil, engine.MoveTypeEnergyToActive{Energy: engine.Psychic}),
	"Once during your turn, you may switch out your opponent's Active Pokémon to the Bench. (Your opponent chooses the new Active Pokémon.)":                           once(engine.AnySpot, nil, engine.SwitchOppActive{}),
	"Once during your turn, if this Pokémon is on your Bench, you may discard all Pokémon Tools from your opponent's Active Pokémon. If you do, discard this Pokémon.": once(engine.BenchSpot, nil, engine.DiscardOppActiveTools{}, engine.DiscardSelf{}),
	"Once during your turn, you may look at the top card of your deck.":                                                                                                once(engine.AnySpot, nil, engine.LookAtTop{N: 1}),

	// --- when you evolve it (4.4's evolvedNow) ---
	"Once during your turn, when you play this Pokémon from your hand to evolve 1 of your Pokémon, you may do 30 damage to your opponent's Active Pokémon.":                                                                                    once(engine.AnySpot, evolvedNow, engine.DamageOpp{N: 30}),
	"Once during your turn, when you play this Pokémon from your hand to evolve 1 of your Pokémon, you may heal 60 damage from 1 of your [W] Pokémon.":                                                                                         once(engine.AnySpot, evolvedNow, engine.HealOne{Amount: 60, Which: engine.OfType(engine.Water)}),
	"Once during your turn, when you play this Pokémon from your hand to evolve 1 of your Pokémon, you may have your opponent shuffle their hand into their deck. For each remaining point that your opponent needs to win, they draw a card.": once(engine.AnySpot, evolvedNow, engine.OppRedrawPointsLeft{}),
	"Once during your turn, when you play this Pokémon from your hand to evolve 1 of your Pokémon, you may flip a coin. If heads, your opponent's Active Pokémon is now Paralyzed.":                                                            once(engine.AnySpot, evolvedNow, engine.OnHeads{Op: engine.ApplyStatus{Status: engine.Paralyzed}}),

	// --- when you put it onto your Bench from your hand (benchedNow) ---
	"Once during your turn, when you put this Pokémon from your hand onto your Bench, you may switch it with your Active Pokémon. If you do, move all of your Energy in play to this Pokémon.": once(engine.BenchSpot, benchedNow, engine.SwitchInTakeAllEnergy{}),
	"Once during your turn, when you put this Pokémon from your hand onto your Bench, you may have your opponent reveal their hand.":                                                           once(engine.BenchSpot, benchedNow, engine.RevealOppHand{}),
	"Once during your turn, when you put this Pokémon from your hand onto your Bench, you may heal 20 damage from your Active [G] Pokémon.": once(engine.BenchSpot, benchedNow,
		engine.Require{Test: func(ctx *engine.EffectContext) bool { return ctx.Player().Active.Card.Type == engine.Grass }, Why: "your Active isn't a Grass Pokémon"}, engine.HealActive{N: 20}),

	// --- during Pokémon Checkup ---
	"During Pokémon Checkup, heal 10 damage from each of your Pokémon.":                                              {CheckupOf: engine.EveryCheckup, AtCheckup: []engine.EffectOp{engine.HealEach{Amount: 10}}},
	"During Pokémon Checkup, if this Pokémon is in the Active Spot, do 10 damage to your opponent's Active Pokémon.": {Where: engine.ActiveSpot, CheckupOf: engine.EveryCheckup, AtCheckup: []engine.EffectOp{engine.DamageOpp{N: 10}}},

	// --- retreat with a friend in play ---
	"If you have Arceus or Arceus ex in play, this Pokémon has no Retreat Cost.": passive(engine.HookNoRetreatCost, 0, func(c engine.PassiveCheck) bool { return itself(c) && arceusLink(c) }),
	"If you have Latias in play, this Pokémon has no Retreat Cost.": passive(engine.HookNoRetreatCost, 0, func(c engine.PassiveCheck) bool {
		return itself(c) && countNamed(c.Game, c.Source.Player, "Latias") > 0
	}),

	"If this Pokémon has any Energy attached, it has no Retreat Cost.": passive(engine.HookNoRetreatCost, 0,
		func(c engine.PassiveCheck) bool { return itself(c) && hasEnergy(c.Target.InPlay) }),
	"Your Active Pokémon has no Retreat Cost.":    passive(engine.HookNoRetreatCost, 0, yourActive),
	"This Pokémon takes -20 damage from attacks.": passive(engine.HookDamageTaken, -20, itself),
	"This Pokémon takes −20 damage from attacks from [R] or [W] Pokémon.": passive(engine.HookDamageTaken, -20,
		func(c engine.PassiveCheck) bool {
			return itself(c) && c.Foe != nil && (c.Foe.Card.Type == engine.Fire || c.Foe.Card.Type == engine.Water)
		}),
	"As long as this Pokémon is in the Active Spot, attacks used by your opponent's Active Pokémon do −20 damage.": passive(engine.HookDamageDealt, -20,
		func(c engine.PassiveCheck) bool { return c.Source.Pos == 0 && c.Target.Player != c.Source.Player }),
	"Attacks used by your [F] Pokémon do +20 damage to your opponent's Active Pokémon.": passive(engine.HookDamageDealt, 20,
		yourPokemon(typeIs(engine.Fighting))),
	"Attacks used by your Future Pokémon cost 1 less [C] Energy.": passive(engine.HookColorlessCost, -1,
		yourPokemon(isFuture)),
	"If this Pokémon's remaining HP is 50 or less, attacks used by this Pokémon do +60 damage to your opponent's Active Pokémon.": passive(engine.HookDamageDealt, 60,
		func(c engine.PassiveCheck) bool { return itself(c) && c.Target.Card.HP-c.Target.Damage <= 50 }),
	"As long as this Pokémon is on your Bench, attacks used by your Pokémon that evolve from Poliwhirl do +40 damage to your opponent's Active Pokémon.": passive(engine.HookDamageDealt, 40,
		func(c engine.PassiveCheck) bool {
			return c.Source.Pos > 0 && yourPokemon(func(k engine.Card) bool { return k.EvolvesFrom == "Poliwhirl" })(c)
		}),
	"Prevent all damage done to this Pokémon by attacks from your opponent's Pokémon ex.": passive(engine.HookDamageTaken, -999,
		func(c engine.PassiveCheck) bool { return itself(c) && c.Foe != nil && c.Foe.Card.IsEX }),
}
