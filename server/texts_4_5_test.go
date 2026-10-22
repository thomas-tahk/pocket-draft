package main

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// Ability texts, batch 4.5 (issue #78): one test per wording, on a real card.

// abilityGame45 is a game past setup with the real card id as player p's
// Active (pos 0) or Bench slot (pos n), no Trainer in hand.
func abilityGame45(t *testing.T, id string) (*engine.Game, int, *engine.InPlay) {
	t.Helper()
	useRealCards(t)
	g, p := trainerGame(t, realCard(t, "A4b-373"))
	return g, p, onBoard(t, g, p, 0, id)
}

func useAbility(t *testing.T, g *engine.Game, p, pos int, target engine.Spot) {
	t.Helper()
	if err := g.Submit(engine.UseAbility{Player: p, Pos: pos, Target: target}); err != nil {
		t.Fatalf("use ability: %v", err)
	}
}

func abilityRefused(t *testing.T, g *engine.Game, p, pos int, target engine.Spot) {
	t.Helper()
	if g.Submit(engine.UseAbility{Player: p, Pos: pos, Target: target}) == nil {
		t.Fatal("ability was allowed when it shouldn't be")
	}
}

// Every wording in this batch is the exact printed text of some card.
func TestAbilityTexts45AreRealWordings(t *testing.T) {
	raw, err := os.ReadFile(cardsPath)
	if err != nil {
		t.Fatal(err)
	}
	var cards []rawCard
	if err := json.Unmarshal(raw, &cards); err != nil {
		t.Fatal(err)
	}
	printed := map[string]bool{}
	for _, c := range cards {
		if c.Ability != nil {
			printed[c.Ability.Effect] = true
		}
	}
	for wording := range abilityTexts45 {
		if !printed[wording] {
			t.Errorf("no card prints %q", wording)
		}
	}
}

func TestPsyShadowAttachesToAPsychicActive(t *testing.T) {
	g, p, active := abilityGame45(t, "A1-132") // Gardevoir
	useAbility(t, g, p, 0, engine.Spot{})
	if active.Energy[engine.Psychic] != 1 {
		t.Errorf("Psychic energy = %d, want 1", active.Energy[engine.Psychic])
	}
	abilityRefused(t, g, p, 0, engine.Spot{}) // once per turn
}

func TestPsyShadowNeedsAPsychicActive(t *testing.T) {
	g, p, _ := abilityGame45(t, "A4b-069") // Torkoal in the Active
	onBoard(t, g, p, 1, "A1-132")          // Gardevoir on the Bench
	abilityRefused(t, g, p, 1, engine.Spot{})
}

func TestBrokenSpaceBellowAttachesThenEndsTurn(t *testing.T) {
	g, p, mon := abilityGame45(t, "A2b-035") // Giratina ex
	useAbility(t, g, p, 0, engine.Spot{})
	if mon.Energy[engine.Psychic] != 1 {
		t.Errorf("Psychic energy = %d, want 1", mon.Energy[engine.Psychic])
	}
	if g.S.Active == p {
		t.Error("turn did not end")
	}
}

func TestZoneToSelfAbilities(t *testing.T) {
	cases := []struct {
		name, id string
		energy   engine.Energy
	}{{"Volt Charge", "A1-098", engine.Lightning}, {"Aqua Charge", "B3b-015", engine.Water}}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g, p, mon := abilityGame45(t, tc.id)
			useAbility(t, g, p, 0, engine.Spot{})
			if mon.Energy[tc.energy] != 1 {
				t.Errorf("%s energy = %d, want 1", tc.energy, mon.Energy[tc.energy])
			}
		})
	}
}

func TestRoarInUnison(t *testing.T) {
	g, p, mon := abilityGame45(t, "B1-157") // Hydreigon
	useAbility(t, g, p, 0, engine.Spot{})
	if mon.Energy[engine.Darkness] != 2 || mon.Damage != 30 {
		t.Errorf("Darkness %d, damage %d; want 2 and 30", mon.Energy[engine.Darkness], mon.Damage)
	}
}

