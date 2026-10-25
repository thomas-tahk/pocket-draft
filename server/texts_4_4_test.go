package main

import (
	"slices"
	"testing"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// One test per wording in batch 4.4 (issue #77), each on a real printing.

// abGame is a game past setup, Torkoal against Torkoal; p is on turn.
func abGame(t *testing.T) (*engine.Game, int) {
	t.Helper()
	useRealCards(t)
	g, p := trainerGame(t, realCard(t, "A4b-069"))
	g.S.Players[p].Hand = nil
	return g, p
}

func use(g *engine.Game, p, pos int, target engine.Spot) error {
	return g.Submit(engine.UseAbility{Player: p, Pos: pos, Target: target})
}

func mustUse(t *testing.T, g *engine.Game, p, pos int) {
	t.Helper()
	if err := use(g, p, pos, engine.Spot{}); err != nil {
		t.Fatalf("use ability at %d: %v", pos, err)
	}
}

func canUse(g *engine.Game, p, pos int) bool {
	for _, ev := range g.LegalMoves(p) {
		if u, ok := ev.(engine.UseAbility); ok && u.Pos == pos {
			return true
		}
	}
	return false
}

func hasModifier(g *engine.Game, p, pos int, hook engine.Hook, value int) bool {
	return slices.ContainsFunc(g.Modifiers(p, pos), func(m engine.Modifier) bool { return m.Hook == hook && m.Value == value })
}

func cardOfKind(t *testing.T, k engine.CardKind) engine.Card {
	t.Helper()
	id := map[engine.CardKind]string{engine.KindTool: "A4b-322", engine.KindSupporter: "A1-219"}[k]
	return realCard(t, id)
}

// --- loading ---

// A mapped wording loads onto every printing under its own name.
func TestAbilityWordingLoadsOnEveryPrinting(t *testing.T) {
	useRealCards(t)
	for _, id := range []string{"A2a-071", "B1-328"} { // Arceus ex
		c := realCard(t, id)
		if c.Ability == nil || c.Ability.Name != "Fabled Luster" {
			t.Errorf("%s ability = %+v, want Fabled Luster", id, c.Ability)
		}
	}
	for _, id := range []string{"A4a-044", "A3-109"} { // Donphan Exoskeleton, Alolan Persian Fur Coat: one wording
		if c := realCard(t, id); c.Ability == nil || len(c.Ability.Passive) == 0 {
			t.Errorf("%s has no working Ability", id)
		}
	}
}

// --- always-on ---

// "This Pokémon can't be affected by any Special Conditions." — Arceus ex.
func TestFabledLusterBlocksConditions(t *testing.T) {
	g, p := abGame(t)
	attacker := put(g, p, 0, realCard(t, "B2b-010")) // Ponyta, Singe: Burned
	attacker.Energy[engine.Fire] = 1
	put(g, 1-p, 0, realCard(t, "A2a-071"))

	mustSubmit(t, g, engine.UseAttack{Player: p, Index: 0})

	if g.S.Players[1-p].Active.Has(engine.Burn) {
		t.Error("Arceus ex was Burned")
	}
}

// "If you have Arceus or Arceus ex in play, attacks used by this Pokémon cost 1 less [C]." — Abomasnow.
func TestVigorLinkCutsOneColorless(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 0, realCard(t, "A2a-021"))
	if hasModifier(g, p, 0, engine.HookColorlessCost, -1) {
		t.Fatal("cost cut without Arceus")
	}

	put(g, p, 1, realCard(t, "A2a-071"))

	if !hasModifier(g, p, 0, engine.HookColorlessCost, -1) {
		t.Error("no cost cut with Arceus ex in play")
	}
}

// "If you have Arceus or Arceus ex in play, attacks used by this Pokémon do +30 damage." — Carnivine.
func TestPowerLinkAddsThirty(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 0, realCard(t, "A2a-009"))
	put(g, p, 1, realCard(t, "A2a-071"))

	if !hasModifier(g, p, 0, engine.HookDamageDealt, 30) {
		t.Error("no +30 with Arceus ex in play")
	}
}

