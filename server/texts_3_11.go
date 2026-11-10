package main

import (
	"strings"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// Supporter texts, part 3 (issue #71). Each printed wording maps to the
// engine ops that play it (engine/texts_3_11.go).

// ancientIDs and futureIDs are the Ancient and Future Pokémon printings
// (Paradox Pokémon from B3a on). The scrape carries no such tag. They are the
// one list: isAncient / isFuture (texts_3_13.go) read them for every card text
// and the draft's niche groups are checked against them.
var ancientIDs = idSet("B3a-003", "B3b-035", "B3a-026", "B3a-083", "B3a-091", "B3a-054", "B3a-034",
	"B3a-036", "B3a-084", "B3a-092", "B3a-055", "B3a-079", "B3a-047", "B3a-109", "B3a-035", "B3a-025",
	"B3a-004", "B3a-053",
	// B4b "Deluxe Pack: Mega" reprints of the printings above.
	"B4b-136", "B4b-429", "B4b-120", "B4b-152", "B4b-333")

var futureIDs = idSet("B3a-029", "B3a-077", "B3b-020", "B3a-013", "B3a-081", "B3a-089", "B3a-030",
	"B3a-017", "B3a-046", "B3a-028", "B3a-005", "B3a-018", "B3a-051", "B3a-027", "B3a-108",
	"B3a-019", "B3a-082", "B3a-090",
	// B4b "Deluxe Pack: Mega" reprints of the printings above.
	"B4b-103", "B4b-428", "B4b-077", "B4b-102", "B4b-303")

func idSet(ids ...string) map[string]bool {
	out := make(map[string]bool, len(ids))
	for _, id := range ids {
		out[id] = true
	}
	return out
}

func named(names ...string) func(engine.Card) bool {
	set := idSet(names...)
	return func(c engine.Card) bool { return set[c.Name] }
}

func hasAttack(name string) func(engine.Card) bool {
	return func(c engine.Card) bool {
		for _, a := range c.Attacks {
			if a.Name == name {
				return true
			}
		}
		return false
	}
}

func isPokemon(c engine.Card) bool { return c.Kind == engine.KindPokemon }

// hasAbility reads the scrape: the engine doesn't load Abilities yet.
func hasAbility(c engine.Card) bool { return rawByID[c.ID].Ability != nil }

// thisTurnDamage is "During this turn, attacks used by your Pokémon that
// attacker accepts do +n damage to your opponent's Active Pokémon that foe
// accepts."
func thisTurnDamage(n int, attacker, foe func(engine.Card) bool) engine.ThisTurn {
	return engine.ThisTurn{Passive: engine.Passive{
		Hook: engine.HookDamageDealt, Value: n,
		When: func(c engine.PassiveCheck) bool {
			return c.Target.Player == c.Source.Player && c.Foe != nil &&
				(attacker == nil || attacker(c.Target.Card)) && (foe == nil || foe(c.Foe.Card))
		},
	}}
}

var allConditions = []engine.Status{engine.Poison, engine.Burn, engine.Asleep, engine.Paralyzed, engine.Confused}

var texts311 = map[string][]engine.EffectOp{
	"You can use this card only if you have Galarian Obstagoon in play.Discard 2 random Energy from your opponent's Active Pokémon.": { // Piers
		engine.NeedInPlay{Name: "Galarian Obstagoon"}, engine.DiscardOppEnergyN{N: 2}},
	"Heal 30 damage from 1 of your Pokémon, and it recovers from all Special Conditions.": { // Pokémon Center Lady
		engine.Heal{Amount: 30, Cures: allConditions}},
	"Attach 3 different types of Energy from your discard pile to your Ancient Pokémon in any way you like.": { // Professor Sada
		engine.AttachChosenFromDiscard{N: 3, Type: engine.NoEnergy, Distinct: true, Fits: isAncient}},
	"Shuffle 1 of your Future Pokémon in play into your deck.": { // Professor Turo
		engine.ShuffleIntoDeck{Fits: isFuture}},
	"You can use this card only if your Pokémon in the Active Spot has the Psychic attack.Choose 1 of your opponent's Benched Pokémon and move a random Energy from it to your opponent's Active Pokémon.": { // Psychic
		engine.NeedActiveAttack{Attack: "Psychic"}, engine.MoveOppEnergyToActive{}},
	"Look at the top 4 cards of your deck. Put all Pokémon you find there that have the Puppy Pile attack into your hand. Shuffle the other cards back into your deck.": { // Puppy-Loving Girl
		engine.TopCardsToHand{N: 4, Fits: hasAttack("Puppy Pile")}},
	"During this turn, attacks used by your Pokémon do +20 damage to your opponent's Active Pokémon ex.": { // Red
		thisTurnDamage(20, nil, func(c engine.Card) bool { return c.IsEX })},
	"Switch out your opponent's Active Pokémon to the Bench. (Your opponent chooses the new Active Pokémon.)": { // Sabrina
		engine.SwitchOppActive{}},
	"Look at the top 4 cards of your deck. Put all Stage 1 Pokémon you find there into your hand. Shuffle the other cards back into your deck.": { // Sightseer
		engine.TopCardsToHand{N: 4, Fits: func(c engine.Card) bool { return isPokemon(c) && c.Stage == engine.Stage1 }}},
	"Your opponent reveals their hand. Choose a Supporter card you find there and shuffle it into your opponent's deck.": { // Silver
		engine.ShuffleOppSupporter{}},
	"Put a random Mega Evolution Pokémon ex from your deck into your hand.": { // Serena
		engine.SearchDeck{Match: func(c engine.Card) bool { return c.IsMegaEX }}},
	"Put 1 random Glameow, Stunky, or Croagunk from your deck into your hand.": { // Team Galactic Grunt
		engine.SearchDeck{Match: named("Glameow", "Stunky", "Croagunk")}},
	"Flip a coin until you get tails. For each heads, put a random Pokémon that has “Team Rocket” in its name from your deck into your hand.": { // Team Rocket's Researcher
		engine.PerHeadsUntilTails{Op: engine.SearchDeck{Match: func(c engine.Card) bool {
			return isPokemon(c) && strings.Contains(c.Name, "Team Rocket")
		}}}},
	"Switch your Active Stage 1 Pokémon with 1 of your Benched Pokémon.": { // Skyla
		engine.SwitchOwnActive{Fits: func(c engine.Card) bool { return c.Stage == engine.Stage1 }}},
	"During this turn, attacks used by your Alolan Golem, Vikavolt, or Togedemaru do +30 damage to your opponent's Active Pokémon.": { // Sophocles
		thisTurnDamage(30, named("Alolan Golem", "Vikavolt", "Togedemaru"), nil)},
	"Flip a coin until you get tails. For each heads, discard a random Energy from your opponent's Active Pokémon.": { // Team Rocket Grunt
		engine.PerHeadsUntilTails{Op: engine.DiscardOppEnergyN{N: 1}}},
	"Look at your opponent's hand and put any number of Basic Pokémon you find there onto your opponent's Bench.": { // Team Rocket's Boss
		engine.BenchOppBasics{}},
	"Your opponent's Active Pokémon is now Confused. Flip a coin. If tails, your Active Pokémon is now also Confused.": { // Team Rocket's Master Plan
		engine.ApplyStatus{Status: engine.Confused}, engine.SelfStatusOnTails{Status: engine.Confused}},
	"Discard a random Energy from among the Energy attached to all of your opponent's Pokémon that have an Ability.": { // Team Star Grunt
		engine.DiscardEnergyAmong{Fits: hasAbility}},
	"Look at the top 4 cards of your deck. Put all Pokémon Tool cards you find there into your hand. Shuffle the other cards back into your deck.": { // Traveling Merchant
		engine.TopCardsToHand{N: 4, Fits: func(c engine.Card) bool { return c.Kind == engine.KindTool }}},
	"Choose 1 of your Electivire or Luxray. Attach 2 [L] Energy from your discard pile to that Pokémon.": { // Volkner
		engine.AttachChosenFromDiscard{N: 2, Type: engine.Lightning, Fits: named("Electivire", "Luxray")}},
	"Choose 1 of your [W] Pokémon in play with a maximum HP of 50 or less. Put a random [W] Pokémon from your deck that evolves from that Pokémon onto that Pokémon to evolve it.": { // Wallace
		engine.EvolveFitFromDeck{
			Fits:    func(c engine.Card) bool { return c.Type == engine.Water && c.HP <= 50 },
			Evolves: func(c engine.Card) bool { return c.Type == engine.Water }}},
	"Take a [C] Energy from your Energy Zone and attach it to 1 of your Stage 2 Pokémon.": { // Wally
		engine.AttachNewEnergy{Type: engine.Colorless, Fits: func(c engine.Card) bool { return isPokemon(c) && c.Stage == engine.Stage2 }}},
	"Heal 60 damage from 1 of your Miltank, and it recovers from being Asleep, Paralyzed, and Confused.": { // Whitney
		engine.Heal{Amount: 60, Cures: []engine.Status{engine.Asleep, engine.Paralyzed, engine.Confused}, Fits: named("Miltank")}},
	"The next time you flip any number of coins for the effect of an attack, Ability, or Trainer card after using this card on this turn, the first coin flip will definitely be heads.": { // Will
		engine.NextFlipHeads{}},
}

func init() {
	for text, ops := range texts311 {
		trainerEffects[text] = ops
	}
}
