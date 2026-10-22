package main

import (
	"strings"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// Item texts, batch 3.12 (issue #72). Each wording is registered into
// trainerEffects exactly as scraped; the ops live in engine/items.go,
// engine/choice.go and engine/turneffect.go.

// ultraBeasts names every Ultra Beast in the data; the scrape has no flag for
// them, and an "ex" printing is the same Ultra Beast.
var ultraBeasts = map[string]bool{
	"Nihilego": true, "Buzzwole": true, "Pheromosa": true, "Xurkitree": true,
	"Celesteela": true, "Kartana": true, "Guzzlord": true, "Poipole": true,
	"Naganadel": true, "Stakataka": true, "Blacephalon": true,
}

func isUltraBeast(c engine.Card) bool { return ultraBeasts[strings.TrimSuffix(c.Name, " ex")] }

// basicOf names the Basic a Stage 1 evolves from, from the card data. It is
// looked up when played, since the data loads after init.
func basicOf(stage1 string) string {
	for _, rc := range rawByID {
		if rc.Name == stage1 && rc.EvolvesFrom != "" {
			return rc.EvolvesFrom
		}
	}
	return ""
}

func evolvesFromEevee(c engine.Card) bool {
	return c.Kind == engine.KindPokemon && c.EvolvesFrom == "Eevee"
}

// ownedBy: the passive applies to the Pokémon of the player who played it.
func ownedBy(c engine.PassiveCheck) bool { return c.Target.Player == c.Source.Player }

func ownedNamed(names ...string) func(engine.PassiveCheck) bool {
	return func(c engine.PassiveCheck) bool {
		if !ownedBy(c) {
			return false
		}
		for _, n := range names {
			if c.Target.Card.Name == n {
				return true
			}
		}
		return false
	}
}

var itemTexts = map[string][]engine.EffectOp{
	"You can use this card only if your opponent hasn't gotten any points.During your opponent's next turn, all of your Ultra Beasts take −20 damage from attacks from your opponent's Pokémon.": {
		engine.OppNoPoints{},
		engine.ThisTurn{Through: 1, Passive: engine.Passive{Hook: engine.HookDamageTaken, Value: -20,
			When: func(c engine.PassiveCheck) bool { return ownedBy(c) && isUltraBeast(c.Target.Card) }}},
	}, // Beast Wall
	"Heal 10 damage and remove a random Special Condition from your Active Pokémon.": {
		engine.HealActive{N: 10}, engine.CureRandomCondition{},
	}, // Big Malasada
	"During this turn, attacks used by your Magneton or Heliolisk do +20 damage to your opponent's Pokémon.": {
		engine.ThisTurn{Passive: engine.Passive{Hook: engine.HookDamageDealt, Value: 20, When: ownedNamed("Magneton", "Heliolisk")}},
	}, // Clemont's Backpack
	"Choose 1:① During this turn, attacks used by your Pokémon that evolve from Eevee do +10 damage to your opponent's Active Pokémon.② Heal 20 damage from each of your Pokémon that evolves from Eevee.": {
		engine.ChooseOne{
			Labels: []string{"+10 damage this turn", "Heal 20 damage"},
			Ops: [][]engine.EffectOp{
				{engine.ThisTurn{Passive: engine.Passive{Hook: engine.HookDamageDealt, Value: 10,
					When: func(c engine.PassiveCheck) bool { return ownedBy(c) && evolvesFromEevee(c.Target.Card) }}}},
				{engine.HealEach{Amount: 20, Which: func(m *engine.InPlay) bool { return evolvesFromEevee(m.Card) }}},
			},
		},
	}, // Eevee Bag
	"Flip a coin. If heads, take a [L] Energy from your Energy Zone and attach it to 1 of your Benched [L] Pokémon.": {
		engine.FlipForBenchEnergy{Type: engine.Lightning},
	}, // Electric Generator
	"Move a [R], [W], or [L] Energy from 1 of your Benched Pokémon to your Active Pokémon.": {
		engine.MoveEnergyToActive{Types: []engine.Energy{engine.Fire, engine.Water, engine.Lightning}},
	}, // Elemental Switch
	"Put a random Basic [W] Pokémon from your discard pile into your hand.": {
		engine.RandomFromDiscard{Match: func(c engine.Card) bool { return c.IsBasicPokemon() && c.Type == engine.Water }},
	}, // Fishing Net
	"Attach a [R] Energy from your discard pile to your Active [R] Pokémon.": {
		engine.AttachDiscardedEnergy{Type: engine.Fire},
	}, // Flame Patch
	"Your opponent reveals their hand.": {engine.RevealOppHand{}}, // Hand Scope
	"Flip 2 coins. If both of them are heads, discard a random Energy from your opponent's Active Pokémon.": {
		engine.DiscardOppEnergy{Type: engine.NoEnergy, BothHeads: true},
	}, // Hitting Hammer
	"Heal 20 damage from your Active Pokémon. If you healed any damage in this way, flip a coin. If heads, put this Lucky Ice Pop into your hand instead of the discard pile.": {
		engine.HealActive{N: 20, FlipToReturn: true},
	}, // Lucky Ice Pop
	"Shuffle 2 cards from your hand into your deck. Draw a card. If you can't shuffle in 2 cards, you can't use this card.": {
		engine.HandToDeckThenDraw{N: 2, Draw: 1},
	}, // Maintenance
	"Look at the top card of your deck. If that card is a [P] Pokémon, put it into your hand. If it is not a [P] Pokémon, put it on the bottom of your deck.": {
		engine.TopCardIf{Match: func(c engine.Card) bool { return c.Kind == engine.KindPokemon && c.Type == engine.Psychic }},
	}, // Mythical Slab
	"Your opponent discards cards from their hand until they have 4 cards in their hand.": {
		engine.OppDiscardDownTo{N: 4},
	}, // Nasty Notice
	"Look at the top 3 cards of your deck.": {engine.LookAtTop{N: 3}}, // Pokédex
	"Put a Basic Pokémon from your opponent's discard pile onto their Bench.": {
		engine.BenchOppBasicFromDiscard{},
	}, // Pokémon Flute
	"Heal 20 damage from 1 of your Pokémon.": {engine.HealTarget{N: 20}}, // Potion
	"A card from among both player's hands is chosen at random, revealed to the other player, and shuffled into its owner's deck.": {
		engine.ShuffleRandomHandCard{},
	}, // Prank Spinner
	"Choose 1 of your Basic Pokémon in play. If you have a Stage 2 card in your hand that evolves from that Pokémon, put that card onto the Basic Pokémon to evolve it, skipping the Stage 1. You can't use this card during your first turn or on a Basic Pokémon that was put into play this turn.": {
		engine.RareCandy{BasicOf: basicOf},
	}, // Rare Candy
	"Your opponent shuffles their hand into their deck and draws 3 cards.": {
		engine.OppShuffleHandDraw{N: 3},
	}, // Red Card
	"Switch out your opponent's Active Basic Pokémon to the Bench. (Your opponent chooses the new Active Pokémon.)": {
		engine.SwitchOppActiveBasic{},
	}, // Repel
	"Look at the top card of your deck. Then, you may shuffle your deck.": {
		engine.LookAtTop{N: 1, MayShuffle: true},
	}, // Rotom Dex
	"Discard a [R] Energy from your opponent's Active Pokémon.": {
		engine.DiscardOppEnergy{Type: engine.Fire},
	}, // Squirt Bottle
	"Until the end of your opponent's next turn, your opponent's Active Pokémon's Retreat Cost is 1 more.": {
		engine.OppActiveRetreatUp{N: 1},
	}, // Team Rocket's Goo-zooka
	"Put a random Item card, except any Team Rocket's Thieving Machine, from your opponent's discard pile into your hand.": {
		engine.RandomFromDiscard{FromOpp: true, Match: func(c engine.Card) bool {
			return c.Kind == engine.KindItem && c.Name != "Team Rocket's Thieving Machine"
		}},
	}, // Team Rocket's Thieving Machine
	"Flip a coin. If heads, put a random Item card from your deck into your hand.": {
		engine.OnHeads{Op: engine.SearchDeck{Match: func(c engine.Card) bool { return c.Kind == engine.KindItem }}},
	}, // Order Pad
	"Choose a Pokémon in your hand and switch it with a random Pokémon in your deck.": {
		engine.SwapHandPokemon{},
	}, // Pokémon Communication
	"Choose 1 of your [G] Pokémon in play. Put a random [G] Pokémon from your deck that evolves from that Pokémon onto that Pokémon to evolve it. You can't use this card during your first turn or on a Pokémon that was put into play this turn.": {
		engine.EvolveFromDeck{Type: engine.Grass},
	}, // Quick-Grow Extract
	"During this turn, the Retreat Cost of your Active Pokémon is 1 less.": {
		engine.ThisTurn{Passive: engine.Passive{Hook: engine.HookRetreatCost, Value: -1,
			When: func(c engine.PassiveCheck) bool { return ownedBy(c) && c.Target.Pos == 0 }}},
	}, // X Speed
}

func init() {
	for text, ops := range itemTexts {
		trainerEffects[text] = ops
	}
}
