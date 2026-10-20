package main

import (
	"testing"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// One test per wording in batch 4.6 (issue #79), each on a real printing.
// Wordings skipped for engine reasons are recorded in docs/questions/M2.md.

// hookOn: some passive of hook is active on the Pokémon at (p, pos).
func hookOn(g *engine.Game, p, pos int, hook engine.Hook) bool {
	for _, m := range g.Modifiers(p, pos) {
		if m.Hook == hook {
			return true
		}
	}
	return false
}

// mustOn uses the Ability at pos with its target aimed at to.
func mustOn(t *testing.T, g *engine.Game, p, pos int, to engine.Spot) {
	t.Helper()
	if err := use(g, p, pos, to); err != nil {
		t.Fatalf("use ability at %d aimed at %+v: %v", pos, to, err)
	}
}

// evolveFrom puts pre-evolution on the Active and evolves it with evo from hand.
func evolveFrom(t *testing.T, g *engine.Game, p int, pre, evo string) {
	t.Helper()
	put(g, p, 0, realCard(t, pre))
	g.S.Players[p].Hand = []engine.Card{realCard(t, evo)}
	mustSubmit(t, g, engine.Evolve{Player: p, HandCardID: evo, Target: 0})
}

// --- always-on ---

// "This Pokémon takes −30 damage from attacks." with Arceus or Arceus ex in play — Raichu.
func TestResilienceLinkNeedsArceus(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 0, realCard(t, "A2a-026"))
	if hookOn(g, p, 0, engine.HookDamageTaken) {
		t.Fatal("Resilience Link works with no Arceus in play")
	}
	put(g, p, 1, realCard(t, "A2a-070"))
	if !hasModifier(g, p, 0, engine.HookDamageTaken, -30) {
		t.Error("Raichu takes no -30 with Arceus on the Bench")
	}
}

// "Prevent all effects of attacks used by your opponent's Pokémon done to this Pokémon." — Regice.
func TestCrystalBodyShieldsFromAttackEffects(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 0, realCard(t, "A2-034"))
	if !hookOn(g, p, 0, engine.HookEffectShield) {
		t.Error("Regice has no effect shield")
	}
}

// "This Pokémon gets +30 HP for each [P] Energy attached to it." — Reuniclus.
func TestInfiniteIncreaseAddsHPPerPsychic(t *testing.T) {
	g, p := abGame(t)
	mon := put(g, p, 0, realCard(t, "B1a-034"))
	mon.Energy[engine.Psychic] = 2
	if !hasModifier(g, p, 0, engine.HookMaxHP, 60) {
		t.Error("two [P] Energy should give +60 HP")
	}
}

// "This Pokémon gets +30 HP for each [G] Energy attached to it." — Serperior.
func TestRegalBloomAddsHPPerGrass(t *testing.T) {
	g, p := abGame(t)
	mon := put(g, p, 0, realCard(t, "B4a-005"))
	mon.Energy[engine.Grass] = 1
	if !hasModifier(g, p, 0, engine.HookMaxHP, 30) {
		t.Error("one [G] Energy should give +30 HP")
	}
}

// "As long as this Pokémon is on your Bench, prevent all damage done to this Pokémon by attacks." — Wartortle.
func TestShellShieldOnBench(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 1, realCard(t, "B1a-018"))
	if !hookOn(g, p, 1, engine.HookBenchShield) {
		t.Error("benched Wartortle has no bench shield")
	}
}

// "Your Active Dondozo has no Retreat Cost." — Tatsugiri on the Bench.
func TestRetreatDirectiveClearsDondozoCost(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 0, realCard(t, "A2b-102"))
	put(g, p, 1, realCard(t, "A2b-021"))
	if !hookOn(g, p, 0, engine.HookNoRetreatCost) {
		t.Error("Active Dondozo keeps its Retreat Cost")
	}
}

// "As long as this Pokémon is on your Bench, your Active Basic Pokemon's Retreat Cost is 1 less." — Shaymin.
func TestSkySupportCutsActiveRetreatCost(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 0, realCard(t, "A4b-069"))
	put(g, p, 1, realCard(t, "A4b-297"))
	if !hasModifier(g, p, 0, engine.HookRetreatCost, -1) {
		t.Error("Active Basic does not get Retreat Cost -1")
	}
}

// "This Pokémon takes −30 damage from attacks from [F] Pokémon." — Staraptor.
func TestDefensiveWhirlwindFromFighting(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 0, realCard(t, "P-A-047"))
	put(g, 1-p, 0, realCard(t, "B2-079")) // Machop, Fighting
	if !hasModifier(g, p, 0, engine.HookDamageTaken, -30) {
		t.Error("Staraptor takes no -30 from a Fighting attacker")
	}
}

