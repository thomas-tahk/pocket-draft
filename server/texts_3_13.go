package main

import (
	"github.com/thomas-tahk/pocket-draft/engine"
)

// Tool and Stadium texts, batch 3.13 (issue #73): the Trainers with no
// trainerKind. Each wording builds its card's passives, event triggers,
// Stadium use or lent attacks; the verbs live in engine/texts_3_13.go and
// engine/passive.go.

func init() {
	for text, build := range texts313 {
		trainerBuilt[text] = build
	}
}

// isAncient and isFuture: a card counts only if the card itself says Ancient /
// Future (owner 2026-10-28, decision 3), i.e. its printing is in ancientIDs /
// futureIDs (texts_3_11.go). The B2a Koraidon and Miraidon are not.
func isAncient(c engine.Card) bool { return ancientIDs[c.ID] }

func isFuture(c engine.Card) bool { return futureIDs[c.ID] }

func anyPokemon(engine.Card) bool { return true }

func ofType(t engine.Energy) func(engine.Card) bool {
	return func(c engine.Card) bool { return c.Type == t }
}

func ofStage(s engine.Stage) func(engine.Card) bool {
	return func(c engine.Card) bool { return c.Kind == engine.KindPokemon && c.Stage == s }
}

// onSelf: the passive applies to the Pokémon its Tool is attached to.
func onSelf(c engine.PassiveCheck) bool {
	return c.Target.InPlay != nil && c.Target.InPlay == c.Source.InPlay
}

// selfPassive is a Tool passive for its Pokémon when that card matches.
func selfPassive(h engine.Hook, v int, match func(engine.Card) bool) engine.Passive {
	return engine.Passive{Hook: h, Value: v, When: func(c engine.PassiveCheck) bool { return onSelf(c) && match(c.Target.Card) }}
}

func hpPassive(n int, match func(engine.Card) bool) engine.Passive {
	return selfPassive(engine.HookMaxHP, n, match)
}

// beastitePassives is +10 for each point (a game ends at 3).
func beastitePassives() []engine.Passive {
	var out []engine.Passive
	for k := 1; k <= 3; k++ {
		out = append(out, engine.Passive{Hook: engine.HookDamageDealt, Value: 10, When: func(c engine.PassiveCheck) bool {
			return onSelf(c) && isUltraBeast(c.Target.Card) && c.Game.S.Players[c.Source.Player].Points >= k
		}})
	}
	return out
}

// retreatCostNow is the Retreat Cost of the Pokémon a passive is asked about.
func retreatCostNow(c engine.PassiveCheck) int {
	if c.Target.Pos == 0 {
		return c.Game.RetreatCost(c.Target.Player)
	}
	return c.Target.Card.RetreatCost
}

func toolCard(ps ...engine.Passive) func(*engine.Card) {
	return func(c *engine.Card) { c.Passive = ps }
}

func stadiumCard(ps ...engine.Passive) func(*engine.Card) { return toolCard(ps...) }

func toolTriggers(ts ...engine.Trigger) func(*engine.Card) {
	return func(c *engine.Card) { c.Triggers = ts }
}

func stadiumTriggers(ts ...engine.Trigger) func(*engine.Card) { return toolTriggers(ts...) }

func stadiumUse(ops ...engine.EffectOp) func(*engine.Card) {
	return func(c *engine.Card) { c.StadiumUse = ops }
}

type triggerWhen = func(engine.TriggerCheck) bool

// activeOf: the Tool's Pokémon is in the Active Spot and matches.
func activeOf(match func(engine.Card) bool) triggerWhen {
	return func(c engine.TriggerCheck) bool { return c.Source.Pos == 0 && match(c.Source.Card) }
}

// ownTurn: it is the Tool owner's turn.
func ownTurn(c engine.TriggerCheck) bool { return c.Turn == c.Source.Player }

// foeTurn: it is the Tool owner's opponent's turn.
func foeTurn(c engine.TriggerCheck) bool { return c.Turn != c.Source.Player }

func both(a, b triggerWhen) triggerWhen {
	return func(c engine.TriggerCheck) bool { return a(c) && b(c) }
}

