package main

import (
	"testing"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// One test per wording of issue #71, each on a real card.

func inPlay(t *testing.T, id string, energy map[engine.Energy]int) *engine.InPlay {
	t.Helper()
	if energy == nil {
		energy = map[engine.Energy]int{}
	}
	return &engine.InPlay{Card: realCard(t, id), Energy: energy}
}

// game311 is a game past setup where the player on turn holds the Trainer.
func game311(t *testing.T, trainerID string) (*engine.Game, int, *engine.Player, *engine.Player) {
	t.Helper()
	useRealCards(t)
	g, p := trainerGame(t, realCard(t, trainerID))
	return g, p, &g.S.Players[p], &g.S.Players[1-p]
}

func play311(t *testing.T, g *engine.Game, p int, id string, at engine.Spot) {
	t.Helper()
	if err := g.Submit(engine.PlayTrainer{Player: p, CardID: id, Target: at}); err != nil {
		t.Fatalf("play %s: %v", id, err)
	}
}

func refused311(t *testing.T, g *engine.Game, p int, id string, at engine.Spot) {
	t.Helper()
	if g.Submit(engine.PlayTrainer{Player: p, CardID: id, Target: at}) == nil {
		t.Fatalf("%s was played; want refused", id)
	}
}

func energyOn(m *engine.InPlay) int {
	n := 0
	for _, v := range m.Energy {
		n += v
	}
	return n
}

func lastFlips311(g *engine.Game) []engine.CoinFlip { return g.Actions[len(g.Actions)-1].Flips }

func headsBeforeTails(flips []engine.CoinFlip) int {
	n := 0
	for _, f := range flips {
		if !f.Heads {
			break
		}
		n++
	}
	return n
}

// Every wording here is printed on a real card.
func TestTexts311AreRealWordings(t *testing.T) {
	useRealCards(t)
	seen := map[string]bool{}
	for _, rc := range rawByID {
		seen[rc.TrainerText] = true
	}
	for text := range texts311 {
		if !seen[text] {
			t.Errorf("no card prints %q", text)
		}
	}
}

func TestPiers(t *testing.T) {
	g, p, me, opp := game311(t, "B2-152")
	opp.Active.Energy[engine.Fire] = 3
	refused311(t, g, p, "B2-152", engine.NoSpot) // no Galarian Obstagoon

	me.Bench[0] = inPlay(t, "B2-100", nil)
	play311(t, g, p, "B2-152", engine.NoSpot)

	if n := energyOn(opp.Active); n != 1 {
		t.Errorf("opponent's Active has %d Energy, want 1", n)
	}
}

func TestPokemonCenterLady(t *testing.T) {
	g, p, me, _ := game311(t, "A2b-070")
	me.Active.Damage = 50
	me.Active.Inflict(engine.Poison)
	me.Active.Inflict(engine.Asleep)

	play311(t, g, p, "A2b-070", engine.Spot{Player: p, Pos: 0})

	if me.Active.Damage != 20 || len(me.Active.Conditions()) != 0 {
		t.Errorf("damage %d, conditions %v; want 20 and none", me.Active.Damage, me.Active.Conditions())
	}
}

func TestProfessorSada(t *testing.T) {
	g, p, me, _ := game311(t, "B3a-072")
	me.Bench[0] = inPlay(t, "B3a-034", nil) // Great Tusk, Ancient
	me.DiscardedEnergy = []engine.Energy{engine.Fire, engine.Fire, engine.Water, engine.Grass, engine.Metal}
	if got := len(g.LegalMoves(p)); got == 0 {
		t.Fatal("no legal moves")
	}
	refused311(t, g, p, "B3a-072", engine.Spot{Player: p, Pos: 0}) // Torkoal isn't Ancient

	play311(t, g, p, "B3a-072", engine.Spot{Player: p, Pos: 1})

	tusk := me.Bench[0]
	if tusk.Energy[engine.Fire] != 1 || tusk.Energy[engine.Water] != 1 || tusk.Energy[engine.Grass] != 1 {
		t.Errorf("Great Tusk energy %v, want one each of 3 types", tusk.Energy)
	}
	if len(me.DiscardedEnergy) != 2 {
		t.Errorf("discard energy %v, want 2 left", me.DiscardedEnergy)
	}
}

func TestProfessorTuro(t *testing.T) {
	g, p, me, _ := game311(t, "B3a-073")
	me.Bench[1] = inPlay(t, "B3a-027", map[engine.Energy]int{engine.Psychic: 2}) // Iron Valiant, Future
	deck := len(me.Deck)

	play311(t, g, p, "B3a-073", engine.Spot{Player: p, Pos: 2})

	if me.Bench[1] != nil || len(me.Deck) != deck+1 || len(me.DiscardedEnergy) != 2 {
		t.Errorf("bench %v, deck %d→%d, discarded energy %v", me.Bench[1], deck, len(me.Deck), me.DiscardedEnergy)
	}
}

// Turo on the Active: you choose the new one.
func TestProfessorTuroActive(t *testing.T) {
	g, p, me, _ := game311(t, "B3a-073")
	me.Active = inPlay(t, "B3a-027", nil)
	refused311(t, g, p, "B3a-073", engine.Spot{Player: p, Pos: 0}) // nothing to replace it

	me.Bench[2] = inPlay(t, "A4b-069", nil)
	play311(t, g, p, "B3a-073", engine.Spot{Player: p, Pos: 0})

	if me.Active == nil || me.Active.Card.Name != "Torkoal" || me.Bench[2] != nil {
		t.Errorf("Active %v, want the benched Torkoal promoted", me.Active)
	}
}

func TestPsychicCard(t *testing.T) {
	g, p, me, opp := game311(t, "B4-150")
	opp.Bench[0] = inPlay(t, "A4b-069", map[engine.Energy]int{engine.Fire: 1})
	refused311(t, g, p, "B4-150", engine.Spot{Player: 1 - p, Pos: 1}) // Torkoal has no Psychic attack

	me.Active = inPlay(t, "B4-070", nil) // Mewtwo
	play311(t, g, p, "B4-150", engine.Spot{Player: 1 - p, Pos: 1})

	if energyOn(opp.Bench[0]) != 0 || opp.Active.Energy[engine.Fire] != 1 {
		t.Errorf("bench %v, active %v; want the Fire moved", opp.Bench[0].Energy, opp.Active.Energy)
	}
}

// topFour puts cards on top of p's deck.
func topFour(t *testing.T, pl *engine.Player, ids ...string) {
	t.Helper()
	var top []engine.Card
	for _, id := range ids {
		top = append(top, realCard(t, id))
	}
	pl.Deck = append(top, pl.Deck...)
}

func handHas311(pl *engine.Player, name string) int {
	n := 0
	for _, c := range pl.Hand {
		if c.Name == name {
			n++
		}
	}
	return n
}

func TestPuppyLovingGirl(t *testing.T) {
	g, p, me, _ := game311(t, "B3b-067")
	topFour(t, me, "B3b-010", "A2-057", "B3b-010", "A4b-069") // Growlithe, Electivire, Growlithe, Torkoal
	deck := len(me.Deck)

	play311(t, g, p, "B3b-067", engine.NoSpot)

	if handHas311(me, "Growlithe") != 2 || handHas311(me, "Electivire") != 0 || len(me.Deck) != deck-2 {
		t.Errorf("hand %v, deck %d→%d", me.Hand, deck, len(me.Deck))
	}
}

func TestSightseer(t *testing.T) {
	g, p, me, _ := game311(t, "B2-150")
	topFour(t, me, "A2-057", "B3b-010", "A2-060", "A2-057") // Electivire (Stage 1) ×2, Growlithe, Luxray
	play311(t, g, p, "B2-150", engine.NoSpot)

	if handHas311(me, "Electivire") != 2 || handHas311(me, "Luxray") != 0 || handHas311(me, "Growlithe") != 0 {
		t.Errorf("hand %v, want both Electivire only", me.Hand)
	}
}

func TestTravelingMerchant(t *testing.T) {
	g, p, me, _ := game311(t, "A4a-070")
	topFour(t, me, "A4b-322", "A2-057", "A2-149", "A4b-373") // Rocky Helmet, Electivire, Lum Berry, Professor's Research
	play311(t, g, p, "A4a-070", engine.NoSpot)

	if handHas311(me, "Rocky Helmet") != 1 || handHas311(me, "Lum Berry") != 1 || handHas311(me, "Electivire") != 0 {
		t.Errorf("hand %v, want the two Tools", me.Hand)
	}
}

// attack has p's Active use its first attack and returns the damage dealt.
func attack311(t *testing.T, g *engine.Game, p int) int {
	t.Helper()
	def := g.S.Players[1-p].Active
	before := def.Damage
	if err := g.Submit(engine.UseAttack{Player: p, Index: 0}); err != nil {
		t.Fatalf("attack: %v", err)
	}
	return def.Damage - before
}

func TestRed(t *testing.T) {
	g, p, me, opp := game311(t, "A2b-071")
	me.Active.Energy[engine.Fire] = 2      // Torkoal, Flamethrower 70
	opp.Active = inPlay(t, "B2a-126", nil) // Entei ex
	play311(t, g, p, "A2b-071", engine.NoSpot)

	if dmg := attack311(t, g, p); dmg != 90 {
		t.Errorf("damage to an ex = %d, want 70+20", dmg)
	}
}

func TestRedOnlyAgainstEx(t *testing.T) {
	g, p, me, _ := game311(t, "A2b-071")
	me.Active.Energy[engine.Fire] = 2
	play311(t, g, p, "A2b-071", engine.NoSpot)

	if dmg := attack311(t, g, p); dmg != 70 {
		t.Errorf("damage to a non-ex = %d, want 70", dmg)
	}
}

func TestSophocles(t *testing.T) {
	g, p, me, _ := game311(t, "A3-153")
	me.Active = inPlay(t, "B4a-045", map[engine.Energy]int{engine.Metal: 1}) // Togedemaru, 20+
	play311(t, g, p, "A3-153", engine.NoSpot)

	if dmg := attack311(t, g, p); dmg != 50 && dmg != 80 {
		t.Errorf("Togedemaru did %d, want 20+30 (or +30 heads)", dmg)
	}
}

func TestSabrina(t *testing.T) {
	g, p, _, opp := game311(t, "A1-225")
	refused311(t, g, p, "A1-225", engine.NoSpot) // empty Bench
	out := opp.Active
	out.Inflict(engine.Poison)
	opp.Bench[0] = inPlay(t, "B2a-126", nil)
	opp.Bench[2] = inPlay(t, "B4-070", nil)

	play311(t, g, p, "A1-225", engine.NoSpot)
	if pr := g.S.Pending; pr == nil || pr.Player != 1-p || pr.Kind != engine.PromptNewActive {
		t.Fatalf("pending %+v, want the opponent choosing a new Active", pr)
	}
	if err := g.Submit(engine.ChooseNewActive{Player: 1 - p, BenchIndex: 2}); err != nil {
		t.Fatal(err)
	}

	if opp.Active.Card.Name != "Mewtwo" || opp.Bench[2] != out || out.Has(engine.Poison) {
		t.Errorf("active %s, bench slot 3 %v; want Mewtwo in, old Active benched and cured", opp.Active.Card.Name, opp.Bench[2])
	}
	if g.S.Active != p || g.S.Pending != nil {
		t.Error("the turn didn't go back to the Sabrina player")
	}
}

func TestSilver(t *testing.T) {
	g, p, _, opp := game311(t, "A4-158")
	opp.Hand = []engine.Card{realCard(t, "A4b-069"), realCard(t, "A1-225")}
	deck := len(opp.Deck)

	play311(t, g, p, "A4-158", engine.NoSpot)

	if len(opp.Hand) != 1 || opp.Hand[0].Name != "Torkoal" || len(opp.Deck) != deck+1 {
		t.Errorf("opponent hand %v, deck %d→%d; want Sabrina shuffled in", opp.Hand, deck, len(opp.Deck))
	}
}

func TestSkyla(t *testing.T) {
	g, p, me, _ := game311(t, "B4-152")
	me.Bench[1] = inPlay(t, "B4-070", nil)
	refused311(t, g, p, "B4-152", engine.Spot{Player: p, Pos: 2}) // Torkoal is Basic

	me.Active = inPlay(t, "A2-057", nil) // Electivire, Stage 1
	play311(t, g, p, "B4-152", engine.Spot{Player: p, Pos: 2})

	if me.Active.Card.Name != "Mewtwo" || me.Bench[1].Card.Name != "Electivire" {
		t.Errorf("active %s, bench %s", me.Active.Card.Name, me.Bench[1].Card.Name)
	}
}

func TestTeamRocketGrunt(t *testing.T) {
	g, p, _, opp := game311(t, "A2b-072")
	refused311(t, g, p, "A2b-072", engine.NoSpot) // no Energy to discard
	opp.Active.Energy[engine.Fire] = 5

	play311(t, g, p, "A2b-072", engine.NoSpot)

	want := 5 - min(5, headsBeforeTails(lastFlips311(g)))
	if n := energyOn(opp.Active); n != want {
		t.Errorf("energy left %d, want %d (flips %v)", n, want, lastFlips311(g))
	}
}

func TestTeamRocketsBoss(t *testing.T) {
	g, p, _, opp := game311(t, "B4a-071")
	opp.Bench[1] = inPlay(t, "B4-070", nil)
	opp.Hand = []engine.Card{realCard(t, "A4b-069"), realCard(t, "A1-225"), realCard(t, "A2-057"), realCard(t, "B2a-126"), realCard(t, "B4-070")}

	play311(t, g, p, "B4a-071", engine.NoSpot)

	if opp.BenchCount() != 3 || len(opp.Hand) != 3 || opp.Bench[0].Card.Name != "Torkoal" || opp.Bench[2].Card.Name != "Entei ex" {
		t.Errorf("bench %d, hand %v", opp.BenchCount(), opp.Hand)
	}
}

func TestTeamRocketsMasterPlan(t *testing.T) {
	g, p, me, opp := game311(t, "B4a-070")
	play311(t, g, p, "B4a-070", engine.NoSpot)

	if !opp.Active.Has(engine.Confused) {
		t.Error("opponent's Active isn't Confused")
	}
	tails := !lastFlips311(g)[0].Heads
	if me.Active.Has(engine.Confused) != tails {
		t.Errorf("own Active Confused = %v, flip tails = %v", me.Active.Has(engine.Confused), tails)
	}
}

func TestTeamStarGrunt(t *testing.T) {
	g, p, _, opp := game311(t, "B2a-088")
	opp.Active.Energy[engine.Fire] = 2 // Torkoal: no Ability
	refused311(t, g, p, "B2a-088", engine.NoSpot)

	opp.Bench[0] = inPlay(t, "B4a-062", map[engine.Energy]int{engine.Darkness: 2}) // Team Rocket's Kecleon
	play311(t, g, p, "B2a-088", engine.NoSpot)

	if energyOn(opp.Bench[0]) != 1 || energyOn(opp.Active) != 2 {
		t.Errorf("Kecleon %v, Torkoal %v", opp.Bench[0].Energy, opp.Active.Energy)
	}
}

func TestVolkner(t *testing.T) {
	g, p, me, _ := game311(t, "A2-153")
	me.Bench[0] = inPlay(t, "A2-057", nil)                        // Electivire
	refused311(t, g, p, "A2-153", engine.Spot{Player: p, Pos: 1}) // no [L] in discard
	me.DiscardedEnergy = []engine.Energy{engine.Lightning, engine.Fire, engine.Lightning, engine.Lightning}

	play311(t, g, p, "A2-153", engine.Spot{Player: p, Pos: 1})

	if me.Bench[0].Energy[engine.Lightning] != 2 || len(me.DiscardedEnergy) != 2 {
		t.Errorf("Electivire %v, discard %v", me.Bench[0].Energy, me.DiscardedEnergy)
	}
}

func TestWallace(t *testing.T) {
	g, p, me, _ := game311(t, "B3b-068")
	me.Bench[0] = inPlay(t, "B4a-097", nil) // Goldeen 50 HP
	me.Deck = append(me.Deck, realCard(t, "B4a-098"))
	refused311(t, g, p, "B3b-068", engine.Spot{Player: p, Pos: 1}) // first turn

	g.S.Turn = 5
	play311(t, g, p, "B3b-068", engine.Spot{Player: p, Pos: 1})

	if me.Bench[0].Card.Name != "Seaking" {
		t.Errorf("benched %s, want Seaking", me.Bench[0].Card.Name)
	}
}

func TestWally(t *testing.T) {
	g, p, me, _ := game311(t, "B4-153")
	refused311(t, g, p, "B4-153", engine.Spot{Player: p, Pos: 0}) // no Stage 2
	me.Bench[2] = inPlay(t, "A2-060", nil)                        // Luxray

	play311(t, g, p, "B4-153", engine.Spot{Player: p, Pos: 3})

	if me.Bench[2].Energy[engine.Colorless] != 1 {
		t.Errorf("Luxray %v, want 1 Colorless", me.Bench[2].Energy)
	}
}

func TestWhitney(t *testing.T) {
	g, p, me, _ := game311(t, "A4a-069")
	me.Active.Damage = 30
	refused311(t, g, p, "A4a-069", engine.Spot{Player: p, Pos: 0}) // Torkoal isn't Miltank

	me.Active = inPlay(t, "A4a-062", nil)
	me.Active.Damage = 80
	me.Active.Inflict(engine.Poison)
	me.Active.Inflict(engine.Paralyzed)
	play311(t, g, p, "A4a-069", engine.Spot{Player: p, Pos: 0})

	if me.Active.Damage != 20 || me.Active.Has(engine.Paralyzed) || !me.Active.Has(engine.Poison) {
		t.Errorf("damage %d, conditions %v; want 20 and Poison only", me.Active.Damage, me.Active.Conditions())
	}
}

func TestWill(t *testing.T) {
	g, p, me, _ := game311(t, "A4-156")
	me.Active = inPlay(t, "B3a-038", map[engine.Energy]int{engine.Darkness: 1}) // Sneasel, Quick Attack 10+20 on heads
	play311(t, g, p, "A4-156", engine.NoSpot)

	if dmg := attack311(t, g, p); dmg != 30 {
		t.Errorf("Quick Attack after Will did %d, want 30", dmg)
	}
	if f := lastFlips311(g); len(f) != 1 || !f[0].Heads {
		t.Errorf("flips %v, want one heads", f)
	}
}

// searchFinds plays a search Trainer over a deck of Torkoal plus one match.
func searchFinds(t *testing.T, trainerID, matchID string) {
	t.Helper()
	g, p, me, _ := game311(t, trainerID)
	match := realCard(t, matchID)
	me.Deck = []engine.Card{realCard(t, "A4b-069"), match, realCard(t, "A4b-069")}

	play311(t, g, p, trainerID, engine.NoSpot)

	if handHas311(me, match.Name) != 1 || len(me.Deck) != 2 {
		t.Errorf("hand %v, deck %d; want %s found", me.Hand, len(me.Deck), match.Name)
	}
}

func TestSerena(t *testing.T)            { searchFinds(t, "B1a-069", "B4a-108") } // Mega Mawile ex
func TestTeamGalacticGrunt(t *testing.T) { searchFinds(t, "A2-151", "B2-101") }   // Stunky

func TestTeamRocketsResearcher(t *testing.T) {
	g, p, me, _ := game311(t, "B4a-069")
	me.Deck = nil
	for range 6 {
		me.Deck = append(me.Deck, realCard(t, "B4a-008"), realCard(t, "A4b-069")) // Team Rocket's Houndour, Torkoal
	}
	torkoal := handHas311(me, "Torkoal")

	play311(t, g, p, "B4a-069", engine.NoSpot)

	want := min(6, headsBeforeTails(lastFlips311(g)))
	if got := handHas311(me, "Team Rocket's Houndour"); got != want || handHas311(me, "Torkoal") != torkoal {
		t.Errorf("found %d Houndour (flips %v), want %d and no Torkoal", got, lastFlips311(g), want)
	}
}
