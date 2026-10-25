package main

// Ability texts, part 3 of 3 (issue #79, plan 4.6). Each entry is one printed
// wording, keyed exactly as scraped; the ops it is built from are in
// engine/texts_4_6.go. Wordings the engine can't yet express are left out and
// recorded under [4.6] in docs/questions/M2.md.

import (
	"slices"

	"github.com/thomas-tahk/pocket-draft/engine"
)

func init() {
	for text, a := range texts46 {
		abilityTexts[text] = a
	}
}

// energyCount scales a passive by the Energy of one type on its Pokémon.
func energyCount(t engine.Energy) func(engine.PassiveCheck) int {
	return func(c engine.PassiveCheck) int { return c.Target.Energy[t] }
}

// unownOther: an Unown of p's is in play with an Ability not named name.
func unownOther(g *engine.Game, p int, name string) bool {
	return slices.ContainsFunc(monsOf(g, p), func(m *engine.InPlay) bool {
		return m.Card.Name == "Unown" && m.Card.Ability != nil && m.Card.Ability.Name != name
	})
}

// energyTotal is the Energy attached to a Pokémon, of every type.
func energyTotal(m *engine.InPlay) int {
	n := 0
	for _, v := range m.Energy {
		n += v
	}
	return n
}

func hasTool(c engine.AbilityCheck) bool { return c.Self.Tool != nil }

func placedBasicNow(c engine.AbilityCheck) bool {
	return c.Self.PlacedTurn == c.Game.S.Turn && c.Self.Card.Stage == engine.Basic
}

func isItem(c engine.Card) bool { return c.Kind == engine.KindItem }