// "Attacks used by your [P] Pokémon and [M] Pokémon do +30 damage." — Aegislash.
func TestCursedMetalAddsThirtyToPsychicAndMetal(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 1, realCard(t, "B1a-102"))
	put(g, p, 0, realCard(t, "B4-106")) // Beldum [M]
	metal := hasModifier(g, p, 0, engine.HookDamageDealt, 30)
	put(g, p, 0, realCard(t, "A4b-069")) // Torkoal [R]

	if !metal || hasModifier(g, p, 0, engine.HookDamageDealt, 30) {
		t.Errorf("Metal got +30: %v; Fire got +30: %v", metal, !metal)
	}
}

// "If a Stadium is in play, this Pokémon has no Retreat Cost." — Alolan Raichu.
func TestSurgeSurferNeedsAStadium(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 0, realCard(t, "B2-050"))
	before := g.RetreatCost(p)
	g.S.Stadium = &engine.Card{Name: "Some Stadium"}

	if before != 2 || g.RetreatCost(p) != 0 {
		t.Errorf("retreat cost %d without a Stadium, %d with; want 2 and 0", before, g.RetreatCost(p))
	}
}

// "Your opponent's Active Pokémon's Retreat Cost is 1 more." — Ariados.
func TestTrapTerritoryRaisesRetreatCost(t *testing.T) {
	g, p := abGame(t)
	put(g, 1-p, 0, realCard(t, "B4-106")) // Beldum, retreat 2
	before := g.RetreatCost(1 - p)
	put(g, p, 1, realCard(t, "B1a-006"))

	if got := g.RetreatCost(1 - p); before != 2 || got != 3 {
		t.Errorf("retreat cost %d then %d, want 2 then 3", before, got)
	}
}

// "This Pokémon takes −30 damage from attacks." — Armarouge ex (a hit of 30 does none).
func TestArmorTakesThirtyLess(t *testing.T) {
	g, p := abGame(t)
	atk := put(g, p, 0, realCard(t, "B1a-006")) // Ariados, Pierce 30
	atk.Energy[engine.Grass] = 1
	def := put(g, 1-p, 0, realCard(t, "B3a-105"))

	mustSubmit(t, g, engine.UseAttack{Player: p, Index: 0})

	if def.Damage != 0 {
		t.Errorf("Armarouge ex took %d, want 0", def.Damage)
	}
}

// "This Pokémon takes −30 damage from attacks from [R] or [W] Pokémon." — Azumarill.
func TestThickFatOnlyAgainstFireAndWater(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 0, realCard(t, "A4-050"))
	put(g, 1-p, 0, realCard(t, "A4b-069")) // Torkoal [R]
	fire := hasModifier(g, p, 0, engine.HookDamageTaken, -30)
	put(g, 1-p, 0, realCard(t, "B1a-006")) // Ariados [G]

	if !fire || hasModifier(g, p, 0, engine.HookDamageTaken, -30) {
		t.Error("−30 should apply against Fire and not against Grass")
	}
}

// "This Pokémon takes −10 damage from attacks." — Cloyster.
func TestShellArmorTakesTenLess(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 0, realCard(t, "A1-067"))
	if !hasModifier(g, p, 0, engine.HookDamageTaken, -10) {
		t.Error("no −10")
	}
}

// "This Pokémon takes −20 damage from attacks." — Donphan.
func TestExoskeletonTakesTwentyLess(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 0, realCard(t, "A4a-044"))
	if !hasModifier(g, p, 0, engine.HookDamageTaken, -20) {
		t.Error("no −20")
	}
}

// "If this Pokémon has full HP, it takes −40 damage from attacks." — Eiscue.
func TestIceFaceNeedsFullHP(t *testing.T) {
	g, p := abGame(t)
	mon := put(g, p, 0, realCard(t, "B1-080"))
	full := hasModifier(g, p, 0, engine.HookDamageTaken, -40)
	mon.Damage = 10

	if !full || hasModifier(g, p, 0, engine.HookDamageTaken, -40) {
		t.Error("−40 should apply at full HP only")
	}
}

// "If you have another Beldum in play, this Pokémon's Retreat Cost is 2 less." — Beldum.
func TestConductiveBodyNeedsAnotherBeldum(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 0, realCard(t, "B4-106"))
	alone := g.RetreatCost(p)
	put(g, p, 1, realCard(t, "B4-106"))

	if alone != 2 || g.RetreatCost(p) != 0 {
		t.Errorf("retreat cost %d alone, %d with another; want 2 and 0", alone, g.RetreatCost(p))
	}
}