// previousEvolutionAttacks is every attack of the Pokémon c evolves from,
// back to the Basic: the lowest-numbered printing of each, by name. Looked up
// when asked, since the card data loads after init.
func previousEvolutionAttacks(c engine.Card) []engine.Attack {
	var out []engine.Attack
	for from := c.EvolvesFrom; from != ""; {
		prev, ok := firstPrinting(from)
		if !ok {
			break
		}
		if card, err := toEngineCard(prev); err == nil {
			out = append(out, card.Attacks...)
		}
		from = prev.EvolvesFrom
	}
	return out
}

func firstPrinting(name string) (rawCard, bool) {
	var best rawCard
	for id, rc := range rawByID {
		if rc.Name == name && rc.CardType != "Trainer" && (best.ID == "" || id < best.ID) {
			best = rc
		}
	}
	return best, best.ID != ""
}

var texts313 = map[string]func(*engine.Card){
	// B3a-069 Ancient Booster Energy Capsule
	"The Ancient Pokémon this card is attached to gets +40 HP.": toolCard(hpPassive(40, isAncient)),
	// A2-147 Giant Cape
	"The Pokémon this card is attached to gets +20 HP.": toolCard(hpPassive(20, anyPokemon)),
	// A3-147 Leaf Cape
	"The [G] Pokémon this card is attached to gets +30 HP.": toolCard(hpPassive(30, ofType(engine.Grass))),
	// B3b-065 Elegant Cape
	"The Stage 1 Pokémon this card is attached to gets +30 HP.": toolCard(hpPassive(30, ofStage(engine.Stage1))),
	// B2-154 Starting Plains
	"Each Basic Pokémon in play (both yours and your opponent's) gets +20 HP.": stadiumCard(engine.Passive{Hook: engine.HookMaxHP, Value: 20, When: func(c engine.PassiveCheck) bool { return c.Target.Card.IsBasicPokemon() }}),
	// B3a-070 Future Booster Energy Capsule
	"Attacks used by the Future Pokémon this card is attached to do +20 damage to your opponent's Active Pokémon.": toolCard(selfPassive(engine.HookDamageDealt, 20, isFuture)),
	// A3a-066 Beastite
	"Attacks used by the Ultra Beast this card is attached to do +10 damage to your opponent's Active Pokémon for each point you have gotten.": toolCard(beastitePassives()...),
	// B3-154 Arena of Antiquity
	"Attacks used by each [F] Pokémon in play (both yours and your opponent's) do +20 damage to the opponent's Active Pokémon ex.": stadiumCard(engine.Passive{Hook: engine.HookDamageDealt, Value: 20, When: func(c engine.PassiveCheck) bool {
		return c.Target.Card.Type == engine.Fighting && c.Foe != nil && c.Foe.Card.IsEX
	}}),
	// B2-153 Training Area
	"Attacks used by Stage 1 Pokémon in play (both yours and your opponent's) do +10 damage to the opponent's Active Pokémon.": stadiumCard(engine.Passive{Hook: engine.HookDamageDealt, Value: 10, When: func(c engine.PassiveCheck) bool { return c.Target.Card.Stage == engine.Stage1 }}),
	// B3-155 Bounded Field
	"When applying the opponent's Active Pokémon's Weakness to damage from attacks used by Pokémon in play (both yours and your opponent's) that aren't Mega Evolution Pokémon ex, apply Weakness as ×2.": stadiumCard(engine.Passive{Hook: engine.HookWeaknessDouble, When: func(c engine.PassiveCheck) bool { return !c.Target.Card.IsMegaEX }}),
	// B2a-087 Big Air Balloon
	"The Stage 2 Pokémon this card is attached to has no Retreat Cost.": toolCard(selfPassive(engine.HookNoRetreatCost, 0, ofStage(engine.Stage2))),
	// B3b-064 Small Balloon
	"The Retreat Cost of the Basic Pokémon this card is attached to is 1 less.": toolCard(selfPassive(engine.HookRetreatCost, -1, ofStage(engine.Basic))),
	// A4a-067 Inflatable Boat
	"The Retreat Cost of the [W] Pokémon this card is attached to is 1 less.": toolCard(selfPassive(engine.HookRetreatCost, -1, ofType(engine.Water))),
	// B2-155 Peculiar Plaza
	"The Retreat Cost of each [P] Pokémon in play (both yours and your opponent's) is 2 less.": stadiumCard(engine.Passive{Hook: engine.HookRetreatCost, Value: -2, When: func(c engine.PassiveCheck) bool { return c.Target.Card.Type == engine.Psychic }}),
	// B1-219 Heavy Helmet
	"If the Pokémon this card is attached to has a Retreat Cost of 3 or more, it takes −20 damage from attacks from your opponent's Pokémon.": toolCard(engine.Passive{Hook: engine.HookDamageTaken, Value: -20, When: func(c engine.PassiveCheck) bool { return onSelf(c) && retreatCostNow(c) >= 3 }}),
	// B4-149 Clear Veil
	"Prevent all effects of attacks used by your opponent's Pokémon done to the Pokémon this card is attached to. (Existing effects are not removed.)": toolCard(selfPassive(engine.HookEffectShield, 0, anyPokemon)),
	// B2-147 Protective Poncho
	"As long as the Pokémon this card is attached to is on your Bench, prevent all damage done to that Pokémon by your opponent's attacks and Abilities.": toolCard(selfPassive(engine.HookBenchShield, 0, anyPokemon)),
	// B2b-117 Metal Core Barrier
	"If this card is attached to 1 of your Pokémon, discard it at the end of your opponent's turn.The [M] Pokémon this card is attached to takes −50 damage from attacks from your opponent's Pokémon.": func(c *engine.Card) {
		toolCard(selfPassive(engine.HookDamageTaken, -50, ofType(engine.Metal)))(c)
		c.Triggers = []engine.Trigger{{On: engine.OnTurnEnd, When: foeTurn, Do: engine.DiscardThis{}}}
	},
	// A4-153 Steel Apron
	"The [M] Pokémon this card is attached to takes −10 damage from attacks from your opponent's Pokémon, recovers from all Special Conditions, and can't be affected by any Special Conditions.": func(c *engine.Card) {
		toolCard(selfPassive(engine.HookDamageTaken, -10, ofType(engine.Metal)), selfPassive(engine.HookNoConditions, 0, ofType(engine.Metal)))(c)
		cure := engine.Trigger{When: func(tc engine.TriggerCheck) bool { return tc.Source.Card.Type == engine.Metal }, Do: engine.CureSelf{}}
		attached, turnEnd := cure, cure
		attached.On, turnEnd.On = engine.OnAttached, engine.OnTurnEnd
		c.Triggers = []engine.Trigger{attached, turnEnd}
	},
	// B4-154 Soothing Shore
	"At the end of each player's turn, that player heals 20 damage from each of their Pokémon that has any [W] Energy attached.": stadiumTriggers(engine.Trigger{On: engine.OnTurnEnd, Do: engine.HealEachWithEnergy{N: 20, Type: engine.Water}}),
	// B2b-069 Hiking Trail
	"At the end of each player's turn, that player draws cards until they have 3 cards in their hand.": stadiumTriggers(engine.Trigger{On: engine.OnTurnEnd, Do: engine.TurnPlayerDrawsTo{N: 3}}),
	// B4-155 Rainbow Cave
	"Once during each player's turn, that player may discard the Energy that has been generated in their Energy Zone. If they do, the next Energy is produced.": stadiumUse(engine.RerollEnergy{}),
	// B4a-072 Arcade
	"Once during each player's turn, that player may flip 3 coins. If all of them are heads, that player draws cards until they have 7 cards in their hand.": stadiumUse(engine.FlipAllDrawTo{Coins: 3, N: 7}),
	// B3a-074 Area Zero
	"Once during each player's turn, that player may shuffle a Basic Pokémon from their hand into their deck. If they do, they draw a card.": stadiumUse(engine.BasicToDeckDraw{}),
	// B3b-069 Kid's Room
	"Once during each player's turn, that player may choose a card in their hand and switch it with a random Pokémon Tool card in their deck.": stadiumUse(engine.SwapForTool{}),
	// B3-153 Fragrant Forest
	"Once during each player's turn, that player may put a random Basic [G] Pokémon from their deck into their hand.": stadiumUse(engine.SearchDeck{Match: func(c engine.Card) bool { return c.IsBasicPokemon() && c.Type == engine.Grass }}),
	// B2a-093 Mesagoza
	"Once during each player's turn, that player may flip a coin. If heads, that player puts a random Pokémon from their deck into their hand.": stadiumUse(engine.FlipSearch{Heads: func(c engine.Card) bool { return c.Kind == engine.KindPokemon }, Tails: func(engine.Card) bool { return false }}),
	// A4-154 Dark Pendant
	"If the [D] Pokémon this card is attached to is in the Active Spot and is damaged by an attack from your opponent's Pokémon, your opponent reveals a random card from their hand and shuffles it into their deck.": toolTriggers(engine.Trigger{On: engine.OnAttackHit, When: activeOf(ofType(engine.Darkness)), Do: engine.FoeHandCardToDeck{}}),
	// B4-148 Deceptive Needle
	"At the end of your turn, if the [D] Pokémon this card is attached to is in the Active Spot, do 10 damage to your opponent's Active Pokémon.": toolTriggers(engine.Trigger{On: engine.OnTurnEnd, When: both(ownTurn, activeOf(ofType(engine.Darkness))), Do: engine.DamageFoeActive{N: 10}}),
	// A3a-065 Electrical Cord
	"If the [L] Pokémon this card is attached to is in the Active Spot and is Knocked Out by damage from an attack from your opponent's Pokémon, move 2 [L] Energy from that Pokémon and attach 1 Energy each to 2 of your Benched Pokémon.": toolTriggers(engine.Trigger{On: engine.OnKnockedOut, When: activeOf(ofType(engine.Lightning)), Do: engine.EnergyToBench{N: 2, Type: engine.Lightning}}),
	// B3-148 Lucky Egg
	"If the Pokémon this card is attached to is Knocked Out by damage from an attack from your opponent's Pokémon, draw cards until you have 5 cards in your hand.": toolTriggers(engine.Trigger{On: engine.OnKnockedOut, Do: engine.OwnerDrawsTo{N: 5}}),
	// B1-220 Lucky Mittens
	"Whenever your opponent's Pokémon is Knocked Out by damage from an attack used by the Pokémon this card is attached to, draw a card.": toolTriggers(engine.Trigger{On: engine.OnKnocksOut, Do: engine.OwnerDraws{N: 1}}),
	// A2-149 Lum Berry
	"At the end of each turn, if the Pokémon this card is attached to is affected by any Special Conditions, it recovers from all of them, and discard this card.": toolTriggers(engine.Trigger{On: engine.OnTurnEnd, Do: engine.CureThenDiscard{}}),
	// A3b-067 Leftovers
	"At the end of your turn, if the Pokémon this card is attached to is in the Active Spot, heal 10 damage from that Pokémon.": toolTriggers(engine.Trigger{On: engine.OnTurnEnd, When: both(ownTurn, activeOf(anyPokemon)), Do: engine.HealToolPokemon{N: 10}}),
	// B1-218 Sitrus Berry
	"At the end of each turn, if the Pokémon this card is attached to has half of its maximum HP or less remaining, heal 30 damage from it. If you do, discard this card.": toolTriggers(engine.Trigger{On: engine.OnTurnEnd, Do: engine.HealAtHalf{N: 30}}),
	// A3-146 Poison Barb
	"If the Pokémon this card is attached to is your Active Pokémon and is damaged by an attack from your opponent's Pokémon, the Attacking Pokémon is now Poisoned.": toolTriggers(engine.Trigger{On: engine.OnAttackHit, When: activeOf(anyPokemon), Do: engine.PoisonAttacker{}}),
	// A2-148 Rocky Helmet
	"If the Pokémon this card is attached to is in the Active Spot and is damaged by an attack from your opponent's Pokémon, do 20 damage to the Attacking Pokémon.": toolTriggers(engine.Trigger{On: engine.OnAttackHit, When: activeOf(anyPokemon), Do: engine.DamageAttacker{N: 20}}),
	// A4-155 Rescue Scarf
	"If the Pokémon this card is attached to is Knocked Out by damage from an attack from your opponent's Pokémon, put it into your hand instead of the discard pile.": toolTriggers(engine.Trigger{On: engine.OnKnockedOut, Do: engine.ToHandInstead{}}),
	// A4a-068 Memory Light
	"The Pokémon this card is attached to can use any attack from its previous Evolutions. (You still need the necessary Energy to use each attack.)": func(c *engine.Card) { c.LendAttacks = previousEvolutionAttacks },
}