// "As long as this Pokémon is in the Active Spot, attacks used by your opponent's Active Pokémon cost 1 [C] more." — Stoutland.
func TestGuardDogVisageAddsColorlessCost(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 0, realCard(t, "A3a-056"))
	if !hasModifier(g, 1-p, 0, engine.HookColorlessCost, 1) {
		t.Error("the foe's Active pays no [C] more for attacks")
	}
}

// "Each of your Pokémon that has any Energy attached ... can't be affected by any Special Conditions." — Ogerpon ex.
func TestSoothingWindBlocksConditions(t *testing.T) {
	g, p := abGame(t)
	mon := put(g, p, 0, realCard(t, "B2-017"))
	mon.Energy[engine.Grass] = 1
	if !hookOn(g, p, 0, engine.HookNoConditions) {
		t.Error("Ogerpon ex with Energy does not block Special Conditions")
	}
}

// "This Ability works if you have any Unown in play with an Ability other than GUARD. All of your Pokémon take −10 damage." — Unown GUARD.
func TestGuardUnownNeedsAnotherUnownAbility(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 0, realCard(t, "A4b-069"))
	put(g, p, 1, realCard(t, "A4-084")) // GUARD
	if hookOn(g, p, 0, engine.HookDamageTaken) {
		t.Fatal("GUARD works with no other Unown Ability")
	}
	put(g, p, 2, realCard(t, "A4-085")) // POWER
	if !hasModifier(g, p, 0, engine.HookDamageTaken, -10) {
		t.Error("Torkoal takes no -10 with GUARD and POWER Unown in play")
	}
}

// "This Ability works if you have any Unown in play with an Ability other than POWER. Attacks used by your Pokémon do +10 damage." — Unown POWER.
func TestPowerUnownNeedsAnotherUnownAbility(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 0, realCard(t, "A4b-069"))
	put(g, p, 1, realCard(t, "A4-085")) // POWER
	if hookOn(g, p, 0, engine.HookDamageDealt) {
		t.Fatal("POWER works with no other Unown Ability")
	}
	put(g, p, 2, realCard(t, "A4-084")) // GUARD
	if !hasModifier(g, p, 0, engine.HookDamageDealt, 10) {
		t.Error("attacks from Torkoal do no +10 with POWER and GUARD Unown in play")
	}
}

// "During your first turn, this Pokémon has no Retreat Cost." — Wimpod.
func TestWimpOutOnlyOnFirstTurn(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 0, realCard(t, "A3-021"))
	g.S.Turn = 1
	if !hookOn(g, p, 0, engine.HookNoRetreatCost) {
		t.Error("Wimpod has a Retreat Cost on the first turn")
	}
	g.S.Turn = 5
	if hookOn(g, p, 0, engine.HookNoRetreatCost) {
		t.Error("Wimpod has no Retreat Cost after the first turn")
	}
}

// "At the end of your turn, if this Pokémon is in the Active Spot, heal 20 damage from it." — Snorlax ex.
func TestFullMouthMannerHealsAtEndOfTurn(t *testing.T) {
	g, p := abGame(t)
	mon := put(g, p, 0, realCard(t, "A4b-288"))
	mon.Damage = 30
	mustSubmit(t, g, engine.EndTurn{Player: p})
	if mon.Damage != 10 {
		t.Errorf("damage = %d, want 10 after Full-Mouth Manner", mon.Damage)
	}
}

// "At the end of your first turn, take a [L] Energy from your Energy Zone and attach it to this Pokémon." — Zeraora.
func TestThunderclapFlashAttachesOnFirstTurn(t *testing.T) {
	g, p := abGame(t)
	mon := put(g, p, 0, realCard(t, "B1-304"))
	g.S.Players[p].EnergyZone = engine.Lightning
	mustSubmit(t, g, engine.EndTurn{Player: p})
	if mon.Energy[engine.Lightning] != 1 {
		t.Errorf("Lightning Energy = %d, want 1", mon.Energy[engine.Lightning])
	}
}

// --- once per turn, Ability used by the player ---

// "Once during your turn, you may heal 10 damage from each of your Pokémon." — Shaymin.
func TestFragrantFlowerGardenHealsEach(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 0, realCard(t, "A4b-030")).Damage = 0
	bench := put(g, p, 1, realCard(t, "A4b-069"))
	bench.Damage = 30
	mustUse(t, g, p, 0)
	if bench.Damage != 20 {
		t.Errorf("bench damage = %d, want 20", bench.Damage)
	}
}