// "As long as this Pokémon is on your Bench, your Active [D] Pokémon's Retreat Cost is 1 less." — Bombirdier.
func TestVillainousDeliveryCutsDarkRetreat(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 0, realCard(t, "B2b-040")) // Darkrai [D], retreat 2
	before := g.RetreatCost(p)
	put(g, p, 1, realCard(t, "B3-115"))

	if before != 2 || g.RetreatCost(p) != 1 {
		t.Errorf("retreat cost %d then %d, want 2 then 1", before, g.RetreatCost(p))
	}
}

// "If you have another Falinks in play, +20 damage and −20 damage taken." — Falinks.
func TestCoordinatedUnit(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 0, realCard(t, "B4-215"))
	if hasModifier(g, p, 0, engine.HookDamageDealt, 20) {
		t.Fatal("bonus alone")
	}
	put(g, p, 1, realCard(t, "B4-215"))

	if !hasModifier(g, p, 0, engine.HookDamageDealt, 20) || !hasModifier(g, p, 0, engine.HookDamageTaken, -20) {
		t.Error("with another Falinks: want +20 dealt and −20 taken")
	}
}

// "As long as this Pokémon is in the Active Spot, it can evolve during your first turn or the turn you play it." — Eevee.
func TestBoostedEvolutionOnlyInTheActiveSpot(t *testing.T) {
	g, p := abGame(t)
	g.S.Players[p].Hand = []engine.Card{realCard(t, "A4-083")} // Espeon ex
	canEvolve := func() bool {
		return slices.ContainsFunc(g.LegalMoves(p), func(ev engine.Event) bool { _, ok := ev.(engine.Evolve); return ok })
	}
	put(g, p, 0, realCard(t, "B1-184"))
	active := canEvolve()
	g.S.Players[p].Active, g.S.Players[p].Bench[0] = nil, g.S.Players[p].Active
	put(g, p, 0, realCard(t, "A4b-069"))

	if !active || canEvolve() {
		t.Errorf("Active Eevee can evolve: %v; Benched Eevee can evolve: %v", active, canEvolve())
	}
}

// "Pokémon (both yours and your opponent's) can't be healed." — Claydol.
func TestHealBlockStopsHealing(t *testing.T) {
	g, p := abGame(t)
	mon := put(g, p, 0, realCard(t, "A4-083")) // Espeon ex, Psychic Healing
	mon.Damage = 40
	put(g, 1-p, 1, realCard(t, "A3a-031"))

	if err := use(g, p, 0, engine.Spot{Player: p, Pos: 0}); err != nil {
		t.Fatal(err)
	}

	if mon.Damage != 40 {
		t.Errorf("damage = %d, want 40 (healing blocked)", mon.Damage)
	}
}

// "Each of your Pokémon that has any [P] Energy attached ... can't be affected by any Special Conditions." — Comfey.
func TestFlowerShieldProtectsPsychicEnergyHolders(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 1, realCard(t, "A3-080"))
	holder := put(g, 1-p, 0, realCard(t, "A4b-069"))
	_ = holder
	mine := put(g, p, 0, realCard(t, "A4b-069"))
	mine.Energy[engine.Psychic] = 1
	foe := put(g, 1-p, 0, realCard(t, "B2b-010"))
	foe.Energy[engine.Fire] = 1
	// the foe's turn: swap sides so it attacks my Psychic-Energy holder
	g.S.Players[p].Hand = nil
	mustSubmit(t, g, engine.EndTurn{Player: p})
	mustSubmit(t, g, engine.UseAttack{Player: 1 - p, Index: 0}) // Singe

	if mine.Has(engine.Burn) {
		t.Error("a Pokémon with [P] Energy was Burned")
	}
}

// --- once during your turn ---

// "Once during your turn, you may put a random Pokémon Tool card from your deck into your hand." — Ambipom.
func TestCatchingTailFindsATool(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 0, realCard(t, "A3b-059"))
	tool := cardOfKind(t, engine.KindTool)
	g.S.Players[p].Deck = append([]engine.Card{realCard(t, "A4b-069")}, tool)

	mustUse(t, g, p, 0)

	if h := g.S.Players[p].Hand; len(h) != 1 || h[0].Kind != engine.KindTool {
		t.Errorf("hand = %v, want the Tool", h)
	}
	if canUse(g, p, 0) {
		t.Error("usable twice in a turn")
	}
}

