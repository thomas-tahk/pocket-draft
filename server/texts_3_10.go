package main

import "github.com/thomas-tahk/pocket-draft/engine"

// Supporter texts, part 2 of 3 (issue #70, plan 3.10). Not built yet: Hiker
// and Morty (put cards back "in any order" needs a reorder picker) and May
// (choosing cards from hand needs a hand picker); see docs/questions/M2.md.

func init() {
	for text, ops := range texts3_10 {
		trainerEffects[text] = ops
	}
}

func stage2(m *engine.InPlay) bool { return m.Card.Stage == engine.Stage2 }

func notEX(m *engine.InPlay) bool { return !m.Card.IsEX }

func ultraBeast(m *engine.InPlay) bool { return m.Card.IsUltraBeast() }

func hasWater(m *engine.InPlay) bool { return m.Energy[engine.Water] > 0 }

// attackBonusVsEX is "attacks used by your <which> do +n damage to your
// opponent's Active Pokémon ex" for the rest of the turn.
func attackBonusVsEX(n int, which engine.MonFilter) engine.ForTurn {
	return engine.ForTurn{Passive: engine.Passive{Hook: engine.HookDamageDealt, Value: n,
		When: func(pc engine.PassiveCheck) bool {
			return yours(pc) && which(pc.Target.InPlay) && pc.Foe != nil && pc.Foe.Card.IsEX
		}}}
}

func stage2Card(c engine.Card) bool { return c.Kind == engine.KindPokemon && c.Stage == engine.Stage2 }

func smallBasic(c engine.Card) bool { return c.IsBasicPokemon() && c.HP <= 50 }

func energyTypesInPlay(ctx *engine.EffectContext) int {
	types := map[engine.Energy]bool{}
	for _, m := range ctx.InPlay() {
		for e, n := range m.Energy {
			if n > 0 {
				types[e] = true
			}
		}
	}
	return len(types)
}

func hasInPlay(name string) func(*engine.EffectContext) bool {
	return func(ctx *engine.EffectContext) bool {
		for _, m := range ctx.InPlay() {
			if m.Card.Name == name {
				return true
			}
		}
		return false
	}
}

