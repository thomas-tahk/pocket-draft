package main

import (
	"strings"
	"testing"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// One test per Supporter wording of issue #70, each on a real printing.

func mon(t *testing.T, id string, energy ...engine.Energy) *engine.InPlay {
	t.Helper()
	m := &engine.InPlay{Card: realCard(t, id), Energy: map[engine.Energy]int{}}
	for _, e := range energy {
		m.Energy[e]++
	}
	return m
}

// supporterGame is a game past setup where the player on turn holds the
// Supporter id, has active as their Active and bench in slots from the left.
func supporterGame(t *testing.T, id string, active *engine.InPlay, bench ...*engine.InPlay) (*engine.Game, int) {
	t.Helper()
	g, p := trainerGame(t, realCard(t, id))
	pl := &g.S.Players[p]
	pl.Active = active
	for i, b := range bench {
		pl.Bench[i] = b
	}
	return g, p
}

func playOn(t *testing.T, g *engine.Game, p int, id string, at engine.Spot) {
	t.Helper()
	if err := g.Submit(engine.PlayTrainer{Player: p, CardID: id, Target: at}); err != nil {
		t.Fatalf("play %s on %+v: %v", id, at, err)
	}
}

func refused(g *engine.Game, p int, id string, at engine.Spot) bool {
	return g.Submit(engine.PlayTrainer{Player: p, CardID: id, Target: at}) != nil
}

func inHand(pl *engine.Player, name string) bool {
	for _, c := range pl.Hand {
		if c.Name == name {
			return true
		}
	}
	return false
}

func attack(t *testing.T, g *engine.Game, p int) {
	t.Helper()
	if err := g.Submit(engine.UseAttack{Player: p, Index: 0}); err != nil {
		t.Fatalf("attack: %v", err)
	}
}

// Ilima (A3-149): a damaged [C] Benched Pokémon goes to hand; an undamaged one can't be chosen.
func TestIlima(t *testing.T) {
	useRealCards(t)
	hurt := mon(t, "B3b-055", engine.Colorless)
	hurt.Damage = 30
	g, p := supporterGame(t, "A3-149", mon(t, "B3b-055"), hurt)
	pl := &g.S.Players[p]

	if !refused(g, p, "A3-149", engine.Spot{Player: p, Pos: 0}) {
		t.Error("undamaged Active was a target")
	}
	playOn(t, g, p, "A3-149", engine.Spot{Player: p, Pos: 1})

	if pl.Bench[0] != nil || !inHand(pl, "Snorlax") {
		t.Error("damaged Snorlax did not go to hand")
	}
}

// Iono (B2a-089): both players reshuffle and draw as many as they held.
func TestIono(t *testing.T) {
	useRealCards(t)
	g, p := supporterGame(t, "B2a-089", mon(t, "A4b-069"))
	me, opp := &g.S.Players[p], &g.S.Players[1-p]
	mine, theirs := len(me.Hand)-1, len(opp.Hand)

	playOn(t, g, p, "B2a-089", engine.NoSpot)

	if len(me.Hand) != mine || len(opp.Hand) != theirs {
		t.Errorf("hands %d/%d, want %d/%d", len(me.Hand), len(opp.Hand), mine, theirs)
	}
}

// Irida (A4b-330): heals 40 only from Pokémon with [W] Energy.
func TestIrida(t *testing.T) {
	useRealCards(t)
	wet, dry := mon(t, "B1a-017", engine.Water), mon(t, "B3b-055")
	wet.Damage, dry.Damage = 50, 50
	g, p := supporterGame(t, "A4b-330", wet, dry)

	playOn(t, g, p, "A4b-330", engine.NoSpot)

	if wet.Damage != 10 || dry.Damage != 50 {
		t.Errorf("damage %d/%d, want 10/50", wet.Damage, dry.Damage)
	}
}

// Iris (B2b-067): a KO by Haxorus's attack this turn scores 1 more.
func TestIris(t *testing.T) {
	useRealCards(t)
	g, p := supporterGame(t, "B2b-067", mon(t, "B2b-056", engine.Fighting, engine.Metal, engine.Colorless))
	opp := g.S.Players[1-p].Active
	opp.Damage = opp.Card.HP - 10

	playOn(t, g, p, "B2b-067", engine.NoSpot)
	attack(t, g, p)

	if got := g.S.Players[p].Points; got != 2 {
		t.Errorf("points = %d, want 2 (1 for the KO, 1 from Iris)", got)
	}
}

// Jasmine (A4-160): during the opponent's next turn Steelix takes 50 less.
func TestJasmine(t *testing.T) {
	useRealCards(t)
	steelix := mon(t, "B3-219")
	g, p := supporterGame(t, "A4-160", steelix)
	playOn(t, g, p, "A4-160", engine.NoSpot)
	mustEnd(t, g, p)
	g.S.Players[1-p].Active.Energy[engine.Fire] = 2 // Torkoal: Flamethrower [RR] 70

	attack(t, g, 1-p)

	if steelix.Damage != 70+20-50 {
		t.Errorf("Steelix took %d, want 40 (70 +20 Weakness −50)", steelix.Damage)
	}
}

func mustEnd(t *testing.T, g *engine.Game, p int) {
	t.Helper()
	if err := g.Submit(engine.EndTurn{Player: p}); err != nil {
		t.Fatalf("end turn: %v", err)
	}
}

// Juggler (B2-151): needs 3 Energy types in play; moves all Bench Energy to the Active.
func TestJuggler(t *testing.T) {
	useRealCards(t)
	bench := mon(t, "B1a-017", engine.Water, engine.Grass)
	g, p := supporterGame(t, "B2-151", mon(t, "A4b-069", engine.Fire), bench)
	pl := &g.S.Players[p]

	playOn(t, g, p, "B2-151", engine.NoSpot)

	if pl.Active.Energy[engine.Water] != 1 || pl.Active.Energy[engine.Grass] != 1 || len(bench.Energy) != 0 {
		t.Errorf("Active %v, Bench %v", pl.Active.Energy, bench.Energy)
	}
}

func TestJugglerNeedsThreeTypes(t *testing.T) {
	useRealCards(t)
	g, p := supporterGame(t, "B2-151", mon(t, "A4b-069", engine.Fire), mon(t, "B1a-017", engine.Water))
	if !refused(g, p, "B2-151", engine.NoSpot) {
		t.Error("Juggler played with 2 Energy types")
	}
}

// Kiawe (A3-150): 2 [R] on Alolan Marowak, then the turn ends.
func TestKiawe(t *testing.T) {
	useRealCards(t)
	marowak := mon(t, "B2-018")
	g, p := supporterGame(t, "A3-150", marowak)

	playOn(t, g, p, "A3-150", engine.Spot{Player: p, Pos: 0})

	if marowak.Energy[engine.Fire] != 2 || g.S.Active == p {
		t.Errorf("Fire %d, turn passed %v", marowak.Energy[engine.Fire], g.S.Active != p)
	}
}

// Koga (A1-222): Active Muk to hand; the Bench Pokémon becomes Active.
func TestKoga(t *testing.T) {
	useRealCards(t)
	g, p := supporterGame(t, "A1-222", mon(t, "B4-092"), mon(t, "B1a-017"))
	pl := &g.S.Players[p]

	playOn(t, g, p, "A1-222", engine.Spot{Player: p, Pos: 0})

	if !inHand(pl, "Muk") || pl.Active == nil || pl.Active.Card.Name != "Squirtle" {
		t.Errorf("Muk in hand %v, Active %+v", inHand(pl, "Muk"), pl.Active)
	}
}

func TestKogaNeedsABench(t *testing.T) {
	useRealCards(t)
	g, p := supporterGame(t, "A1-222", mon(t, "B4-092"))
	if !refused(g, p, "A1-222", engine.Spot{Player: p, Pos: 0}) {
		t.Error("Koga left the player with no Pokémon")
	}
}

// Korrina (B3-149): [F] attacks do +30 to an Active ex this turn.
func TestKorrina(t *testing.T) {
	useRealCards(t)
	g, p := supporterGame(t, "B3-149", mon(t, "B2-079", engine.Fighting))
	g.S.Players[1-p].Active = mon(t, "A4b-158") // Mewtwo ex

	playOn(t, g, p, "B3-149", engine.NoSpot)
	attack(t, g, p)

	if d := g.S.Players[1-p].Active.Damage; d != 10+30 {
		t.Errorf("Mewtwo ex took %d, want 40", d)
	}
}

// Lana (A3-152): with Araquanid in play, an opponent's Benched Pokémon is switched in.
func TestLana(t *testing.T) {
	useRealCards(t)
	g, p := supporterGame(t, "A3-152", mon(t, "B4-047"))
	opp := &g.S.Players[1-p]
	opp.Bench[2] = mon(t, "B1a-017")

	playOn(t, g, p, "A3-152", engine.Spot{Player: 1 - p, Pos: 3})

	if opp.Active.Card.Name != "Squirtle" || opp.Bench[2].Card.Name != "Torkoal" {
		t.Errorf("opponent Active %s", opp.Active.Card.Name)
	}
}

func TestLanaNeedsAraquanid(t *testing.T) {
	useRealCards(t)
	g, p := supporterGame(t, "A3-152", mon(t, "B1a-017"))
	g.S.Players[1-p].Bench[0] = mon(t, "B1a-017")
	if !refused(g, p, "A3-152", engine.Spot{Player: 1 - p, Pos: 1}) {
		t.Error("Lana played without Araquanid")
	}
}

// Leaf (A4b-346): this turn the Active's Retreat Cost is 2 less; next turn it isn't.
func TestLeaf(t *testing.T) {
	useRealCards(t)
	g, p := supporterGame(t, "A4b-346", mon(t, "B3b-055")) // Snorlax, Retreat 3

	playOn(t, g, p, "A4b-346", engine.NoSpot)

	if c := g.RetreatCost(p); c != 1 {
		t.Errorf("retreat cost this turn = %d, want 1", c)
	}
	mustEnd(t, g, p)
	mustEnd(t, g, 1-p)
	if c := g.RetreatCost(p); c != 3 {
		t.Errorf("retreat cost next turn = %d, want 3", c)
	}
}

// Lillie (A4b-348): heals 60 from a Stage 2.
func TestLillie(t *testing.T) {
	useRealCards(t)
	venusaur := mon(t, "B1a-003")
	venusaur.Damage = 80
	g, p := supporterGame(t, "A4b-348", venusaur)

	playOn(t, g, p, "A4b-348", engine.Spot{Player: p, Pos: 0})

	if venusaur.Damage != 20 {
		t.Errorf("damage %d, want 20", venusaur.Damage)
	}
}

// Looker (A3a-068): the opponent's deck Supporters are named in the log.
func TestLooker(t *testing.T) {
	useRealCards(t)
	g, p := supporterGame(t, "A3a-068", mon(t, "A4b-069"))
	opp := &g.S.Players[1-p]
	opp.Deck = append(opp.Deck, realCard(t, "A4b-373"))

	playOn(t, g, p, "A3a-068", engine.NoSpot)

	if log := strings.Join(g.Narration, "\n"); !strings.Contains(log, "revealed from their deck: Professor's Research") {
		t.Errorf("log does not name the Supporter:\n%s", log)
	}
}

// Lt. Surge (A1-226): [L] from the Bench to the Active Raichu; other Energy stays.
func TestLtSurge(t *testing.T) {
	useRealCards(t)
	pikachu := mon(t, "B4-049", engine.Lightning, engine.Lightning, engine.Water)
	raichu := mon(t, "B4-050")
	g, p := supporterGame(t, "A1-226", raichu, pikachu)

	playOn(t, g, p, "A1-226", engine.NoSpot)

	if raichu.Energy[engine.Lightning] != 2 || pikachu.Energy[engine.Water] != 1 || pikachu.Energy[engine.Lightning] != 0 {
		t.Errorf("Raichu %v, Pikachu %v", raichu.Energy, pikachu.Energy)
	}
}

// Lusamine (A4b-350): once the opponent has a point, 2 random discarded Energy on an Ultra Beast.
func TestLusamine(t *testing.T) {
	useRealCards(t)
	buzzwole := mon(t, "B2-014")
	g, p := supporterGame(t, "A4b-350", buzzwole)
	pl := &g.S.Players[p]
	pl.DiscardedEnergy = []engine.Energy{engine.Fire, engine.Water, engine.Grass}
	if !refused(g, p, "A4b-350", engine.Spot{Player: p, Pos: 0}) {
		t.Fatal("Lusamine played before the opponent scored")
	}
	g.S.Players[1-p].Points = 1

	playOn(t, g, p, "A4b-350", engine.Spot{Player: p, Pos: 0})

	n := 0
	for _, k := range buzzwole.Energy {
		n += k
	}
	if n != 2 || len(pl.DiscardedEnergy) != 1 {
		t.Errorf("attached %d, %d left in discard", n, len(pl.DiscardedEnergy))
	}
}

// Lyra (A4b-332): a damaged Active switches with a Benched Pokémon and is cured.
func TestLyra(t *testing.T) {
	useRealCards(t)
	hurt := mon(t, "B3b-055")
	hurt.Damage = 20
	hurt.Inflict(engine.Poison)
	g, p := supporterGame(t, "A4b-332", hurt, mon(t, "B1a-017"))
	pl := &g.S.Players[p]

	playOn(t, g, p, "A4b-332", engine.Spot{Player: p, Pos: 1})

	if pl.Active.Card.Name != "Squirtle" || pl.Bench[0] != hurt || hurt.Has(engine.Poison) {
		t.Errorf("Active %s, Snorlax still poisoned %v", pl.Active.Card.Name, hurt.Has(engine.Poison))
	}
}

func TestLyraNeedsDamage(t *testing.T) {
	useRealCards(t)
	g, p := supporterGame(t, "A4b-332", mon(t, "B3b-055"), mon(t, "B1a-017"))
	if !refused(g, p, "A4b-332", engine.Spot{Player: p, Pos: 1}) {
		t.Error("Lyra switched an undamaged Active")
	}
}

// Mallow (A3-154): all damage healed from Tsareena, and its Energy discarded.
func TestMallow(t *testing.T) {
	useRealCards(t)
	tsareena := mon(t, "A3b-005", engine.Grass, engine.Grass)
	tsareena.Damage = 100
	g, p := supporterGame(t, "A3-154", tsareena)

	playOn(t, g, p, "A3-154", engine.Spot{Player: p, Pos: 0})

	if tsareena.Damage != 0 || tsareena.Energy[engine.Grass] != 0 {
		t.Errorf("damage %d, energy %v", tsareena.Damage, tsareena.Energy)
	}
}

// Marlon (B1-221): heals 70 from Carracosta.
func TestMarlon(t *testing.T) {
	useRealCards(t)
	carracosta := mon(t, "B1-067")
	carracosta.Damage = 100
	g, p := supporterGame(t, "B1-221", carracosta)

	playOn(t, g, p, "B1-221", engine.Spot{Player: p, Pos: 0})

	if carracosta.Damage != 30 {
		t.Errorf("damage %d, want 30", carracosta.Damage)
	}
}

// Mars (A4b-344): an opponent on 1 point redraws to 2 cards.
func TestMars(t *testing.T) {
	useRealCards(t)
	g, p := supporterGame(t, "A4b-344", mon(t, "A4b-069"))
	opp := &g.S.Players[1-p]
	opp.Points = 1

	playOn(t, g, p, "A4b-344", engine.NoSpot)

	if len(opp.Hand) != 2 {
		t.Errorf("opponent hand = %d, want 2", len(opp.Hand))
	}
}

// Misty (A1-220): one [W] Energy per heads before the first tails.
func TestMisty(t *testing.T) {
	useRealCards(t)
	squirtle := mon(t, "B1a-017")
	g, p := supporterGame(t, "A1-220", squirtle)

	playOn(t, g, p, "A1-220", engine.Spot{Player: p, Pos: 0})

	flips := g.Actions[len(g.Actions)-1].Flips
	if len(flips) == 0 || flips[len(flips)-1].Heads {
		t.Fatalf("flips %+v don't end on tails", flips)
	}
	if squirtle.Energy[engine.Water] != len(flips)-1 {
		t.Errorf("Water %d for %d heads", squirtle.Energy[engine.Water], len(flips)-1)
	}
}

// Nemona (B2a-090): Pawmot does +80 to an Active ex this turn.
func TestNemona(t *testing.T) {
	useRealCards(t)
	g, p := supporterGame(t, "B2a-090", mon(t, "B3a-016", engine.Lightning))
	g.S.Players[1-p].Active = mon(t, "A4b-158") // Mewtwo ex

	playOn(t, g, p, "B2a-090", engine.NoSpot)
	attack(t, g, p)

	if d := g.S.Players[1-p].Active.Damage; d != 60+80 {
		t.Errorf("Mewtwo ex took %d, want 140", d)
	}
}

// Parasol Lady (B3-152): a [W] non-ex goes to hand; a [W] ex can't be chosen.
func TestParasolLady(t *testing.T) {
	useRealCards(t)
	g, p := supporterGame(t, "B3-152", mon(t, "P-A-014"), mon(t, "B1a-017"))
	pl := &g.S.Players[p]
	if !refused(g, p, "B3-152", engine.Spot{Player: p, Pos: 0}) {
		t.Error("Lapras ex was a target")
	}

	playOn(t, g, p, "B3-152", engine.Spot{Player: p, Pos: 1})

	if pl.Bench[0] != nil || !inHand(pl, "Squirtle") {
		t.Error("Squirtle did not go to hand")
	}
}

// Penny (B2a-092): copies the opponent's only deck Supporter, Professor's Research.
func TestPenny(t *testing.T) {
	useRealCards(t)
	g, p := supporterGame(t, "B2a-092", mon(t, "A4b-069"))
	pl, opp := &g.S.Players[p], &g.S.Players[1-p]
	opp.Deck = append(opp.Deck, realCard(t, "A4b-373"))
	hand, oppDeck := len(pl.Hand), len(opp.Deck)

	playOn(t, g, p, "B2a-092", engine.NoSpot)

	if len(pl.Hand) != hand-1+2 || len(opp.Deck) != oppDeck {
		t.Errorf("hand %d→%d (want +1), opponent deck %d→%d", hand, len(pl.Hand), oppDeck, len(opp.Deck))
	}
}

// Juliana (B3a-071): a random Stage 2 from the deck into hand.
func TestJuliana(t *testing.T) {
	useRealCards(t)
	g, p := supporterGame(t, "B3a-071", mon(t, "A4b-069"))
	pl := &g.S.Players[p]
	pl.Deck = append(pl.Deck, realCard(t, "B1a-003"))
	hand := len(pl.Hand)

	playOn(t, g, p, "B3a-071", engine.NoSpot)

	if len(pl.Hand) != hand || !inHand(pl, "Venusaur") {
		t.Errorf("hand %d→%d, Venusaur in hand %v", hand, len(pl.Hand), inHand(pl, "Venusaur"))
	}
}

// Lisia (B1-226): 2 random Basics with 50 HP or less; bigger Basics stay.
func TestLisia(t *testing.T) {
	useRealCards(t)
	g, p := supporterGame(t, "B1-226", mon(t, "A4b-069"))
	pl := &g.S.Players[p]
	small := realCard(t, "B4a-063") // Happiny, 30 HP
	pl.Deck = append(pl.Deck, small, small, small)
	hand, deck := len(pl.Hand), len(pl.Deck)

	playOn(t, g, p, "B1-226", engine.NoSpot)

	if len(pl.Hand) != hand-1+2 || len(pl.Deck) != deck-2 {
		t.Errorf("hand %d→%d, deck %d→%d; want 2 searched (Torkoal is too big)", hand, len(pl.Hand), deck, len(pl.Deck))
	}
}

// Every printing of the built wordings plays its text.
func TestBatch3_10EveryPrintingMapped(t *testing.T) {
	useRealCards(t)
	ids := []string{"A3-149", "A3-191", "B2a-089", "B2a-106", "A4b-340", "A4b-341", "A2b-069", "A2b-088",
		"A4b-330", "A4b-331", "A2a-072", "A2a-087", "B2b-067", "B2b-081", "A4-160", "A4-200", "B2-151", "B2-192",
		"A3-150", "A3-192", "B3a-071", "B3a-086", "B1-226", "B1-271", "A1-222", "A1-269", "B3-149", "B3-190", "A3-152", "A3-194",
		"A4b-346", "A4b-347", "A1a-068", "A1a-082", "A4b-348", "A4b-349", "A4b-374", "A3-155", "A3-197", "A3-209",
		"A3a-068", "A3a-082", "A1-226", "A1-273", "A4b-350", "A4b-351", "A4b-375", "A3a-069", "A3a-083",
		"A4b-332", "A4b-333", "A4-157", "A4-197", "A3-154", "A3-196", "B1-221", "B1-266",
		"A4b-344", "A4b-345", "A2-155", "A2-195", "A1-220", "A1-267", "B2a-090", "B2a-107",
		"B3-152", "B3-193", "B2a-092", "B2a-109", "A3b-069", "A3b-086"}
	for _, id := range ids {
		if realCard(t, id).Trainer == nil {
			t.Errorf("%s (%s) has no effect", id, rawByID[id].Name)
		}
	}
}