// "Heal 60 damage from 1 of your Pokémon ex that has any Energy attached. If you do, discard a random Energy." — Arboliva.
func TestExtraHealHealsAnExAndDiscardsEnergy(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 1, realCard(t, "B2b-116"))
	ex := put(g, p, 0, realCard(t, "A2a-071"))
	ex.Damage, ex.Energy[engine.Fire] = 80, 1

	if err := use(g, p, 1, engine.Spot{Player: p, Pos: 0}); err != nil {
		t.Fatal(err)
	}

	if ex.Damage != 20 || ex.Energy[engine.Fire] != 0 || len(g.S.Players[p].DiscardedEnergy) != 1 {
		t.Errorf("damage %d, Fire %d, discarded %d; want 20, 0, 1", ex.Damage, ex.Energy[engine.Fire], len(g.S.Players[p].DiscardedEnergy))
	}
}

// "Take a [W] Energy from your Energy Zone and attach it to the [W] Pokémon in the Active Spot." — Baxcalibur.
func TestIceMakerAttachesWaterToAWaterActive(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 1, realCard(t, "B2a-036"))
	water := put(g, p, 0, realCard(t, "A1-067"))
	fire := canUse(g, p, 1)

	mustUse(t, g, p, 1)
	put(g, p, 0, realCard(t, "A4b-069")).Energy[engine.Fire] = 0

	if !fire || water.Energy[engine.Water] != 1 {
		t.Errorf("usable %v, Water %d; want true, 1", fire, water.Energy[engine.Water])
	}
	g.S.Players[p].Bench[0].AbilityUsedTurn = 0
	if canUse(g, p, 1) {
		t.Error("usable with a Fire Active")
	}
}

// "Once during your turn, you may remove a random Special Condition from your Active Pokémon." — Blissey.
func TestHappinessSupplementCuresACondition(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 1, realCard(t, "B3-128"))
	active := put(g, p, 0, realCard(t, "A4b-069"))
	if canUse(g, p, 1) {
		t.Fatal("usable with no condition")
	}
	active.Inflict(engine.Burn)

	mustUse(t, g, p, 1)

	if active.Has(engine.Burn) {
		t.Error("still Burned")
	}
}

// "...you may move 30 damage that your Active Pokémon has on it to this Pokémon." — Brambleghast.
func TestAcceptPainMovesThirty(t *testing.T) {
	g, p := abGame(t)
	bramble := put(g, p, 1, realCard(t, "B3-073"))
	active := put(g, p, 0, realCard(t, "A4b-069"))
	active.Damage = 40

	mustUse(t, g, p, 1)

	if active.Damage != 10 || bramble.Damage != 30 {
		t.Errorf("Active %d, Brambleghast %d; want 10 and 30", active.Damage, bramble.Damage)
	}
}

// "Once during your turn, you may heal 20 damage from each of your Pokémon." — Butterfree.
func TestPowderHealHealsEveryPokemon(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 0, realCard(t, "A1-007")).Damage = 30
	other := put(g, p, 1, realCard(t, "A4b-069"))
	other.Damage = 20

	mustUse(t, g, p, 0)

	if g.S.Players[p].Active.Damage != 10 || other.Damage != 0 {
		t.Errorf("damage %d and %d, want 10 and 0", g.S.Players[p].Active.Damage, other.Damage)
	}
}

// "Once during your turn, you may switch your Active Ultra Beast with 1 of your Benched Ultra Beasts." — Celesteela.
func TestUltraThrustersSwitchesUltraBeasts(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 0, realCard(t, "A3a-062"))
	bench := put(g, p, 1, realCard(t, "A3a-075"))
	put(g, p, 2, realCard(t, "A4b-069")) // not an Ultra Beast

	if err := use(g, p, 0, engine.Spot{Player: p, Pos: 2}); err == nil {
		t.Fatal("switched with a non-Ultra Beast")
	}
	if err := use(g, p, 0, engine.Spot{Player: p, Pos: 1}); err != nil {
		t.Fatal(err)
	}

	if g.S.Players[p].Active != bench {
		t.Error("the Benched Celesteela is not the Active")
	}
}