func TestForestBreathAttachesToAGrassPokemon(t *testing.T) {
	g, p, _ := abilityGame45(t, "A2a-010") // Leafeon ex
	bench := onBoard(t, g, p, 1, "A1-001") // Bulbasaur
	useAbility(t, g, p, 0, engine.Spot{Player: p, Pos: 1})
	if bench.Energy[engine.Grass] != 1 {
		t.Errorf("Grass energy on the Bench = %d, want 1", bench.Energy[engine.Grass])
	}
}

func TestForestBreathNeedsTheActiveSpot(t *testing.T) {
	g, p, _ := abilityGame45(t, "A4b-069")
	onBoard(t, g, p, 1, "A2a-010")
	abilityRefused(t, g, p, 1, engine.Spot{Player: p, Pos: 1})
}

func TestCoinAbilitiesInflictOnHeads(t *testing.T) {
	cases := []struct {
		name, id string
		hit      func(*engine.InPlay) bool
	}{
		{"Poison Coating", "A2b-051", func(m *engine.InPlay) bool { return m.Poisoned }},
		{"Sleep Pendulum", "A1-125", func(m *engine.InPlay) bool { return m.Has(engine.Asleep) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			hits := 0
			for seed := int64(1); seed <= 40; seed++ {
				g, p, _ := abilityGame45(t, tc.id)
				useAbility(t, g, p, 0, engine.Spot{})
				heads := lastFlips(g)[0].Heads
				if got := tc.hit(g.S.Players[1-p].Active); got != heads {
					t.Fatalf("heads=%v but condition=%v", heads, got)
				}
				if heads {
					hits++
				}
				_ = seed
			}
			if hits == 0 {
				t.Error("never flipped heads in 40 games")
			}
		})
	}
}

func TestPerplexingEarsConfuses(t *testing.T) {
	g, p, _ := abilityGame45(t, "B3-066") // Meowstic
	useAbility(t, g, p, 0, engine.Spot{})
	if !g.S.Players[1-p].Active.Has(engine.Confused) {
		t.Error("opponent's Active is not Confused")
	}
}

func TestWaterShurikenHitsAnyOpposingPokemon(t *testing.T) {
	g, p, _ := abilityGame45(t, "A1-089") // Greninja
	bench := onBoard(t, g, 1-p, 1, "A1-001")
	useAbility(t, g, p, 0, engine.Spot{Player: 1 - p, Pos: 1})
	if bench.Damage != 20 {
		t.Errorf("Benched damage = %d, want 20", bench.Damage)
	}
}

func TestWaterShurikenKnocksOutABenchedPokemon(t *testing.T) {
	g, p, _ := abilityGame45(t, "A1-089")
	bench := onBoard(t, g, 1-p, 1, "A1-001")
	bench.Damage = bench.Card.HP - 10
	useAbility(t, g, p, 0, engine.Spot{Player: 1 - p, Pos: 1})
	if g.S.Players[1-p].Bench[0] != nil || g.S.Players[p].Points != 1 {
		t.Errorf("bench %v, points %d; want knocked out and 1 point", g.S.Players[1-p].Bench[0], g.S.Players[p].Points)
	}
}

func TestShiftingStreamSwitchesAWaterActive(t *testing.T) {
	g, p, _ := abilityGame45(t, "B1-073") // Greninja ex (Water)
	bench := onBoard(t, g, p, 1, "A4b-069")
	useAbility(t, g, p, 0, engine.Spot{Player: p, Pos: 1})
	if g.S.Players[p].Active != bench {
		t.Error("the Benched Pokémon was not switched in")
	}
}

func TestWatchOverHealsTheActive(t *testing.T) {
	g, p, mon := abilityGame45(t, "B1-121") // Indeedee ex
	mon.Damage = 50
	useAbility(t, g, p, 0, engine.Spot{})
	if mon.Damage != 30 {
		t.Errorf("damage = %d, want 30", mon.Damage)
	}
}

func TestMelodiousHealingHealsEachWaterPokemon(t *testing.T) {
	g, p, mon := abilityGame45(t, "A3-048") // Primarina
	other := onBoard(t, g, p, 1, "B1-073")
	skip := onBoard(t, g, p, 2, "A4b-069")
	mon.Damage, other.Damage, skip.Damage = 40, 50, 40
	useAbility(t, g, p, 0, engine.Spot{})
	if mon.Damage != 10 || other.Damage != 20 || skip.Damage != 40 {
		t.Errorf("damage %d, %d, %d; want 10, 20, 40", mon.Damage, other.Damage, skip.Damage)
	}
}