var texts46 = map[string]engine.Ability{
	// --- always-on ---
	"If you have Arceus or Arceus ex in play, this Pokémon takes −30 damage from attacks.": takes(30, arceusLink),
	"Prevent all effects of attacks used by your opponent's Pokémon done to this Pokémon.": {Passive: []engine.Passive{{Hook: engine.HookEffectShield, When: isSelf}}},
	"This Pokémon gets +30 HP for each [P] Energy attached to it.": {Passive: []engine.Passive{{Hook: engine.HookMaxHP, Value: 30,
		When: isSelf, Scale: energyCount(engine.Psychic)}}},
	"This Pokémon gets +30 HP for each [G] Energy attached to it.": {Passive: []engine.Passive{{Hook: engine.HookMaxHP, Value: 30,
		When: isSelf, Scale: energyCount(engine.Grass)}}},
	"As long as this Pokémon is on your Bench, prevent all damage done to this Pokémon by attacks.": {Passive: []engine.Passive{{Hook: engine.HookBenchShield, When: isSelf}}},
	"Your Active Dondozo has no Retreat Cost.": {Passive: []engine.Passive{{Hook: engine.HookNoRetreatCost, When: func(c engine.PassiveCheck) bool {
		return c.Target.Player == c.Source.Player && c.Target.Pos == 0 && c.Target.Card.Name == "Dondozo"
	}}}},
	"As long as this Pokémon is on your Bench, your Active Basic Pokemon's Retreat Cost is 1 less.": {Passive: []engine.Passive{{Hook: engine.HookRetreatCost, Value: -1,
		When: func(c engine.PassiveCheck) bool {
			return c.Source.Pos > 0 && c.Target.Player == c.Source.Player && c.Target.Pos == 0 && c.Target.Card.Stage == engine.Basic
		}}}},
	"This Pokémon takes −30 damage from attacks from [F] Pokémon.": takes(30, foeOfType(engine.Fighting)),
	"As long as this Pokémon is in the Active Spot, attacks used by your opponent's Active Pokémon cost 1 [C] more.": {Passive: []engine.Passive{{Hook: engine.HookColorlessCost, Value: 1,
		When: func(c engine.PassiveCheck) bool {
			return c.Source.Pos == 0 && c.Target.Player != c.Source.Player && c.Target.Pos == 0
		}}}},
	"Each of your Pokémon that has any Energy attached recovers from all Special Conditions and can't be affected by any Special Conditions.": {Passive: []engine.Passive{{Hook: engine.HookNoConditions, When: func(c engine.PassiveCheck) bool {
		return c.Target.Player == c.Source.Player && energyTotal(c.Target.InPlay) > 0
	}}}},
	"This Ability works if you have any Unown in play with an Ability other than GUARD. All of your Pokémon take −10 damage from attacks from your opponent's Pokémon.": {Passive: []engine.Passive{{Hook: engine.HookDamageTaken, Value: -10,
		When: func(c engine.PassiveCheck) bool {
			return c.Target.Player == c.Source.Player && unownOther(c.Game, c.Source.Player, "GUARD")
		}}}},
	"This Ability works if you have any Unown in play with an Ability other than POWER. Attacks used by your Pokémon do +10 damage to your opponent's Active Pokémon.": {Passive: []engine.Passive{{Hook: engine.HookDamageDealt, Value: 10,
		When: func(c engine.PassiveCheck) bool {
			return c.Target.Player == c.Source.Player && unownOther(c.Game, c.Source.Player, "POWER")
		}}}},
	"During your first turn, this Pokémon has no Retreat Cost.": {Passive: []engine.Passive{{Hook: engine.HookNoRetreatCost,
		When: selfAnd(func(c engine.PassiveCheck) bool { return firstTurnOwn(c.Game, c.Source.Player) })}}},

	// --- run at the end of your turn ---
	"At the end of your turn, if this Pokémon is in the Active Spot, heal 20 damage from it.": {Where: engine.ActiveSpot, CheckupOf: engine.OwnCheckup,
		AtCheckup: []engine.EffectOp{engine.HealSelf{N: 20}}},
	"At the end of your first turn, take a [L] Energy from your Energy Zone and attach it to this Pokémon.": {CheckupOf: engine.OwnCheckup,
		When:      func(c engine.AbilityCheck) bool { return firstTurnOwn(c.Game, c.Self.Player) },
		AtCheckup: []engine.EffectOp{engine.AttachZoneToSelf{Type: engine.Lightning}}},

	// --- once during your turn, when you play this Pokémon to evolve it ---
	"Once during your turn, when you play this Pokémon from your hand to evolve 1 of your Pokémon, you may look at the top 4 cards of your deck and put all Item cards you find there into your hand. Shuffle the other cards back into your deck.": once(engine.AnySpot, evolvedNow, engine.TopCardsToHand{N: 4, Fits: isItem}),
	"Once during your turn, when you play this Pokémon from your hand to evolve 1 of your Pokémon, you may draw 2 cards.":                                                                                                                           once(engine.AnySpot, evolvedNow, engine.DrawCards{N: 2}),
	"Once during your turn, when you play this Pokémon from your hand to evolve 1 of your Pokémon, you may move a random Energy from your opponent's Active Pokémon to this Pokémon.":                                                               once(engine.AnySpot, evolvedNow, engine.MoveFoeEnergyHere{}),
	"Once during your turn, when you play this Pokémon from your hand to evolve 1 of your Pokémon, you may make your opponent's Active Pokémon Poisoned and Burned.": once(engine.AnySpot, evolvedNow,
		engine.ApplyStatus{Status: engine.Poison}, engine.ApplyStatus{Status: engine.Burn}),
	"Once during your turn, when you play this Pokémon from your hand to evolve 1 of your Pokémon, you may prevent all damage from—and effects of—attacks from your opponent's Pokémon done to this Pokémon until the end of your opponent's next turn.": once(engine.AnySpot, evolvedNow, engine.ShieldSelfNextTurn{}),

	// --- once during your turn ---
	"Once during your turn, you may flip a coin. If heads, switch in 1 of your opponent's Benched Pokémon to the Active Spot.":                                                                         once(engine.AnySpot, nil, engine.FlipGust{}),
	"Once during your turn, you may switch your Active [M] Pokémon with 1 of your Benched Pokémon.":                                                                                                    once(engine.AnySpot, nil, engine.SwitchTypedActive{Type: engine.Metal}),
	"Once during your turn, when you put this Pokémon from your hand onto your Bench, you may switch out your opponent's Active Pokémon to the Bench. (Your opponent chooses the new Active Pokémon.)": once(engine.BenchSpot, placedBasicNow, engine.SwitchOppActive{}),
	"Once during your turn, you may heal 10 damage from each of your Pokémon.":                                                                                                                         once(engine.AnySpot, nil, engine.HealEach{Amount: 10}),
	"Once during your turn, you may put a random Pokémon from your deck into your hand.":                                                                                                               once(engine.AnySpot, nil, engine.SearchDeck{Match: ofKind(engine.KindPokemon)}),
	"Once during your turn, you may switch out your opponent's Active Basic Pokémon to the Bench. (Your opponent chooses the new Active Pokémon.)":                                                     once(engine.AnySpot, nil, engine.SwitchOppActiveBasic{}),
	"Once during your turn, if this Pokémon has a Pokémon Tool attached, you may heal 30 damage from 1 of your Pokémon.":                                                                               once(engine.AnySpot, hasTool, engine.HealOne{Amount: 30}),
	"Once during your turn, if this Pokémon is in the Active Spot, you may draw a card.":                                                                                                               once(engine.ActiveSpot, nil, engine.DrawCards{N: 1}),
	"Once during your turn, if this Pokémon is in the Active Spot, you may switch in 1 of your opponent's Benched Pokémon that has damage on it to the Active Spot.":                                   once(engine.ActiveSpot, nil, engine.SwitchInDamagedBench{}),
	"Once during your turn, if this Pokémon is in the Active Spot, you may switch out your opponent's Active Pokémon to the Bench. (Your opponent chooses the new Active Pokémon.)":                    once(engine.ActiveSpot, nil, engine.SwitchOppActive{}),
	"If this Pokémon is in the Active Spot, once during your turn, you may switch in 1 of your opponent's Benched Basic Pokémon to the Active Spot.":                                                   once(engine.ActiveSpot, nil, engine.SwitchInOppBasic{}),
	"Once during your turn, if this Pokémon is in the Active Spot, you may make your opponent's Active Pokémon Poisoned.":                                                                              once(engine.ActiveSpot, nil, engine.ApplyStatus{Status: engine.Poison}),
	"Once during your turn, you may make your opponent's Active Pokémon Burned.":                                                                                                                       once(engine.AnySpot, nil, engine.ApplyStatus{Status: engine.Burn}),
	"Once during your turn, you may heal 20 damage from your Active Pokémon.":                                                                                                                          once(engine.AnySpot, nil, engine.HealActive{N: 20}),
	"Once during your turn, if this Pokémon is on your Bench, you may switch it with your Active Pokémon.":                                                                                             once(engine.BenchSpot, nil, engine.SwitchSelfActive{}),
	"Once during your turn, you may move all [D] Energy from each of your Pokémon to this Pokémon.":                                                                                                    once(engine.AnySpot, nil, engine.MoveTypeEnergyHere{Type: engine.Darkness}),
	"Once during your turn, you may discard 1 [R] Energy from this Pokémon in order to use this Ability. During this turn, attacks used by your [R] Pokémon do +50 damage to your opponent's Active Pokémon.": once(engine.AnySpot, nil,
		engine.DiscardOwnEnergy{Type: engine.Fire}, engine.ThisTurn{Passive: engine.Passive{Hook: engine.HookDamageDealt, Value: 50, When: func(c engine.PassiveCheck) bool {
			return c.Target.Player == c.Source.Player && c.Target.Card.Type == engine.Fire
		}}}),

	// --- used as often as you like ---
	"As often as you like during your turn, you may move a [W] Energy from 1 of your Benched [W] Pokémon to your Active [W] Pokémon.": {Use: []engine.EffectOp{engine.MoveOneTypeEnergyToActive{Type: engine.Water}}},
}

// firstTurnOwn reports whether p is on its first turn.
func firstTurnOwn(g *engine.Game, p int) bool { return g.S.Turn <= 2 && g.S.Active == p }