// "Once during your turn, you may discard the top card of your opponent's deck." — Chandelure.
func TestSlowSearDiscardsTheTopCard(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 0, realCard(t, "B1-043"))
	before := len(g.S.Players[1-p].Deck)

	mustUse(t, g, p, 0)

	if len(g.S.Players[1-p].Deck) != before-1 || len(g.S.Players[1-p].Discard) != 1 {
		t.Error("the top card did not go to the discard pile")
	}
}

// "...if you have Arceus or Arceus ex in play, you may do 30 damage to your opponent's Active Pokémon." — Crobat.
func TestCunningLinkNeedsArceus(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 0, realCard(t, "A4b-230"))
	without := canUse(g, p, 0)
	put(g, p, 1, realCard(t, "A2a-071"))

	mustUse(t, g, p, 0)

	if without || g.S.Players[1-p].Active.Damage != 30 {
		t.Errorf("usable without Arceus: %v; damage %d, want false and 30", without, g.S.Players[1-p].Active.Damage)
	}
}

// "...if this Pokémon is on your Bench, you may attach an Energy from your discard pile to your Active [N] Pokémon." — Dragonair.
func TestDragonsBlessingAttachesFromDiscard(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 1, realCard(t, "B4-117"))
	dragon := put(g, p, 0, realCard(t, "B4-175"))
	g.S.Players[p].DiscardedEnergy = []engine.Energy{engine.Fire}

	mustUse(t, g, p, 1)

	if dragon.Energy[engine.Fire] != 1 || len(g.S.Players[p].DiscardedEnergy) != 0 {
		t.Error("the Energy did not move from the discard pile to the Dragon")
	}
}

// "Once during your turn, you may use this Ability. 1 Special Condition from among Burned, Confused, and Poisoned..." — Dustox.
func TestVarietyPowderPutsOneCondition(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 0, realCard(t, "B4-005"))
	foe := g.S.Players[1-p].Active

	mustUse(t, g, p, 0)

	if n := len(foe.Conditions()); n != 1 {
		t.Errorf("%d conditions, want 1", n)
	}
}

// "...if this Pokémon is in the Active Spot, you may heal 30 damage from 1 of your Pokémon." — Espeon ex.
func TestPsychicHealingHealsThirty(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 0, realCard(t, "A4-083")).Damage = 50
	bench := put(g, p, 1, realCard(t, "A4b-069"))
	bench.Damage = 40

	if err := use(g, p, 0, engine.Spot{Player: p, Pos: 1}); err != nil {
		t.Fatal(err)
	}

	if bench.Damage != 10 || g.S.Players[p].Active.Damage != 50 {
		t.Errorf("bench %d, Active %d; want 10 and 50", bench.Damage, g.S.Players[p].Active.Damage)
	}
}

// "...attach a [R] Energy from your discard pile to this Pokémon. If you do, do 20 damage to this Pokémon." — Flareon ex.
func TestCombustAttachesFireAndHurts(t *testing.T) {
	g, p := abGame(t)
	flareon := put(g, p, 0, realCard(t, "B2-225"))
	if canUse(g, p, 0) {
		t.Fatal("usable with no Fire in the discard pile")
	}
	g.S.Players[p].DiscardedEnergy = []engine.Energy{engine.Fire}

	mustUse(t, g, p, 0)

	if flareon.Energy[engine.Fire] != 1 || flareon.Damage != 20 {
		t.Errorf("Fire %d, damage %d; want 1 and 20", flareon.Energy[engine.Fire], flareon.Damage)
	}
}

// "As often as you like during your turn, ... move all of its damage to this Pokémon." — Dusknoir.
func TestShadowVoidMovesDamageAsOftenAsYouLike(t *testing.T) {
	g, p := abGame(t)
	dusk := put(g, p, 0, realCard(t, "A2-072"))
	a, b := put(g, p, 1, realCard(t, "A4b-069")), put(g, p, 2, realCard(t, "A4b-069"))
	a.Damage, b.Damage = 30, 20

	if err := use(g, p, 0, engine.Spot{Player: p, Pos: 1}); err != nil {
		t.Fatal(err)
	}
	if err := use(g, p, 0, engine.Spot{Player: p, Pos: 2}); err != nil {
		t.Fatal(err)
	}

	if dusk.Damage != 50 || a.Damage != 0 || b.Damage != 0 {
		t.Errorf("Dusknoir %d, others %d and %d; want 50, 0, 0", dusk.Damage, a.Damage, b.Damage)
	}
}