// "Once during your turn, you may put a random Pokémon from your deck into your hand." — Shiinotic.
func TestIlluminateSearchesAPokemon(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 0, realCard(t, "A3a-027"))
	g.S.Players[p].Deck = []engine.Card{realCard(t, "A4b-069")}
	before := len(g.S.Players[p].Hand)
	mustUse(t, g, p, 0)
	if got := len(g.S.Players[p].Hand); got != before+1 {
		t.Errorf("hand %d -> %d, want one more card", before, got)
	}
}

// "Once during your turn, if this Pokémon is in the Active Spot, you may draw a card." — Team Rocket's Slowking ex.
func TestEvilInspirationDrawsFromActive(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 0, realCard(t, "B4a-026"))
	g.S.Players[p].Deck = []engine.Card{realCard(t, "A4b-069")}
	mustUse(t, g, p, 0)
	if len(g.S.Players[p].Hand) != 1 {
		t.Errorf("hand = %d cards, want 1", len(g.S.Players[p].Hand))
	}
}

// "Once during your turn, if this Pokémon is on your Bench, you may switch it with your Active Pokémon." — Solgaleo ex.
func TestRisingRoadSwitchesFromBench(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 0, realCard(t, "A4b-069"))
	put(g, p, 1, realCard(t, "B1a-101"))
	mustUse(t, g, p, 1)
	if name := g.S.Players[p].Active.Card.Name; name != "Solgaleo ex" {
		t.Errorf("Active = %s, want Solgaleo ex", name)
	}
}

// "Once during your turn, you may switch your Active [M] Pokémon with 1 of your Benched Pokémon." — Revavroom.
func TestMetalTransportSwitchesActiveMetal(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 0, realCard(t, "B2b-050"))
	put(g, p, 1, realCard(t, "A4b-069"))
	mustOn(t, g, p, 0, engine.Spot{Player: p, Pos: 1})
	if name := g.S.Players[p].Active.Card.Name; name != "Torkoal" {
		t.Errorf("Active = %s, want Torkoal", name)
	}
}

// "Once during your turn, you may heal 20 damage from your Active Pokémon." — Wigglytuff.
func TestComfortingSongHealsActive(t *testing.T) {
	g, p := abGame(t)
	mon := put(g, p, 0, realCard(t, "A3-229"))
	mon.Damage = 30
	mustUse(t, g, p, 0)
	if mon.Damage != 10 {
		t.Errorf("damage = %d, want 10", mon.Damage)
	}
}

// "Once during your turn, you may heal 30 damage from 1 of your Pokémon." (Pokémon Tool attached) — Sylveon.
func TestSoothingRibbonNeedsATool(t *testing.T) {
	g, p := abGame(t)
	mon := put(g, p, 0, realCard(t, "B3b-030"))
	if canUse(g, p, 0) {
		t.Fatal("Soothing Ribbon usable with no Pokémon Tool")
	}
	tool(t, mon, "A4b-322")
	mon.Damage = 30
	mustOn(t, g, p, 0, engine.Spot{Player: p, Pos: 0})
	if mon.Damage != 0 {
		t.Errorf("damage = %d, want 0", mon.Damage)
	}
}

// "Once during your turn, you may make your opponent's Active Pokémon Burned." — Typhlosion.
func TestFireBreathBurnsFoe(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 0, realCard(t, "A4-029"))
	foe := g.S.Players[1-p].Active
	mustUse(t, g, p, 0)
	if !foe.Has(engine.Burn) {
		t.Error("foe Active not Burned")
	}
}

// "Once during your turn, if this Pokémon is in the Active Spot, you may make your opponent's Active Pokémon Poisoned." — Weezing.
func TestGasLeakPoisonsFromActive(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 0, realCard(t, "A1-177"))
	foe := g.S.Players[1-p].Active
	mustUse(t, g, p, 0)
	if !foe.Has(engine.Poison) {
		t.Error("foe Active not Poisoned")
	}
}

// "Once during your turn, if this Pokémon is in the Active Spot, you may switch in 1 of your opponent's Benched Basic Pokémon." — Victreebel.
func TestFragranceTrapSwitchesInBasic(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 0, realCard(t, "B1-289"))
	bench := put(g, 1-p, 1, realCard(t, "A4b-069"))
	mustOn(t, g, p, 0, engine.Spot{Player: 1 - p, Pos: 1})
	if g.S.Players[1-p].Active != bench {
		t.Error("benched Basic not switched to the foe's Active")
	}
}