func TestPsychicConnectMovesBenchedPsychicEnergy(t *testing.T) {
	g, p, active := abilityGame45(t, "A3-087") // Lunala ex
	bench := onBoard(t, g, p, 1, "A1-132")
	bench.Energy[engine.Psychic] = 2
	useAbility(t, g, p, 0, engine.Spot{Player: p, Pos: 1})
	if bench.Energy[engine.Psychic] != 0 || active.Energy[engine.Psychic] != 2 {
		t.Errorf("bench %d, active %d; want 0 and 2", bench.Energy[engine.Psychic], active.Energy[engine.Psychic])
	}
}

func TestDriveOffSwitchesOutTheOpposingActive(t *testing.T) {
	g, p, _ := abilityGame45(t, "A1-188") // Pidgeot
	foeBench := onBoard(t, g, 1-p, 1, "A1-001")
	onBoard(t, g, 1-p, 2, "A4b-069") // two choices, so the opponent is asked
	useAbility(t, g, p, 0, engine.Spot{})
	if g.S.Players[1-p].Active != nil || g.S.Pending == nil {
		t.Fatalf("active %v pending %v", g.S.Players[1-p].Active != nil, g.S.Pending)
	}
	if err := g.Submit(engine.ChooseNewActive{Player: 1 - p, BenchIndex: 0}); err != nil {
		t.Fatal(err)
	}
	if g.S.Players[1-p].Active != foeBench {
		t.Error("the chosen Benched Pokémon is not Active")
	}
}

func TestDismantlingKeysDiscardsToolThenItself(t *testing.T) {
	g, p, _ := abilityGame45(t, "A4b-069")
	onBoard(t, g, p, 1, "B1-120") // Klefki on the Bench
	foe := g.S.Players[1-p].Active
	abilityRefused(t, g, p, 1, engine.Spot{}) // no Tool to discard
	tool := realCard(t, "A2-147")             // Giant Cape
	foe.Tool = &tool
	useAbility(t, g, p, 1, engine.Spot{})
	if foe.Tool != nil || g.S.Players[p].Bench[0] != nil {
		t.Errorf("tool %v, Klefki %v; want both gone", foe.Tool, g.S.Players[p].Bench[0])
	}
}

func TestDataScanIsAvailable(t *testing.T) {
	g, p, _ := abilityGame45(t, "A1-209") // Porygon
	useAbility(t, g, p, 0, engine.Spot{})
	if g.S.Pending == nil {
		t.Error("no look at the top card was offered")
	}
}

func TestLevitateNeedsEnergy(t *testing.T) {
	g, p, mon := abilityGame45(t, "A2-078") // Giratina
	if cost := g.RetreatCost(p); cost == 0 {
		t.Fatal("Giratina retreats free with no Energy")
	}
	mon.Energy[engine.Psychic] = 1
	if cost := g.RetreatCost(p); cost != 0 {
		t.Errorf("retreat cost = %d with Energy, want 0", cost)
	}
}

func TestFluffyFlightFreesTheActive(t *testing.T) {
	g, p, _ := abilityGame45(t, "A4b-069") // Torkoal, which has a Retreat Cost
	before := g.RetreatCost(p)
	onBoard(t, g, p, 1, "A4-015") // Jumpluff on the Bench
	if before == 0 || g.RetreatCost(p) != 0 {
		t.Errorf("retreat cost %d → %d, want >0 → 0", before, g.RetreatCost(p))
	}
}