// --- when you play this Pokémon to evolve ---

func evolved(g *engine.Game, p, pos int, c engine.Card) *engine.InPlay {
	mon := put(g, p, pos, c)
	mon.PlacedTurn = g.S.Turn
	return mon
}

// "...when you play this Pokémon from your hand to evolve 1 of your Pokémon, you may take a [R] Energy..." — Charmeleon.
func TestIgnitionAttachesFireAfterEvolving(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 1, realCard(t, "B2b-008"))
	fire := put(g, p, 0, realCard(t, "A4b-069"))
	if canUse(g, p, 1) {
		t.Fatal("usable by a Charmeleon that didn't just evolve")
	}
	g.S.Players[p].Bench[0].PlacedTurn = g.S.Turn

	mustUse(t, g, p, 1)

	if fire.Energy[engine.Fire] != 1 {
		t.Error("no Fire Energy on the Active")
	}
}

// "...you may discard a random Energy from your opponent's Active Pokémon." — Crawdaunt.
func TestUnrulyClawDiscardsOppEnergy(t *testing.T) {
	g, p := abGame(t)
	evolved(g, p, 0, realCard(t, "A4b-104"))
	g.S.Players[1-p].Active.Energy[engine.Fire] = 1

	mustUse(t, g, p, 0)

	if g.S.Players[1-p].Active.Energy[engine.Fire] != 0 {
		t.Error("Energy not discarded")
	}
}

// "...you may put a Supporter card from your discard pile into your hand." — Delcatty.
func TestSearchForFriendsTakesASupporter(t *testing.T) {
	g, p := abGame(t)
	evolved(g, p, 0, realCard(t, "B1-194"))
	g.S.Players[p].Discard = []engine.Card{realCard(t, "A4b-069"), cardOfKind(t, engine.KindSupporter)}

	mustUse(t, g, p, 0)

	if h := g.S.Players[p].Hand; len(h) != 1 || h[0].Kind != engine.KindSupporter {
		t.Errorf("hand = %v, want the Supporter", h)
	}
}

// "...you may do 20 damage to your opponent's Active Pokémon." — Drizzile.
func TestSwiftShotDoesTwenty(t *testing.T) {
	g, p := abGame(t)
	evolved(g, p, 0, realCard(t, "B3-049"))

	mustUse(t, g, p, 0)

	if g.S.Players[1-p].Active.Damage != 20 {
		t.Errorf("damage = %d, want 20", g.S.Players[1-p].Active.Damage)
	}
}

// "...you may put 2 random Pokémon Tool cards from your discard pile into your hand." — Galarian Perrserker.
func TestDigUpTakesTwoTools(t *testing.T) {
	g, p := abGame(t)
	evolved(g, p, 0, realCard(t, "B4-219"))
	tool := cardOfKind(t, engine.KindTool)
	g.S.Players[p].Discard = []engine.Card{tool, tool, tool}

	mustUse(t, g, p, 0)

	if len(g.S.Players[p].Hand) != 2 || len(g.S.Players[p].Discard) != 1 {
		t.Errorf("hand %d, discard %d; want 2 and 1", len(g.S.Players[p].Hand), len(g.S.Players[p].Discard))
	}
}

// --- run by the engine ---

// "At the end of your turn, if this Pokémon is in the Active Spot, draw a card." — Entei ex.
func TestLegendaryPulseDrawsAtTheEndOfYourTurn(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 0, realCard(t, "A4a-010"))
	before := len(g.S.Players[p].Hand)

	mustSubmit(t, g, engine.EndTurn{Player: p})

	if got := len(g.S.Players[p].Hand); got != before+1 {
		t.Errorf("hand %d -> %d, want one more card", before, got)
	}
}