var texts3_10 = map[string][]engine.EffectOp{
	// Ilima
	"Put 1 of your [C] Pokémon that has damage on it into your hand.": {
		engine.ToHand{Which: engine.Both(engine.OfType(engine.Colorless), engine.Damaged)}},
	// Iono
	"Each player shuffles the cards in their hand into their deck, then draws that many cards.": {
		engine.ShuffleHandsDraw{}},
	// Irida
	"Heal 40 damage from each of your Pokémon that has any [W] Energy attached.": {
		engine.HealEach{Amount: 40, Which: hasWater}},
	// Iris
	"During this turn, if your opponent's Active Pokémon is Knocked Out by damage from an attack used by your Haxorus, you get 1 more point.": {
		engine.ForTurn{Passive: engine.Passive{Hook: engine.HookKOPoints, Value: 1,
			When: func(pc engine.PassiveCheck) bool {
				return pc.Target.Player != pc.Source.Player && pc.Foe.Card.Name == "Haxorus"
			}}}},
	// Jasmine
	"During your opponent's next turn, all of your Steelix and Skarmory ex take −50 damage from attacks from your opponent's Pokémon.": {
		engine.ForTurn{Next: true, Passive: engine.Passive{Hook: engine.HookDamageTaken, Value: -50,
			When: func(pc engine.PassiveCheck) bool {
				return yours(pc) && engine.Named("Steelix", "Skarmory ex")(pc.Target.InPlay)
			}}}},
	// Juggler
	"You can use this card only if your Pokémon in play have 3 or more different types of Energy attached.Move all Energy from each of your Benched Pokémon to your Active Pokémon.": {
		engine.Require{Why: "your Pokémon need 3 or more types of Energy attached",
			Test: func(ctx *engine.EffectContext) bool { return energyTypesInPlay(ctx) >= 3 }},
		engine.MoveBenchEnergy{Type: engine.NoEnergy}},
	// Juliana
	"Put a random Stage 2 Pokémon from your deck into your hand.": {
		engine.SearchDeck{Match: stage2Card}},
	// Lisia
	"Put 2 random Basic Pokémon with 50 HP or less from your deck into your hand.": {
		engine.SearchDeck{Match: smallBasic}, engine.SearchDeck{Match: smallBasic}},
	// Kiawe
	"Choose 1 of your Alolan Marowak or Turtonator. Take 2 [R] Energy from your Energy Zone and attach it to that Pokémon. Your turn ends.": {
		engine.AttachFromZone{Which: engine.Named("Alolan Marowak", "Turtonator"), Type: engine.Fire, N: 2},
		engine.EndTurnNow{}},
	// Koga
	"Put your Muk or Weezing in the Active Spot into your hand.": {
		engine.ToHand{Which: engine.Named("Muk", "Weezing"), ActiveOnly: true}},
	// Korrina
	"During this turn, attacks used by your [F] Pokémon do +30 damage to your opponent's Active Pokémon ex.": {
		attackBonusVsEX(30, engine.OfType(engine.Fighting))},
	// Lana
	"You can use this card only if you have Araquanid in play.Switch in 1 of your opponent's Benched Pokémon to the Active Spot.": {
		engine.Require{Why: "you need Araquanid in play", Test: hasInPlay("Araquanid")},
		engine.GustBench{}},
	// Leaf
	"During this turn, the Retreat Cost of your Active Pokémon is 2 less.": {
		engine.ForTurn{Passive: engine.Passive{Hook: engine.HookRetreatCost, Value: -2,
			When: func(pc engine.PassiveCheck) bool { return yours(pc) && pc.Target.Pos == 0 }}}},
	// Lillie
	"Heal 60 damage from 1 of your Stage 2 Pokemon.": {
		engine.HealOne{Amount: 60, Which: stage2}},
	// Looker
	"Your opponent reveals all of the Supporter cards in their deck.": {
		engine.RevealOppDeck{Kind: engine.KindSupporter}},
	// Lt. Surge
	"Move all [L] Energy from your Benched Pokémon to your Raichu, Electrode, or Electabuzz in the Active Spot.": {
		engine.Require{Why: "your Active must be Raichu, Electrode or Electabuzz",
			Test: func(ctx *engine.EffectContext) bool {
				return engine.Named("Raichu", "Electrode", "Electabuzz")(ctx.Player().Active)
			}},
		engine.MoveBenchEnergy{Type: engine.Lightning}},
	// Lusamine
	"You can use this card only if your opponent has gotten at least 1 point.Choose 1 of your Ultra Beasts. Attach 2 random Energy from your discard pile to that Pokémon.": {
		engine.Require{Why: "your opponent has no points yet",
			Test: func(ctx *engine.EffectContext) bool { return ctx.Opponent().Points >= 1 }},
		engine.AttachFromDiscard{Which: ultraBeast, N: 2}},
	// Lyra
	"Switch your Active Pokémon that has damage on it with 1 of your Benched Pokémon.": {
		engine.SwitchActive{Which: engine.Damaged}},
	// Mallow
	"Heal all damage from 1 of your Shiinotic or Tsareena. If you do, discard all Energy from that Pokémon.": {
		engine.HealOne{Which: engine.Named("Shiinotic", "Tsareena"), DiscardEnergy: true}},
	// Marlon
	"Heal 70 damage from 1 of your Carracosta or Jellicent.": {
		engine.HealOne{Amount: 70, Which: engine.Named("Carracosta", "Jellicent")}},
	// Mars
	"Your opponent shuffles their hand into their deck and draws a card for each of their remaining points needed to win.": {
		engine.OppRedrawPointsLeft{}},
	// Misty
	"Choose 1 of your [W] Pokémon, and flip a coin until you get tails. For each heads, take a [W] Energy from your Energy Zone and attach it to that Pokémon.": {
		engine.AttachPerHeads{Which: engine.OfType(engine.Water), Type: engine.Water}},
	// Nemona
	"During this turn, attacks used by your Pawmot do +80 damage to your opponent's Active Pokémon ex.": {
		attackBonusVsEX(80, engine.Named("Pawmot"))},
	// Parasol Lady
	"Put 1 of your [W] Pokémon in play, except any Pokémon ex, into your hand.": {
		engine.ToHand{Which: engine.Both(engine.OfType(engine.Water), notEX)}},
	// Penny
	"Look at a random Supporter card that's not Penny from your opponent's deck and shuffle it back into their deck. Use the effect of that card as the effect of this card.": {
		engine.CopyOppSupporter{}},
}
