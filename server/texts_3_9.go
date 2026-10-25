package main

import (
	"slices"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// Supporter texts, batch 3.9 (issue #69). Each wording maps to its effect
// ops; the verbs live in engine/texts_3_9.go and engine/timed.go. A Pokémon
// is matched by its exact name, so "Snorlax" doesn't include "Snorlax ex".

func init() {
	for text, ops := range texts39 {
		trainerEffects[text] = ops
	}
}

var texts39 = map[string][]engine.EffectOp{
	// Acerola
	"Choose 1 of your Palossand or Mimikyu that has damage on it, and move 40 of its damage to your opponent's Active Pokémon.": {
		engine.MoveDamageToOppActive{Amount: 40, Names: []string{"Palossand", "Mimikyu"}}},
	// Adaman
	"During your opponent's next turn, all of your [M] Pokémon take −20 damage from attacks from your opponent's Pokémon.": {
		nextTurn(engine.HookDamageTaken, -20, yourType(engine.Metal))},
	// Arven
	"Flip a coin. If heads, put a random Item card from your deck into your hand. If tails, put a random Pokémon Tool card from your deck into your hand.": {
		engine.FlipSearch{Heads: ofKind(engine.KindItem), Tails: ofKind(engine.KindTool)}},
	// Cabbie
	"Put a random Stadium card from your deck into your hand.": {engine.SearchDeck{Match: ofKind(engine.KindStadium)}},
	// Clemont
	"Put 2 random cards from among Magneton, Heliolisk, and Clemont's Backpack from your deck into your hand.": {
		engine.SearchDeck{Match: cardNamed(clemont...)}, engine.SearchDeck{Match: cardNamed(clemont...)}},
	// Gladion
	"Put 1 random Type: Null or Silvally from your deck into your hand.": {
		engine.SearchDeck{Match: cardNamed("Type: Null", "Silvally")}},
	// Barry
	"During this turn, attacks used by your Snorlax, Heracross, and Staraptor cost 2 less [C] Energy.": {
		thisTurn(engine.HookColorlessCost, -2, yourNamed("Snorlax", "Heracross", "Staraptor"))},
	// Blaine
	"During this turn, attacks used by your Ninetales, Rapidash, or Magmar do +30 damage to your opponent's Active Pokémon.": {
		thisTurn(engine.HookDamageDealt, 30, yourNamed("Ninetales", "Rapidash", "Magmar"))},
	// Blue
	"During your opponent's next turn, all of your Pokémon take −10 damage from attacks from your opponent's Pokémon.": {
		nextTurn(engine.HookDamageTaken, -10, yours)},
	// Brock
	"Take a [F] Energy from your Energy Zone and attach it to Golem or Onix.": {
		engine.AttachZoneEnergy{Energy: engine.Fighting, Names: []string{"Golem", "Onix"}}},
	// Budding Expeditioner
	"Put your Mew ex in the Active Spot into your hand.": {engine.ActiveToHand{Name: "Mew ex"}},
	// Calem
	"Draw a card for each Mega Evolution Pokémon ex in play (both yours and your opponent's).": {engine.DrawPerMegaInPlay{}},
	// Celestic Town Elder
	"Put 1 random Basic Pokémon from your discard pile into your hand.": {
		engine.FromDiscard{Match: engine.Card.IsBasicPokemon}},
	// Cheren
	"During your opponent's next turn, all of your Watchog and Stoutland take −100 damage from attacks from your opponent's Pokémon ex.": {
		nextTurn(engine.HookDamageTaken, -100, fromEx(yourNamed("Watchog", "Stoutland")))},
	// Copycat
	"Shuffle your hand into your deck. Draw a card for each card in your opponent's hand.": {engine.ShuffleHandDrawPerOpp{}},
	// Cynthia
	"During this turn, attacks used by your Garchomp or Togekiss do +50 damage to your opponent's Active Pokémon.": {
		thisTurn(engine.HookDamageDealt, 50, yourNamed("Garchomp", "Togekiss"))},
	// Cyrus
	"Switch in 1 of your opponent's Benched Pokémon that has damage on it to the Active Spot.": {engine.SwitchInDamagedBench{}},
	// Dawn
	"Move an Energy from 1 of your Benched Pokémon to your Active Pokémon.": {engine.MoveBenchEnergyToActive{}},
	// Diantha
	"Heal 90 damage from 1 of your [P] Pokémon that has 2 or more [P] Energy attached. If you healed any damage in this way, discard 2 [P] Energy from that Pokémon.": {
		engine.HealThenDiscard{Amount: 90, Type: engine.Psychic, N: 2}},
	// Elesa
	"Return all Pokémon Tools attached to each Pokémon (both yours and your opponent's) to their owner's hand.": {engine.ReturnAllTools{}},
	// Erika
	"Heal 50 damage from 1 of your [G] Pokémon.": {engine.HealType{Amount: 50, Type: engine.Grass}},
	// Fantina
	"Take a [P] Energy from your Energy Zone and attach it to each of your Drifblim and Mismagius.": {
		engine.AttachZoneEnergyEach{Energy: engine.Psychic, Names: []string{"Drifblim", "Mismagius"}}},
	// Fisher
	"Flip 3 coins. For each heads, a [W] Pokémon is chosen at random from your discard pile and put into your hand.": {
		engine.FromDiscard{Match: pokemonOfType(engine.Water), Coins: 3}},
	// Giovanni
	"During this turn, attacks used by your Pokémon do +10 damage to your opponent's Active Pokémon.": {
		thisTurn(engine.HookDamageDealt, 10, yours)},
	// Guzma
	"Discard all Pokémon Tool cards attached to each of your opponent's Pokémon.": {engine.DiscardOppTools{}},
	// Hala
	"During your opponent's next turn, if your Hariyama or Crabominable would be Knocked Out by damage from an attack, it is not Knocked Out and its remaining HP becomes 10.": {
		nextTurn(engine.HookSurviveKO, 0, yourNamed("Hariyama", "Crabominable"))},
	// Hau
	"During this turn, attacks used by your Decidueye ex, Incineroar ex, or Primarina ex do +30 damage to your opponent's Active Pokémon.": {
		thisTurn(engine.HookDamageDealt, 30, yourNamed("Decidueye ex", "Incineroar ex", "Primarina ex"))},
}

type passiveWhen = func(engine.PassiveCheck) bool

func thisTurn(h engine.Hook, v int, when passiveWhen) engine.EffectOp {
	return engine.ForTurn{Passive: engine.Passive{Hook: h, Value: v, When: when}}
}

func nextTurn(h engine.Hook, v int, when passiveWhen) engine.EffectOp {
	return engine.ForTurn{Passive: engine.Passive{Hook: h, Value: v, When: when}, Next: true}
}

// yours: the Pokémon is the card player's.
func yours(c engine.PassiveCheck) bool {
	return c.Target.InPlay != nil && c.Target.Player == c.Source.Player
}

func yourNamed(names ...string) passiveWhen {
	return func(c engine.PassiveCheck) bool { return yours(c) && slices.Contains(names, c.Target.Card.Name) }
}

func yourType(t engine.Energy) passiveWhen {
	return func(c engine.PassiveCheck) bool { return yours(c) && c.Target.Card.Type == t }
}

// fromEx narrows when to attacks by a Pokémon ex.
func fromEx(when passiveWhen) passiveWhen {
	return func(c engine.PassiveCheck) bool { return when(c) && c.Foe != nil && c.Foe.Card.IsEX }
}

func pokemonOfType(t engine.Energy) engine.CardMatch {
	return func(c engine.Card) bool { return c.Kind == engine.KindPokemon && c.Type == t }
}

var clemont = []string{"Magneton", "Heliolisk", "Clemont's Backpack"}

func ofKind(k engine.CardKind) engine.CardMatch {
	return func(c engine.Card) bool { return c.Kind == k }
}

func cardNamed(names ...string) engine.CardMatch {
	return func(c engine.Card) bool { return slices.Contains(names, c.Name) }
}