// "At the end of your opponent's turn, if this Pokémon is in the Active Spot, put a random card ... that evolves from this Pokémon..." — Caterpie.
func TestQuickGrowthEvolvesAtTheEndOfTheOpponentsTurn(t *testing.T) {
	g, p := abGame(t)
	put(g, 1-p, 0, realCard(t, "B3b-001"))
	g.S.Players[1-p].Deck = []engine.Card{realCard(t, "A4b-069"), realCard(t, "B3b-002")}

	mustSubmit(t, g, engine.EndTurn{Player: p})

	if name := g.S.Players[1-p].Active.Card.Name; name != "Metapod" {
		t.Errorf("Active = %s, want Metapod", name)
	}
}

// "At the end of each turn, if your opponent's Active Pokémon is Asleep, do 20 damage to that Pokémon." — Darkrai.
func TestBadDreamsHurtsASleepingActive(t *testing.T) {
	sawAsleep := false
	for seed := uint64(1); seed <= 12; seed++ {
		g, p := abGame(t)
		put(g, p, 0, realCard(t, "B2b-040"))
		foe := g.S.Players[1-p].Active
		foe.Inflict(engine.Asleep)

		mustSubmit(t, g, engine.EndTurn{Player: p}) // the Checkup flips for Sleep first

		want := 0
		if foe.Has(engine.Asleep) {
			want, sawAsleep = 20, true
		}
		if foe.Damage != want {
			t.Fatalf("asleep %v, damage %d, want %d", foe.Has(engine.Asleep), foe.Damage, want)
		}
		_ = seed
	}
	if !sawAsleep {
		t.Skip("the Sleep flip woke it every time")
	}
}

// "If this Pokémon is in the Active Spot and is damaged by an attack ... the Attacking Pokémon is now Poisoned." — Dragalge ex.
func TestPoisonPointPoisonsTheAttacker(t *testing.T) {
	g, p := abGame(t)
	atk := put(g, p, 0, realCard(t, "B1a-006"))
	atk.Energy[engine.Grass] = 1
	put(g, 1-p, 0, realCard(t, "B3-231"))

	mustSubmit(t, g, engine.UseAttack{Player: p, Index: 0})

	if !atk.Has(engine.Poison) {
		t.Error("the attacker was not Poisoned")
	}
}

// "...is damaged by an attack from your opponent's Pokémon, do 20 damage to the Attacking Pokémon." — Druddigon.
func TestRoughSkinHurtsTheAttacker(t *testing.T) {
	g, p := abGame(t)
	atk := put(g, p, 0, realCard(t, "B1a-006"))
	atk.Energy[engine.Grass] = 1
	put(g, 1-p, 0, realCard(t, "A1a-056"))

	mustSubmit(t, g, engine.UseAttack{Player: p, Index: 0})

	if atk.Damage != 20 {
		t.Errorf("attacker damage = %d, want 20", atk.Damage)
	}
}

// "Whenever you attach a [D] Energy from your Energy Zone to this Pokémon, do 20 damage to your opponent's Active Pokémon." — Darkrai ex.
func TestNightmareAuraDamagesOnDarkAttach(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 0, realCard(t, "A4b-245"))
	g.S.Players[p].EnergyZone = engine.Darkness

	mustSubmit(t, g, engine.AttachEnergy{Player: p, Target: 0})

	if g.S.Players[1-p].Active.Damage != 20 {
		t.Errorf("damage = %d, want 20", g.S.Players[1-p].Active.Damage)
	}
}

// "Whenever you attach a [P] Energy from your Energy Zone to this Pokémon, heal 20 damage from this Pokémon." — Cresselia ex.
func TestLunarPlumageHealsOnPsychicAttach(t *testing.T) {
	g, p := abGame(t)
	put(g, p, 0, realCard(t, "P-A-037")).Damage = 50
	g.S.Players[p].EnergyZone = engine.Psychic

	mustSubmit(t, g, engine.AttachEnergy{Player: p, Target: 0})

	if g.S.Players[p].Active.Damage != 30 {
		t.Errorf("damage = %d, want 30", g.S.Players[p].Active.Damage)
	}
}

// UseAbility crosses the wire with its target and back.
func TestUseAbilityRequestRoundTrips(t *testing.T) {
	ev := engine.UseAbility{Player: 1, Pos: 2, Target: engine.Spot{Player: 1, Pos: 0}}

	got, err := toEvent(fromEvent(ev))

	if err != nil || got != ev {
		t.Errorf("round trip = %#v, %v; want %#v", got, err, ev)
	}
}