func TestDamageTakenAbilities(t *testing.T) {
	cases := []struct {
		name, id string
		foe      string
		want     int
	}{
		{"Hard Coat", "A1-182", "A4b-069", -20},
		{"Thick Fat vs Fire", "A2-032", "A4b-069", -20},
		{"Thick Fat vs Grass", "A2-032", "A1-001", 0},
		{"Safeguard vs ex", "A3-066", "A2a-010", -999},
		{"Safeguard vs non-ex", "A3-066", "A4b-069", 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g, p, _ := abilityGame45(t, tc.id)
			onBoard(t, g, 1-p, 0, tc.foe)
			if got := modifier(g, p, 0, engine.HookDamageTaken); got != tc.want {
				t.Errorf("damage taken modifier = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestIntimidatingFangWeakensTheOpposingActive(t *testing.T) {
	g, p, _ := abilityGame45(t, "A3a-015") // Luxray
	if got := modifier(g, 1-p, 0, engine.HookDamageDealt); got != -20 {
		t.Errorf("opponent's damage modifier = %d, want -20", got)
	}
	if got := modifier(g, p, 0, engine.HookDamageDealt); got != 0 {
		t.Errorf("own damage modifier = %d, want 0", got)
	}
}

func TestFightingCoachBoostsFightingAttackers(t *testing.T) {
	g, p, _ := abilityGame45(t, "A2-092") // Lucario
	onBoard(t, g, p, 1, "A4b-069")
	if got := modifier(g, p, 0, engine.HookDamageDealt); got != 20 {
		t.Errorf("Fighting Active modifier = %d, want 20", got)
	}
	if got := modifier(g, p, 1, engine.HookDamageDealt); got != 0 {
		t.Errorf("Fire Bench modifier = %d, want 0", got)
	}
}

func TestFutureSystemCutsFutureAttackCosts(t *testing.T) {
	g, p, _ := abilityGame45(t, "B3a-027") // Iron Valiant
	if got := modifier(g, p, 0, engine.HookColorlessCost); got != -1 {
		t.Errorf("cost modifier = %d, want -1", got)
	}
}

func TestTorrentNeedsLowHP(t *testing.T) {
	g, p, mon := abilityGame45(t, "B2a-024") // Quaquaval
	if got := modifier(g, p, 0, engine.HookDamageDealt); got != 0 {
		t.Errorf("healthy modifier = %d, want 0", got)
	}
	mon.Damage = mon.Card.HP - 50
	if got := modifier(g, p, 0, engine.HookDamageDealt); got != 60 {
		t.Errorf("low-HP modifier = %d, want 60", got)
	}
}

func TestLordlyCheeringBoostsPoliwhirlEvolutions(t *testing.T) {
	g, p, _ := abilityGame45(t, "A1-061") // Poliwrath, evolves from Poliwhirl
	if got := modifier(g, p, 0, engine.HookDamageDealt); got != 0 {
		t.Fatalf("no Politoed yet: modifier = %d", got)
	}
	onBoard(t, g, p, 1, "A4-040") // Politoed on the Bench
	if got := modifier(g, p, 0, engine.HookDamageDealt); got != 40 {
		t.Errorf("modifier = %d, want 40", got)
	}
}

// --- evolve, bench and Checkup wordings ---

// evolvedGame puts the real card id on the Active as if it evolved this turn.
func evolvedGame(t *testing.T, id string) (*engine.Game, int, *engine.InPlay) {
	g, p, mon := abilityGame45(t, id)
	mon.PlacedTurn = g.S.Turn
	return g, p, mon
}

func TestSwiftShotHitsTheOpposingActive(t *testing.T) {
	g, p, _ := evolvedGame(t, "B3-050") // Inteleon
	useAbility(t, g, p, 0, engine.Spot{})
	if d := g.S.Players[1-p].Active.Damage; d != 30 {
		t.Errorf("damage = %d, want 30", d)
	}
}

func TestSwiftShotNeedsAnEvolutionThisTurn(t *testing.T) {
	g, p, _ := abilityGame45(t, "B3-050")
	abilityRefused(t, g, p, 0, engine.Spot{})
}

func TestHealingRipplesHealsAWaterPokemon(t *testing.T) {
	g, p, mon := evolvedGame(t, "A4a-022") // Milotic
	mon.Damage = 80
	useAbility(t, g, p, 0, engine.Spot{Player: p, Pos: 0})
	if mon.Damage != 20 {
		t.Errorf("damage = %d, want 20", mon.Damage)
	}
}

func TestRefreshingTeaRedrawsTheOpposingHand(t *testing.T) {
	g, p, _ := evolvedGame(t, "B2-075") // Polteageist
	useAbility(t, g, p, 0, engine.Spot{})
	if n := len(g.S.Players[1-p].Hand); n != 3 {
		t.Errorf("opponent's hand = %d cards, want 3 (points still needed)", n)
	}
}

func TestEvoshockParalyzesOnHeads(t *testing.T) {
	heads := 0
	for i := 0; i < 40; i++ {
		g, p, _ := evolvedGame(t, "B4-050") // Raichu
		useAbility(t, g, p, 0, engine.Spot{})
		got := g.S.Players[1-p].Active.Has(engine.Paralyzed)
		if got != lastFlips(g)[0].Heads {
			t.Fatalf("paralyzed %v but heads %v", got, lastFlips(g)[0].Heads)
		}
		if got {
			heads++
		}
	}
	if heads == 0 {
		t.Error("never flipped heads in 40 games")
	}
}

func benchedThisTurn(t *testing.T, id string) (*engine.Game, int, *engine.InPlay) {
	g, p, _ := abilityGame45(t, "A4b-069")
	mon := onBoard(t, g, p, 1, id)
	mon.PlacedTurn = g.S.Turn
	return g, p, mon
}

func TestLegendaryDriveSwitchesInAndTakesAllEnergy(t *testing.T) {
	g, p, mon := benchedThisTurn(t, "B3a-019") // Miraidon ex
	g.S.Players[p].Active.Energy[engine.Fire] = 2
	useAbility(t, g, p, 1, engine.Spot{})
	if g.S.Players[p].Active != mon || mon.Energy[engine.Fire] != 2 {
		t.Errorf("active ok %v, Fire %d; want switched in with 2", g.S.Players[p].Active == mon, mon.Energy[engine.Fire])
	}
}

func TestLegendaryDriveOnlyWhenJustBenched(t *testing.T) {
	g, p, mon := benchedThisTurn(t, "B3a-019")
	mon.PlacedTurn = g.S.Turn - 2
	abilityRefused(t, g, p, 1, engine.Spot{})
}

func TestInfiltratingInspectionIsAvailableWhenJustBenched(t *testing.T) {
	g, p, _ := benchedThisTurn(t, "A4a-032") // Misdreavus
	useAbility(t, g, p, 1, engine.Spot{})
}

func TestHospitalityHealsAnActiveGrassPokemon(t *testing.T) {
	g, p, _ := benchedThisTurn(t, "B4-017") // Poltchageist
	grass := onBoard(t, g, p, 0, "A1-001")
	grass.Damage = 30
	useAbility(t, g, p, 1, engine.Spot{})
	if grass.Damage != 10 {
		t.Errorf("damage = %d, want 10", grass.Damage)
	}
}

func TestBlessedSaltHealsEachPokemonAtCheckup(t *testing.T) {
	g, p, mon := abilityGame45(t, "B3a-033") // Garganacl
	mon.Damage = 30
	endTurn(t, g)
	if mon.Damage != 20 {
		t.Errorf("damage = %d, want 20", mon.Damage)
	}
	_ = p
}

func TestSnowyTerrainDamagesTheOpposingActiveAtCheckup(t *testing.T) {
	g, p, _ := abilityGame45(t, "A2a-022") // Glaceon ex
	foe := g.S.Players[1-p].Active
	endTurn(t, g)
	if foe.Damage != 10 {
		t.Errorf("damage = %d, want 10", foe.Damage)
	}
}

func TestSpeedLinkNeedsArceus(t *testing.T) {
	g, p, _ := abilityGame45(t, "A2a-013") // Heatran
	if g.RetreatCost(p) == 0 {
		t.Fatal("Heatran retreats free with no Arceus")
	}
	onBoard(t, g, p, 1, "A2a-070") // Arceus
	if g.RetreatCost(p) != 0 {
		t.Errorf("retreat cost = %d with Arceus, want 0", g.RetreatCost(p))
	}
}

func TestFantasticalFloatingNeedsLatias(t *testing.T) {
	g, p, _ := abilityGame45(t, "A4a-037") // Latios
	if g.RetreatCost(p) == 0 {
		t.Fatal("Latios retreats free with no Latias")
	}
	onBoard(t, g, p, 1, "A4a-036")
	if g.RetreatCost(p) != 0 {
		t.Errorf("retreat cost = %d with Latias, want 0", g.RetreatCost(p))
	}
}