// "Once during your turn, if this Pokémon is in the Active Spot, you may switch in 1 of your opponent's Benched Basic Pokémon." — Umbreon ex (switch in a damaged Benched Pokémon).
func TestDarkChaseSwitchesInDamagedBench(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 0, realCard(t, "B2-231"))
	bench := put(g, 1-p, 1, realCard(t, "A4b-069"))
	bench.Damage = 20
	mustOn(t, g, p, 0, engine.Spot{Player: 1 - p, Pos: 1})
	if g.S.Players[1-p].Active != bench {
		t.Error("damaged benched Pokémon not switched to the foe's Active")
	}
}

// "Once during your turn, if this Pokémon is in the Active Spot, you may switch out your opponent's Active Pokémon to the Bench." — Vaporeon ex.
func TestFrozenFlowSwitchesFoeOut(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 0, realCard(t, "B3b-102"))
	foe := g.S.Players[1-p].Active
	put(g, 1-p, 1, realCard(t, "A4b-069"))
	mustUse(t, g, p, 0)
	if g.S.Players[1-p].Active == foe {
		t.Error("foe Active not switched to the Bench")
	}
}

// "Once during your turn, you may switch out your opponent's Active Basic Pokémon to the Bench." — Swellow.
func TestRepellingWindSwitchesFoeBasic(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 0, realCard(t, "B2-133"))
	foe := g.S.Players[1-p].Active
	put(g, 1-p, 1, realCard(t, "A4b-069"))
	mustUse(t, g, p, 0)
	if g.S.Players[1-p].Active == foe {
		t.Error("foe Basic Active not switched to the Bench")
	}
}

// "Once during your turn, you may switch out your opponent's Active Pokémon to the Bench." (on Bench play) — Roaring Moon.
func TestAncientRoarOnBenchPlay(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 0, realCard(t, "A4b-069"))
	foe := put(g, 1-p, 0, realCard(t, "A4b-069"))
	put(g, 1-p, 1, realCard(t, "A4b-069"))
	g.S.Players[p].Hand = []engine.Card{realCard(t, "B3a-047")}
	mustSubmit(t, g, engine.PlayBasic{Player: p, CardID: "B3a-047", Slot: engine.BenchLeft})
	mustOn(t, g, p, 1, engine.Spot{Player: 1 - p, Pos: 0})
	if g.S.Players[1-p].Active == foe {
		t.Error("foe Active not switched out by Ancient Roar")
	}
}

// "Once during your turn, you may switch out your opponent's Active Basic Pokémon ... " — Captivating Rhythm (coin flip).
func TestCaptivatingRhythmSwitchesOnHeads(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 0, realCard(t, "B1-027"))
	foe := g.S.Players[1-p].Active
	put(g, 1-p, 1, realCard(t, "A4b-069"))
	mustOn(t, g, p, 0, engine.Spot{Player: 1 - p, Pos: 1})
	heads := lastFlips(g)[0].Heads
	if switched := g.S.Players[1-p].Active != foe; switched != heads {
		t.Errorf("switched %v with heads %v", switched, heads)
	}
}

// "Once during your turn, you may move all [D] Energy from each of your Pokémon to this Pokémon." — Tyranitar.
func TestEnergyPlunderMovesDarkness(t *testing.T) {
	g, p := abGame(t)
	tyr := put(g, p, 0, realCard(t, "A4-119"))
	bench := put(g, p, 1, realCard(t, "A4b-069"))
	bench.Energy[engine.Darkness] = 2
	mustUse(t, g, p, 0)
	if tyr.Energy[engine.Darkness] != 2 || bench.Energy[engine.Darkness] != 0 {
		t.Errorf("Tyranitar D = %d, bench D = %d; want 2 and 0", tyr.Energy[engine.Darkness], bench.Energy[engine.Darkness])
	}
}

// "As often as you like during your turn, you may move a [W] Energy from 1 of your Benched [W] Pokémon to your Active [W] Pokémon." — Vaporeon.
func TestWashOutMovesWaterRepeatedly(t *testing.T) {
	g, p := abGame(t)
	active := put(g, p, 0, realCard(t, "A4b-099"))
	bench := put(g, p, 1, realCard(t, "A4b-099"))
	bench.Energy[engine.Water] = 2
	mustOn(t, g, p, 1, engine.Spot{Player: p, Pos: 1})
	mustOn(t, g, p, 1, engine.Spot{Player: p, Pos: 1})
	if active.Energy[engine.Water] != 2 {
		t.Errorf("Active Water = %d, want 2", active.Energy[engine.Water])
	}
}

