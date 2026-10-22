package main

import (
	"testing"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// Tool and Stadium texts, batch 3.13 (issue #73): one test per wording, on a
// real card.

const (
	cubone   = "B4a-035" // [F] Basic, 60 HP, Bone Beatdown [F] 20, weak to [G]
	volbeat  = "B4a-001" // [G] Basic, 80 HP, Tackle [G] 30
	pikachu  = "B3b-021" // [L] Basic, 60 HP
	cufant   = "B4a-046" // [M] Basic, 90 HP, Retreat Cost 3
	magneton = "B3-053"  // [L] Stage 1 from Magnemite
	torkoal  = "A4b-069" // [R] Basic, 100 HP, Retreat Cost 2
)

// duel is a game past setup where player p's Active is atk, ready to attack
// with its first attack, and the opponent's Active is def.
func duel(t *testing.T, atk, def string) (*engine.Game, int, *engine.InPlay, *engine.InPlay) {
	t.Helper()
	useRealCards(t)
	g, p := trainerGame(t, realCard(t, "A4b-373"))
	a, d := onBoard(t, g, p, 0, atk), onBoard(t, g, 1-p, 0, def)
	for _, c := range a.Card.Attacks[0].Cost {
		a.Energy[c] = 1
	}
	if a.Energy[engine.Colorless] > 0 {
		a.Energy[engine.Colorless] = 0
		a.Energy[engine.Fighting] += 3
	}
	return g, p, a, d
}

func tool(t *testing.T, mon *engine.InPlay, id string) {
	t.Helper()
	c := realCard(t, id)
	mon.Tool = &c
}

func stadiumIn(t *testing.T, g *engine.Game, id string) {
	t.Helper()
	c := realCard(t, id)
	g.S.Stadium = &c
}

func useAttack(t *testing.T, g *engine.Game, p, index int) {
	t.Helper()
	if err := g.Submit(engine.UseAttack{Player: p, Index: index}); err != nil {
		t.Fatalf("attack: %v", err)
	}
}

func useStadium(g *engine.Game, p int, pick string) error {
	return g.Submit(engine.UseStadium{Player: p, CardID: pick})
}

// --- HP ---

func TestAncientBoosterEnergyCapsule(t *testing.T) { // B3a-069
	g, p, a, _ := duel(t, "B2a-063", cubone) // Koraidon
	tool(t, a, "B3a-069")
	if got := g.MaxHP(p, 0); got != 160 {
		t.Errorf("Koraidon HP = %d, want 160", got)
	}
	b := onBoard(t, g, p, 1, torkoal)
	tool(t, b, "B3a-069")
	if got := g.MaxHP(p, 1); got != 100 {
		t.Errorf("Torkoal (not Ancient) HP = %d, want 100", got)
	}
}

func TestGiantCape(t *testing.T) { // A2-147
	g, p, a, _ := duel(t, torkoal, cubone)
	tool(t, a, "A2-147")
	if got := g.MaxHP(p, 0); got != 120 {
		t.Errorf("HP = %d, want 120", got)
	}
}

func TestGiantCapeHPDecidesTheKnockout(t *testing.T) {
	g, p, _, d := duel(t, cubone, cubone)
	tool(t, d, "A2-147") // 60 + 20
	d.Damage = 50
	useAttack(t, g, p, 0)
	if g.S.Players[1-p].Active != d || d.Damage != 70 {
		t.Errorf("Cubone with a Giant Cape at 70/80 damage was Knocked Out or lost its damage (%d)", d.Damage)
	}
}

func TestLeafCape(t *testing.T) { // A3-147
	g, p, a, _ := duel(t, volbeat, cubone)
	tool(t, a, "A3-147")
	b := onBoard(t, g, p, 1, torkoal)
	tool(t, b, "A3-147")
	if g.MaxHP(p, 0) != 110 || g.MaxHP(p, 1) != 100 {
		t.Errorf("HP = %d / %d, want 110 / 100", g.MaxHP(p, 0), g.MaxHP(p, 1))
	}
}

func TestElegantCape(t *testing.T) { // B3b-065
	g, p, a, _ := duel(t, magneton, cubone)
	tool(t, a, "B3b-065")
	b := onBoard(t, g, p, 1, volbeat)
	tool(t, b, "B3b-065")
	if g.MaxHP(p, 0) != 120 || g.MaxHP(p, 1) != 80 {
		t.Errorf("HP = %d / %d, want 120 / 80", g.MaxHP(p, 0), g.MaxHP(p, 1))
	}
}

func TestStartingPlains(t *testing.T) { // B2-154
	g, p, _, _ := duel(t, volbeat, cubone)
	onBoard(t, g, p, 1, magneton)
	stadiumIn(t, g, "B2-154")
	if g.MaxHP(p, 0) != 100 || g.MaxHP(1-p, 0) != 80 || g.MaxHP(p, 1) != 90 {
		t.Errorf("HP = %d / %d / %d, want 100 / 80 / 90", g.MaxHP(p, 0), g.MaxHP(1-p, 0), g.MaxHP(p, 1))
	}
}

// --- damage dealt and taken ---

func TestFutureBoosterEnergyCapsule(t *testing.T) { // B3a-070
	g, p, a, _ := duel(t, "B2a-045", cubone) // Miraidon
	tool(t, a, "B3a-070")
	if got := modifier(g, p, 0, engine.HookDamageDealt); got != 20 {
		t.Errorf("Miraidon does %+d, want +20", got)
	}
	tool(t, a, "A2-147")
	onBoard(t, g, p, 0, torkoal).Tool = a.Tool
	if got := modifier(g, p, 0, engine.HookDamageDealt); got != 0 {
		t.Errorf("Torkoal does %+d, want 0", got)
	}
}

func TestBeastite(t *testing.T) { // A3a-066
	g, p, a, _ := duel(t, "A4b-047", cubone) // Kartana, an Ultra Beast
	tool(t, a, "A3a-066")
	g.S.Players[p].Points = 2
	if got := modifier(g, p, 0, engine.HookDamageDealt); got != 20 {
		t.Errorf("with 2 points Kartana does %+d, want +20", got)
	}
}

func TestArenaOfAntiquity(t *testing.T) { // B3-154
	g, p, _, _ := duel(t, cubone, "A3a-006") // against Buzzwole ex
	stadiumIn(t, g, "B3-154")
	if got := modifier(g, p, 0, engine.HookDamageDealt); got != 20 {
		t.Errorf("Cubone does %+d to an ex, want +20", got)
	}
	onBoard(t, g, 1-p, 0, torkoal)
	if got := modifier(g, p, 0, engine.HookDamageDealt); got != 0 {
		t.Errorf("Cubone does %+d to a non-ex, want 0", got)
	}
}

func TestTrainingArea(t *testing.T) { // B2-153
	g, p, _, _ := duel(t, magneton, cubone)
	stadiumIn(t, g, "B2-153")
	onBoard(t, g, 1-p, 0, magneton)
	if modifier(g, p, 0, engine.HookDamageDealt) != 10 || modifier(g, 1-p, 0, engine.HookDamageDealt) != 10 {
		t.Error("each Stage 1 should do +10")
	}
	onBoard(t, g, p, 0, volbeat)
	if got := modifier(g, p, 0, engine.HookDamageDealt); got != 0 {
		t.Errorf("a Basic does %+d, want 0", got)
	}
}

func TestBoundedField(t *testing.T) { // B3-155
	g, p, _, d := duel(t, volbeat, cubone)
	d.Card.HP = 200
	stadiumIn(t, g, "B3-155")
	useAttack(t, g, p, 0)
	if d.Damage != 60 { // Tackle 30, Weakness ×2
		t.Errorf("damage = %d, want 60", d.Damage)
	}
}

func TestHeavyHelmet(t *testing.T) { // B1-219
	g, p, a, _ := duel(t, cufant, cubone)
	tool(t, a, "B1-219")
	if got := modifier(g, p, 0, engine.HookDamageTaken); got != -20 {
		t.Errorf("Cufant (Retreat Cost 3) takes %+d, want -20", got)
	}
	b := onBoard(t, g, p, 0, volbeat)
	tool(t, b, "B1-219")
	if got := modifier(g, p, 0, engine.HookDamageTaken); got != 0 {
		t.Errorf("Volbeat (Retreat Cost 1) takes %+d, want 0", got)
	}
}

func TestMetalCoreBarrier(t *testing.T) { // B2-148
	g, p, a, _ := duel(t, cufant, cubone)
	tool(t, a, "B2-148")
	if got := modifier(g, p, 0, engine.HookDamageTaken); got != -50 {
		t.Errorf("takes %+d, want -50", got)
	}
	endTurn(t, g) // mine
	if a.Tool == nil {
		t.Fatal("discarded at the end of my own turn")
	}
	endTurn(t, g) // the opponent's
	if a.Tool != nil {
		t.Error("still attached after the opponent's turn ended")
	}
}

func TestSteelApron(t *testing.T) { // A4-153
	g, p, a, d := duel(t, cubone, cufant)
	a.Card.Attacks[0].Effect = []engine.EffectOp{engine.ApplyStatus{Status: engine.Poison}}
	tool(t, d, "A4-153")
	d.Inflict(engine.Poison)
	if got := modifier(g, 1-p, 0, engine.HookDamageTaken); got != -10 {
		t.Errorf("takes %+d, want -10", got)
	}
	useAttack(t, g, p, 0) // the turn's end cures the Poison; the new one can't stick
	if d.Has(engine.Poison) {
		t.Error("the Poisoned [M] Pokémon with a Steel Apron is still Poisoned")
	}
}

func TestClearVeil(t *testing.T) { // B4-149
	g, p, a, d := duel(t, cubone, cufant)
	a.Card.Attacks[0].Effect = []engine.EffectOp{engine.ApplyStatus{Status: engine.Poison}}
	a.Energy[engine.Fighting] = 1
	tool(t, d, "B4-149")
	useAttack(t, g, p, 0)
	if d.Damage != 20 || d.Has(engine.Poison) {
		t.Errorf("damage %d, Poisoned %v; want 20 and not Poisoned", d.Damage, d.Has(engine.Poison))
	}
}

func TestProtectivePoncho(t *testing.T) { // B2-147
	g, p, _, _ := duel(t, cubone, volbeat)
	b := onBoard(t, g, p, 1, volbeat)
	tool(t, b, "B2-147")
	if !g.BenchShielded(p, 1) {
		t.Error("a Benched Pokémon with a Poncho should be shielded")
	}
	tool(t, g.S.Players[p].Active, "B2-147")
	if g.BenchShielded(p, 0) {
		t.Error("the Active Spot isn't shielded")
	}
}

// --- retreat ---

func TestBigAirBalloon(t *testing.T) { // B2a-087
	g, p, a, _ := duel(t, "B4a-005", cubone) // Serperior, Stage 2
	tool(t, a, "B2a-087")
	if got := g.RetreatCost(p); got != 0 {
		t.Errorf("Retreat Cost = %d, want 0", got)
	}
	onBoard(t, g, p, 0, torkoal).Tool = a.Tool
	if got := g.RetreatCost(p); got != 2 {
		t.Errorf("a Basic's Retreat Cost = %d, want 2", got)
	}
}

func TestSmallBalloon(t *testing.T) { // B3b-064
	g, p, a, _ := duel(t, torkoal, cubone)
	tool(t, a, "B3b-064")
	if got := g.RetreatCost(p); got != 1 {
		t.Errorf("Retreat Cost = %d, want 1", got)
	}
	onBoard(t, g, p, 0, magneton).Tool = a.Tool
	if got := g.RetreatCost(p); got != 2 {
		t.Errorf("a Stage 1's Retreat Cost = %d, want 2", got)
	}
}

func TestInflatableBoat(t *testing.T) { // A4a-067
	g, p, a, _ := duel(t, "B4a-013", cubone) // Team Rocket's Lapras [W], Retreat 1
	tool(t, a, "A4a-067")
	if got := g.RetreatCost(p); got != 0 {
		t.Errorf("Retreat Cost = %d, want 0", got)
	}
	onBoard(t, g, p, 0, torkoal).Tool = a.Tool
	if got := g.RetreatCost(p); got != 2 {
		t.Errorf("a [R] Pokémon's Retreat Cost = %d, want 2", got)
	}
}

func TestPeculiarPlaza(t *testing.T) { // B2-155
	g, p, _, _ := duel(t, "B4a-025", torkoal) // Team Rocket's Slowpoke [P], Retreat 2
	stadiumIn(t, g, "B2-155")
	if got := g.RetreatCost(p); got != 0 {
		t.Errorf("[P] Retreat Cost = %d, want 0", got)
	}
	if got := g.RetreatCost(1 - p); got != 2 {
		t.Errorf("[R] Retreat Cost = %d, want 2", got)
	}
}

func TestMemoryLight(t *testing.T) { // A4a-068
	g, p, a, d := duel(t, magneton, cubone)
	tool(t, a, "A4a-068")
	atks := a.Attacks()
	if len(atks) != 2 || atks[1].Name != "Lightning Ball" {
		t.Fatalf("attacks = %+v, want Magneton's plus Magnemite's Lightning Ball", atks)
	}
	a.Energy = map[engine.Energy]int{engine.Lightning: 1}
	useAttack(t, g, p, 1)
	if d.Damage != 20 {
		t.Errorf("Lightning Ball did %d, want 20", d.Damage)
	}
}

// --- end of turn ---

func TestSoothingShore(t *testing.T) { // B4-154
	g, p, a, _ := duel(t, "B4a-013", cubone) // Lapras [W]
	stadiumIn(t, g, "B4-154")
	a.Energy[engine.Water] = 1
	a.Damage = 30
	other := onBoard(t, g, p, 1, volbeat)
	other.Damage = 30
	endTurn(t, g)
	if a.Damage != 10 || other.Damage != 30 {
		t.Errorf("damage %d / %d, want 10 / 30", a.Damage, other.Damage)
	}
}

func TestHikingTrail(t *testing.T) { // B2b-069
	g, p, _, _ := duel(t, cubone, volbeat)
	stadiumIn(t, g, "B2b-069")
	pl := &g.S.Players[p]
	pl.Hand = pl.Hand[:1]
	endTurn(t, g)
	if len(pl.Hand) != 3 {
		t.Errorf("hand = %d, want 3", len(pl.Hand))
	}
	foe := &g.S.Players[1-p]
	foe.Hand = foe.Hand[:2]
	endTurn(t, g)
	if len(foe.Hand) != 3 {
		t.Errorf("opponent's hand = %d, want 3", len(foe.Hand))
	}
}

func TestLeftovers(t *testing.T) { // A3b-067
	g, _, a, _ := duel(t, cubone, volbeat)
	tool(t, a, "A3b-067")
	a.Damage = 30
	endTurn(t, g)
	if a.Damage != 20 {
		t.Errorf("damage = %d, want 20", a.Damage)
	}
	endTurn(t, g) // not my turn's end
	if a.Damage != 20 {
		t.Errorf("healed at the end of the opponent's turn: %d", a.Damage)
	}
}

func TestDeceptiveNeedle(t *testing.T) { // B4-148
	g, _, a, d := duel(t, "B3a-046", cubone) // Iron Jugulis [D]
	tool(t, a, "B4-148")
	endTurn(t, g)
	if d.Damage != 10 {
		t.Errorf("opponent's Active damage = %d, want 10", d.Damage)
	}
}

func TestLumBerry(t *testing.T) { // A2-149
	g, _, a, _ := duel(t, cubone, volbeat)
	tool(t, a, "A2-149")
	a.Inflict(engine.Poison)
	endTurn(t, g)
	if a.Has(engine.Poison) || a.Tool != nil {
		t.Errorf("Poisoned %v, Tool %v; want cured and discarded", a.Has(engine.Poison), a.Tool != nil)
	}
}

func TestSitrusBerry(t *testing.T) { // B1-218
	g, _, a, _ := duel(t, volbeat, cubone)
	tool(t, a, "B1-218")
	a.Damage = 30 // 50 of 80 left
	endTurn(t, g)
	if a.Damage != 30 || a.Tool == nil {
		t.Fatalf("healed above half: damage %d", a.Damage)
	}
	endTurn(t, g)
	a.Damage = 40 // exactly half
	endTurn(t, g)
	if a.Damage != 10 || a.Tool != nil {
		t.Errorf("damage %d, Tool %v; want 10 and discarded", a.Damage, a.Tool != nil)
	}
}

// --- when damaged ---

func TestDarkPendant(t *testing.T) { // A4-154
	g, p, _, d := duel(t, cubone, "B3a-046") // Iron Jugulis [D]
	tool(t, d, "A4-154")
	atk := g.S.Players[p]
	hand, deck := len(atk.Hand), len(atk.Deck)
	useAttack(t, g, p, 0)
	now := g.S.Players[p]
	if len(now.Hand) != hand-1 || len(now.Deck) != deck+1 {
		t.Errorf("attacker's hand %d→%d, deck %d→%d; want one card shuffled back", hand, len(now.Hand), deck, len(now.Deck))
	}
}

func TestPoisonBarb(t *testing.T) { // A3-146
	g, p, a, d := duel(t, cubone, volbeat)
	tool(t, d, "A3-146")
	useAttack(t, g, p, 0)
	if !a.Has(engine.Poison) {
		t.Error("the attacker should be Poisoned")
	}
}

func TestRockyHelmet(t *testing.T) { // A2-148
	g, p, a, d := duel(t, cubone, volbeat)
	tool(t, d, "A2-148")
	useAttack(t, g, p, 0)
	if a.Damage != 20 {
		t.Errorf("attacker damage = %d, want 20", a.Damage)
	}
}

// --- when Knocked Out ---

// knockOut has the attacker's Cubone Knock Out a Pikachu with the Tool id and
// two benched Volbeat, and returns the game, attacker and the Pikachu.
func knockOut(t *testing.T, id string) (*engine.Game, int, *engine.InPlay) {
	t.Helper()
	g, p, _, d := duel(t, cubone, pikachu)
	tool(t, d, id)
	d.Damage = 50
	d.Energy[engine.Lightning] = 2
	onBoard(t, g, 1-p, 1, volbeat)
	onBoard(t, g, 1-p, 2, volbeat)
	useAttack(t, g, p, 0)
	return g, p, d
}

func TestElectricalCord(t *testing.T) { // A3a-065
	g, p, _ := knockOut(t, "A3a-065")
	b := g.S.Players[1-p].Bench
	if b[0].Energy[engine.Lightning] != 1 || b[1].Energy[engine.Lightning] != 1 {
		t.Errorf("benched [L] Energy = %d / %d, want 1 / 1", b[0].Energy[engine.Lightning], b[1].Energy[engine.Lightning])
	}
}

func TestLuckyEgg(t *testing.T) { // B3-148
	g, p, _, d := duel(t, cubone, pikachu)
	tool(t, d, "B3-148")
	d.Damage = 50
	g.S.Players[1-p].Hand = g.S.Players[1-p].Hand[:1]
	useAttack(t, g, p, 0)
	if got := len(g.S.Players[1-p].Hand); got != 5 {
		t.Errorf("hand = %d, want 5", got)
	}
}

func TestLuckyMittens(t *testing.T) { // B1-220
	g, p, a, d := duel(t, cubone, pikachu)
	tool(t, a, "B1-220")
	d.Damage = 50
	hand := len(g.S.Players[p].Hand)
	useAttack(t, g, p, 0)
	if got := len(g.S.Players[p].Hand); got != hand+1 {
		t.Errorf("hand = %d, want %d", got, hand+1)
	}
}

func TestRescueScarf(t *testing.T) { // A4-155
	g, p, d := knockOut(t, "A4-155")
	pl := g.S.Players[1-p]
	inHand := false
	for _, c := range pl.Hand {
		inHand = inHand || c.Name == d.Card.Name
	}
	for _, c := range pl.Discard {
		if c.Name == d.Card.Name {
			t.Error("the Knocked Out Pokémon was discarded")
		}
	}
	if !inHand {
		t.Error("the Knocked Out Pokémon isn't in hand")
	}
}

// --- Stadium uses ---

// seeded is a game past setup with the given seed where player p's turn it is.
func seeded(t *testing.T, seed uint64) (*engine.Game, int) {
	t.Helper()
	useRealCards(t)
	basic := realCard(t, torkoal)
	g := engine.NewGame(seed, deckOfCard(basic, 20), deckOfCard(basic, 20))
	doSetup(t, g)
	return g, g.S.Active
}

func TestRainbowCave(t *testing.T) { // B4-155
	g, p := seeded(t, 1)
	stadiumIn(t, g, "B4-155")
	pl := &g.S.Players[p]
	pl.EnergyZone, pl.EnergyNext = engine.Fire, engine.Water
	if err := useStadium(g, p, ""); err != nil {
		t.Fatal(err)
	}
	if pl.EnergyZone != engine.Water {
		t.Errorf("Energy Zone = %v, want the next Energy (Water)", pl.EnergyZone)
	}
	if useStadium(g, p, "") == nil {
		t.Error("a Stadium can only be used once each turn")
	}
}

func TestArcade(t *testing.T) { // B4a-072
	sawDraw, sawNone := false, false
	for seed := uint64(1); seed <= 60 && !(sawDraw && sawNone); seed++ {
		g, p := seeded(t, seed)
		stadiumIn(t, g, "B4a-072")
		hand := len(g.S.Players[p].Hand)
		if err := useStadium(g, p, ""); err != nil {
			t.Fatal(err)
		}
		heads := 0
		for _, f := range lastFlips(g) {
			if f.Heads {
				heads++
			}
		}
		got := len(g.S.Players[p].Hand)
		switch {
		case len(lastFlips(g)) != 3:
			t.Fatalf("flipped %d coins, want 3", len(lastFlips(g)))
		case heads == 3 && got == 7:
			sawDraw = true
		case heads < 3 && got == hand:
			sawNone = true
		default:
			t.Fatalf("%d heads: hand %d→%d", heads, hand, got)
		}
	}
	if !sawDraw || !sawNone {
		t.Errorf("never saw both outcomes (all heads: %v, otherwise: %v)", sawDraw, sawNone)
	}
}

func TestAreaZero(t *testing.T) { // B3a-074
	g, p := seeded(t, 1)
	stadiumIn(t, g, "B3a-074")
	pl := &g.S.Players[p]
	hand, deck := len(pl.Hand), len(pl.Deck)
	if err := useStadium(g, p, pl.Hand[0].ID); err != nil {
		t.Fatal(err)
	}
	if len(pl.Hand) != hand || len(pl.Deck) != deck {
		t.Errorf("hand %d→%d, deck %d→%d; want both unchanged (one in, one drawn)", hand, len(pl.Hand), deck, len(pl.Deck))
	}
	if g.Submit(engine.UseStadium{Player: p, CardID: "nope"}) == nil {
		t.Error("picked a card that isn't in hand")
	}
}

func TestKidsRoom(t *testing.T) { // B3b-069
	g, p := seeded(t, 1)
	stadiumIn(t, g, "B3b-069")
	pl := &g.S.Players[p]
	if useStadium(g, p, pl.Hand[0].ID) == nil {
		t.Fatal("used with no Pokémon Tool in the deck")
	}
	pl.Deck = append(pl.Deck, realCard(t, "A2-147"))
	hand := len(pl.Hand)
	if err := useStadium(g, p, pl.Hand[0].ID); err != nil {
		t.Fatal(err)
	}
	if len(pl.Hand) != hand || !handHas(pl.Hand, "Giant Cape") {
		t.Errorf("hand has %d cards, Giant Cape in it: %v; want the Cape swapped in", len(pl.Hand), handHas(pl.Hand, "Giant Cape"))
	}
}

func TestFragrantForest(t *testing.T) { // B3-153
	g, p := seeded(t, 1)
	stadiumIn(t, g, "B3-153")
	pl := &g.S.Players[p]
	pl.Deck = append(pl.Deck, realCard(t, volbeat))
	if err := useStadium(g, p, ""); err != nil {
		t.Fatal(err)
	}
	if !handHas(pl.Hand, "Volbeat") {
		t.Error("the Basic [G] Pokémon should be in hand")
	}
}

func TestMesagoza(t *testing.T) { // B2a-093
	sawHeads, sawTails := false, false
	for seed := uint64(1); seed <= 60 && !(sawHeads && sawTails); seed++ {
		g, p := seeded(t, seed)
		stadiumIn(t, g, "B2a-093")
		pl := &g.S.Players[p]
		pl.Deck = append(pl.Deck, realCard(t, volbeat))
		pl.Deck = pl.Deck[len(pl.Deck)-1:]
		hand := len(pl.Hand)
		if err := useStadium(g, p, ""); err != nil {
			t.Fatal(err)
		}
		heads := lastFlips(g)[0].Heads
		switch {
		case heads && len(pl.Hand) == hand+1:
			sawHeads = true
		case !heads && len(pl.Hand) == hand:
			sawTails = true
		default:
			t.Fatalf("heads %v: hand %d→%d", heads, hand, len(pl.Hand))
		}
	}
	if !sawHeads || !sawTails {
		t.Errorf("never saw both outcomes (heads: %v, tails: %v)", sawHeads, sawTails)
	}
}