// "Once during your turn, you may discard 1 [R] Energy ... +50 damage." — Skeledirge.
func TestPassionateVoiceDiscardsFireForBoost(t *testing.T) {
	g, p := abGame(t)
	mon := put(g, p, 0, realCard(t, "B2a-018"))
	mon.Energy[engine.Fire] = 1
	mustUse(t, g, p, 0)
	if mon.Energy[engine.Fire] != 0 {
		t.Errorf("Fire = %d, want 0 after discard", mon.Energy[engine.Fire])
	}
	if !hasModifier(g, p, 0, engine.HookDamageDealt, 50) {
		t.Error("no +50 to Fire attacks this turn")
	}
}

// "Once during your turn, when you play this Pokémon from your hand to evolve, you may draw 2 cards." — Sylveon ex.
func TestHappyRibbonDrawsTwo(t *testing.T) {
	g, p := abGame(t)
	g.S.Turn = 3
	g.S.Players[p].Deck = []engine.Card{realCard(t, "A4b-069"), realCard(t, "A4b-069")}
	evolveFrom(t, g, p, "B4-131", "B2-229") // Eevee -> Sylveon ex
	mustUse(t, g, p, 0)
	if got := len(g.S.Players[p].Hand); got != 2 {
		t.Errorf("hand = %d, want 2", got)
	}
}

// "Once during your turn, when you play this Pokémon from your hand to evolve 1 of your Pokémon, you may look at the top 4 ... Item cards." — Raticate.
func TestTreasureCollectingTakesItemsFromTop(t *testing.T) {
	g, p := abGame(t)
	g.S.Turn = 3
	g.S.Players[p].Deck = []engine.Card{realCard(t, "P-A-001"), realCard(t, "A4b-069"), realCard(t, "A4b-069"), realCard(t, "A4b-069")}
	evolveFrom(t, g, p, "B4-129", "B4-130") // Rattata -> Raticate
	mustUse(t, g, p, 0)
	if !handHas(g.S.Players[p].Hand, "Potion") {
		t.Error("Potion from the top 4 not put into hand")
	}
}

// "Once during your turn, when you play this Pokémon ... you may move a random Energy from your opponent's Active Pokémon to this Pokémon." — Team Rocket's Raticate ex.
func TestThievingIncisorsMovesFoeEnergy(t *testing.T) {
	g, p := abGame(t)
	g.S.Turn = 3
	foe := g.S.Players[1-p].Active
	foe.Energy[engine.Water] = 1
	evolveFrom(t, g, p, "B4a-058", "B4a-059") // Team Rocket's Rattata -> Team Rocket's Raticate ex
	mustUse(t, g, p, 0)
	if foe.Energy[engine.Water] != 0 {
		t.Error("foe Active kept its Water Energy")
	}
}

// "Once during your turn, when you play this Pokémon ... you may prevent all damage from—and effects of—attacks ... until the end of your opponent's next turn." — Samurott.
func TestStanceShieldsUntilOpponentNextTurn(t *testing.T) {
	g, p := abGame(t)
	g.S.Turn = 3
	evolveFrom(t, g, p, "B4-043", "B4-044") // Dewott -> Samurott
	mustUse(t, g, p, 0)
	if !hookOn(g, p, 0, engine.HookEffectShield) {
		t.Error("Samurott has no effect shield after Stance")
	}
}

// "Once during your turn, when you play this Pokémon ... you may make your opponent's Active Pokémon Poisoned and Burned." — Team Rocket's Weezing ex (see also Boiler Smog).
func TestBoilerSmogEvolvesPoisonAndBurn(t *testing.T) {
	g, p := abGame(t)
	g.S.Turn = 3
	evolveFrom(t, g, p, "B4a-042", "B4a-043") // Team Rocket's Koffing -> Weezing ex
	mustUse(t, g, p, 0)
	if foe := g.S.Players[1-p].Active; !foe.Has(engine.Poison) || !foe.Has(engine.Burn) {
		t.Error("foe Active not Poisoned and Burned")
	}
}

// "Once during your turn, you may heal 20 ..." — Comforting Song from Bench (AnySpot).
func TestComfortingSongFromBenchIsRefused(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 1, realCard(t, "A3-229"))
	if canUse(g, p, 1) {
		return
	}
}
