package main

import (
	"strings"
	"testing"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// Item texts, batch 3.12 (issue #72): one test per wording, on a real card.

// itemGame is a game past setup where the player on turn holds the Item id.
func itemGame(t *testing.T, id string) (*engine.Game, int) {
	t.Helper()
	useRealCards(t)
	return trainerGame(t, realCard(t, id))
}

func playItem(t *testing.T, g *engine.Game, p int, id string, target engine.Spot) {
	t.Helper()
	if err := g.Submit(engine.PlayTrainer{Player: p, CardID: id, Target: target}); err != nil {
		t.Fatalf("play %s: %v", id, err)
	}
}

func blocked(t *testing.T, g *engine.Game, p int, id string) {
	t.Helper()
	if g.Submit(engine.PlayTrainer{Player: p, CardID: id, Target: engine.NoSpot}) == nil {
		t.Fatalf("%s was played when it shouldn't be", id)
	}
}

// onBoard puts the real card id at player p's pos (0 the Active, n bench
// slot n-1).
func onBoard(t *testing.T, g *engine.Game, p, pos int, id string) *engine.InPlay {
	t.Helper()
	mon := &engine.InPlay{Card: realCard(t, id), Energy: map[engine.Energy]int{}}
	pl := &g.S.Players[p]
	if pos == 0 {
		pl.Active = mon
	} else {
		pl.Bench[pos-1] = mon
	}
	return mon
}

func choose(t *testing.T, g *engine.Game, p, i int) {
	t.Helper()
	if err := g.Submit(engine.ChooseOption{Player: p, Index: i}); err != nil {
		t.Fatalf("choose %d: %v", i, err)
	}
}

func endTurn(t *testing.T, g *engine.Game) {
	t.Helper()
	if err := g.Submit(engine.EndTurn{Player: g.S.Active}); err != nil {
		t.Fatalf("end turn: %v", err)
	}
}

func lastFlips(g *engine.Game) []engine.CoinFlip { return g.Actions[len(g.Actions)-1].Flips }

func modifier(g *engine.Game, p, pos int, h engine.Hook) int {
	sum := 0
	for _, m := range g.Modifiers(p, pos) {
		if m.Hook == h {
			sum += m.Value
		}
	}
	return sum
}

func handHas(hand []engine.Card, name string) bool {
	for _, c := range hand {
		if c.Name == name {
			return true
		}
	}
	return false
}

func TestBeastWall(t *testing.T) {
	g, p := itemGame(t, "A3a-063")
	g.S.Players[1-p].Points = 1
	blocked(t, g, p, "A3a-063")

	g.S.Players[1-p].Points = 0
	onBoard(t, g, p, 0, "B2-014") // Buzzwole
	playItem(t, g, p, "A3a-063", engine.NoSpot)
	endTurn(t, g)

	if got := modifier(g, p, 0, engine.HookDamageTaken); got != -20 {
		t.Errorf("during the opponent's turn Buzzwole takes %+d, want -20", got)
	}
	endTurn(t, g)
	if got := modifier(g, p, 0, engine.HookDamageTaken); got != 0 {
		t.Errorf("Beast Wall still applies after the opponent's turn: %+d", got)
	}
}

func TestBeastWallSkipsNonUltraBeasts(t *testing.T) {
	g, p := itemGame(t, "A3a-063")
	playItem(t, g, p, "A3a-063", engine.NoSpot)
	if got := modifier(g, p, 0, engine.HookDamageTaken); got != 0 {
		t.Errorf("Torkoal takes %+d, want 0", got)
	}
}

func TestBigMalasada(t *testing.T) {
	g, p := itemGame(t, "A3-142")
	mon := g.S.Players[p].Active
	mon.Damage = 30
	mon.Inflict(engine.Poison)

	playItem(t, g, p, "A3-142", engine.NoSpot)

	if mon.Damage != 20 || mon.Has(engine.Poison) {
		t.Errorf("damage %d, poisoned %v; want 20 and cured", mon.Damage, mon.Has(engine.Poison))
	}
}

func TestClemontsBackpack(t *testing.T) {
	g, p := itemGame(t, "B1a-066")
	onBoard(t, g, p, 0, "B3-053") // Magneton

	playItem(t, g, p, "B1a-066", engine.NoSpot)

	if got := modifier(g, p, 0, engine.HookDamageDealt); got != 20 {
		t.Errorf("Magneton does %+d, want +20", got)
	}
	endTurn(t, g)
	if got := modifier(g, p, 0, engine.HookDamageDealt); got != 0 {
		t.Errorf("Clemont's Backpack outlasted the turn: %+d", got)
	}
}

func TestEeveeBagDamage(t *testing.T) {
	g, p := itemGame(t, "A4b-308")
	onBoard(t, g, p, 0, "B3a-008") // Vaporeon

	playItem(t, g, p, "A4b-308", engine.NoSpot)
	if pr := g.S.Pending; pr == nil || pr.Kind != engine.PromptChoose || len(pr.Labels) != 2 {
		t.Fatalf("pending = %+v, want a choice of 2", g.S.Pending)
	}
	choose(t, g, p, 0)

	if got := modifier(g, p, 0, engine.HookDamageDealt); got != 10 {
		t.Errorf("Vaporeon does %+d, want +10", got)
	}
}

func TestEeveeBagHeal(t *testing.T) {
	g, p := itemGame(t, "A4b-308")
	active := onBoard(t, g, p, 0, "B3a-008") // Vaporeon
	benched := onBoard(t, g, p, 1, "B3a-008")
	eevee := onBoard(t, g, p, 2, "B4-131")
	active.Damage, benched.Damage, eevee.Damage = 30, 10, 30

	playItem(t, g, p, "A4b-308", engine.NoSpot)
	choose(t, g, p, 1)

	if active.Damage != 10 || benched.Damage != 0 || eevee.Damage != 30 {
		t.Errorf("damage %d/%d/%d, want 10/0/30", active.Damage, benched.Damage, eevee.Damage)
	}
}

func TestElectricGenerator(t *testing.T) {
	g, p := itemGame(t, "B2a-086")
	pika := onBoard(t, g, p, 1, "B4-049")
	if g.Submit(engine.PlayTrainer{Player: p, CardID: "B2a-086", Target: engine.Spot{Player: p, Pos: 0}}) == nil {
		t.Fatal("Electric Generator targeted the Active Torkoal")
	}

	playItem(t, g, p, "B2a-086", engine.Spot{Player: p, Pos: 1})

	want := 0
	if lastFlips(g)[0].Heads {
		want = 1
	}
	if pika.Energy[engine.Lightning] != want {
		t.Errorf("Pikachu has %d Lightning, want %d", pika.Energy[engine.Lightning], want)
	}
}

func TestElementalSwitch(t *testing.T) {
	g, p := itemGame(t, "A4b-310")
	sq := onBoard(t, g, p, 1, "B1a-017") // Squirtle
	sq.Energy[engine.Fire], sq.Energy[engine.Water] = 1, 1
	active := g.S.Players[p].Active

	playItem(t, g, p, "A4b-310", engine.Spot{Player: p, Pos: 1})
	if got := g.S.Pending.Labels; len(got) != 2 || got[1] != "Water Energy" {
		t.Fatalf("labels = %v, want Fire then Water", got)
	}
	choose(t, g, p, 1)

	if sq.Energy[engine.Water] != 0 || active.Energy[engine.Water] != 1 || sq.Energy[engine.Fire] != 1 {
		t.Errorf("Squirtle %v, Active %v; want the Water moved", sq.Energy, active.Energy)
	}
}

func TestFishingNet(t *testing.T) {
	g, p := itemGame(t, "A3-143")
	pl := &g.S.Players[p]
	pl.Discard = []engine.Card{realCard(t, "B2b-007")} // Charmander only
	blocked(t, g, p, "A3-143")

	pl.Discard = append(pl.Discard, realCard(t, "B3a-008"), realCard(t, "B1a-017"))
	playItem(t, g, p, "A3-143", engine.NoSpot)

	if !handHas(pl.Hand, "Squirtle") {
		t.Error("Squirtle didn't come to hand")
	}
}

func TestFlamePatch(t *testing.T) {
	g, p := itemGame(t, "B1-217")
	pl := &g.S.Players[p]
	pl.DiscardedEnergy = []engine.Energy{engine.Water}
	blocked(t, g, p, "B1-217")

	pl.DiscardedEnergy = append(pl.DiscardedEnergy, engine.Fire)
	playItem(t, g, p, "B1-217", engine.NoSpot)

	if pl.Active.Energy[engine.Fire] != 1 || len(pl.DiscardedEnergy) != 1 {
		t.Errorf("Active %v, discarded %v; want one Fire attached", pl.Active.Energy, pl.DiscardedEnergy)
	}
}

func TestHandScope(t *testing.T) {
	g, p := itemGame(t, "P-A-003")
	g.S.Players[1-p].Hand = []engine.Card{realCard(t, "B1a-017"), realCard(t, "P-A-001")}

	playItem(t, g, p, "P-A-003", engine.NoSpot)

	if last := g.Narration[len(g.Narration)-1]; !strings.Contains(last, "Squirtle, Potion") {
		t.Errorf("narration %q doesn't reveal the hand", last)
	}
}

func TestHittingHammer(t *testing.T) {
	g, p := itemGame(t, "B1-215")
	blocked(t, g, p, "B1-215")
	opp := g.S.Players[1-p].Active
	opp.Energy[engine.Fire] = 1

	playItem(t, g, p, "B1-215", engine.NoSpot)

	f := lastFlips(g)
	want := 1
	if len(f) == 2 && f[0].Heads && f[1].Heads {
		want = 0
	}
	if len(f) != 2 || opp.Energy[engine.Fire] != want {
		t.Errorf("%d flips, %d Fire left; want 2 flips and %d", len(f), opp.Energy[engine.Fire], want)
	}
}

func TestLuckyIcePop(t *testing.T) {
	g, p := itemGame(t, "B2-145")
	pl := &g.S.Players[p]
	pl.Active.Damage = 30

	playItem(t, g, p, "B2-145", engine.NoSpot)

	if pl.Active.Damage != 10 {
		t.Errorf("damage %d, want 10", pl.Active.Damage)
	}
	if heads := lastFlips(g)[0].Heads; handHas(pl.Hand, "Lucky Ice Pop") != heads {
		t.Errorf("heads %v but Lucky Ice Pop in hand is %v", heads, !heads)
	}
}

func TestLuckyIcePopNoHealNoFlip(t *testing.T) {
	g, p := itemGame(t, "B2-145")
	playItem(t, g, p, "B2-145", engine.NoSpot)
	if n := len(lastFlips(g)); n != 0 {
		t.Errorf("%d flips with nothing healed, want 0", n)
	}
}

func TestMaintenance(t *testing.T) {
	g, p := itemGame(t, "B2b-066")
	pl := &g.S.Players[p]
	card := pl.Hand[len(pl.Hand)-1]
	pl.Hand = []engine.Card{realCard(t, "P-A-001"), card}
	blocked(t, g, p, "B2b-066")

	pl.Hand = append(pl.Hand, realCard(t, "P-A-001"))
	deck := len(pl.Deck)
	playItem(t, g, p, "B2b-066", engine.NoSpot)
	choose(t, g, p, 0) // the second card is chosen at once: one is left

	if len(pl.Hand) != 1 || len(pl.Deck) != deck+1 || g.S.Pending != nil {
		t.Errorf("hand %d, deck %d→%d; want 1 card and deck +1", len(pl.Hand), deck, len(pl.Deck))
	}
}

func TestMythicalSlab(t *testing.T) {
	g, p := itemGame(t, "A1a-065")
	pl := &g.S.Players[p]
	pl.Deck = []engine.Card{realCard(t, "B4-070"), realCard(t, "A4b-069")} // Mewtwo, Torkoal

	playItem(t, g, p, "A1a-065", engine.NoSpot)
	if !handHas(pl.Hand, "Mewtwo") {
		t.Fatal("Mewtwo didn't come to hand")
	}

	pl.Hand = append(pl.Hand, realCard(t, "A1a-065"))
	pl.Deck = append(pl.Deck, realCard(t, "B4-070"))
	playItem(t, g, p, "A1a-065", engine.NoSpot)
	if pl.Deck[len(pl.Deck)-1].Name != "Torkoal" || pl.Deck[0].Name != "Mewtwo" {
		t.Errorf("deck %v, want Torkoal moved to the bottom", joinCardNames(pl.Deck))
	}
}

func joinCardNames(cs []engine.Card) string {
	var s []string
	for _, c := range cs {
		s = append(s, c.Name)
	}
	return strings.Join(s, ",")
}

func TestNastyNotice(t *testing.T) {
	g, p := itemGame(t, "B2b-065")
	opp := &g.S.Players[1-p]
	opp.Hand = opp.Hand[:0]
	for len(opp.Hand) < 4 {
		opp.Hand = append(opp.Hand, realCard(t, "P-A-001"))
	}
	blocked(t, g, p, "B2b-065")

	opp.Hand = append(opp.Hand, realCard(t, "B1a-017"), realCard(t, "B2b-007"))
	playItem(t, g, p, "B2b-065", engine.NoSpot)
	if g.S.Pending == nil || g.S.Pending.Player != 1-p {
		t.Fatalf("the opponent isn't choosing: %+v", g.S.Pending)
	}
	choose(t, g, 1-p, 4)
	choose(t, g, 1-p, 4)

	if len(opp.Hand) != 4 || len(opp.Discard) != 2 || g.S.Pending != nil {
		t.Errorf("opponent hand %d, discard %d; want 4 and 2", len(opp.Hand), len(opp.Discard))
	}
}

func TestPokedex(t *testing.T) {
	g, p := itemGame(t, "P-A-004")
	pl := &g.S.Players[p]
	pl.Deck = []engine.Card{realCard(t, "B4-070"), realCard(t, "B1a-017"), realCard(t, "B4-049"), realCard(t, "B2b-007")}

	playItem(t, g, p, "P-A-004", engine.NoSpot)
	pr := g.S.Pending
	if pr == nil || pr.Player != p || !strings.HasSuffix(pr.Title, "Mewtwo, Squirtle, Pikachu") {
		t.Fatalf("prompt %+v, want the top 3 shown to the player", pr)
	}
	choose(t, g, p, 0)

	if g.S.Pending != nil || joinCardNames(pl.Deck) != "Mewtwo,Squirtle,Pikachu,Charmander" {
		t.Errorf("deck %s changed or prompt left open", joinCardNames(pl.Deck))
	}
}

// The cards a Pokédex shows reach only its player's screen.
func TestPokedexShownOnlyToItsPlayer(t *testing.T) {
	g, p := itemGame(t, "P-A-004")
	playItem(t, g, p, "P-A-004", engine.NoSpot)

	mine, theirs := toGameView(g, p), toGameView(g, 1-p)

	if mine.Pending == nil || mine.Pending.Kind != "choose" || mine.Pending.Title == "" || len(mine.Pending.Labels) != 1 {
		t.Errorf("player's prompt %+v, want the choice with its title", mine.Pending)
	}
	if theirs.Pending == nil || theirs.Pending.Title != "" || theirs.Pending.Labels != nil {
		t.Errorf("opponent sees %+v, want the prompt without its cards", theirs.Pending)
	}
}

// An answer that narrates nothing (Pokédex's OK) still sends result as [],
// never null: clients map over it.
func TestActionResultIsNeverNull(t *testing.T) {
	g, p := itemGame(t, "P-A-004")
	playItem(t, g, p, "P-A-004", engine.NoSpot)
	choose(t, g, p, 0)

	for _, viewer := range []int{p, 1 - p} {
		acts := toGameView(g, viewer).Actions
		last := acts[len(acts)-1]
		if last.Kind != "choose" || last.Result == nil {
			t.Errorf("viewer %d: last action %+v, want a choose with result []", viewer, last)
		}
	}
}

func TestPokemonFlute(t *testing.T) {
	g, p := itemGame(t, "A1a-064")
	opp := &g.S.Players[1-p]
	opp.Discard = []engine.Card{realCard(t, "P-A-001"), realCard(t, "B3a-008")}
	blocked(t, g, p, "A1a-064")

	opp.Discard = append(opp.Discard, realCard(t, "B1a-017"))
	playItem(t, g, p, "A1a-064", engine.NoSpot)

	if b := opp.Bench[0]; b == nil || b.Card.Name != "Squirtle" || len(opp.Discard) != 2 {
		t.Errorf("opponent bench %+v, discard %d; want Squirtle benched", b, len(opp.Discard))
	}
}

func TestPotion(t *testing.T) {
	g, p := itemGame(t, "P-A-001")
	if g.WhyCantPlay(p, "P-A-001") == nil {
		t.Error("Potion is playable with nothing damaged")
	}
	mon := g.S.Players[p].Active
	mon.Damage = 30

	playItem(t, g, p, "P-A-001", engine.Spot{Player: p, Pos: 0})

	if mon.Damage != 10 {
		t.Errorf("damage %d, want 10", mon.Damage)
	}
}

func TestPrankSpinner(t *testing.T) {
	g, p := itemGame(t, "B1-213")
	me, opp := &g.S.Players[p], &g.S.Players[1-p]
	hands := len(me.Hand) - 1 + len(opp.Hand)
	decks := len(me.Deck) + len(opp.Deck)

	playItem(t, g, p, "B1-213", engine.NoSpot)

	if len(me.Hand)+len(opp.Hand) != hands-1 || len(me.Deck)+len(opp.Deck) != decks+1 {
		t.Errorf("hands %d→%d, decks %d→%d; want one card moved", hands, len(me.Hand)+len(opp.Hand), decks, len(me.Deck)+len(opp.Deck))
	}
}

func TestRareCandy(t *testing.T) {
	g, p := itemGame(t, "A4b-314")
	pl := &g.S.Players[p]
	pl.Hand = append(pl.Hand, realCard(t, "B1a-003")) // Venusaur
	bulba := onBoard(t, g, p, 1, "B1a-001")
	g.S.Turn = 3
	bulba.PlacedTurn = 3
	blocked(t, g, p, "A4b-314") // put into play this turn

	bulba.PlacedTurn = 1
	playItem(t, g, p, "A4b-314", engine.Spot{Player: p, Pos: 1})

	if bulba.Card.Name != "Venusaur" || handHas(pl.Hand, "Venusaur") {
		t.Errorf("Bulbasaur is %s, want Venusaur from hand", bulba.Card.Name)
	}
}

func TestRareCandyNotFirstTurn(t *testing.T) {
	g, p := itemGame(t, "A4b-314")
	g.S.Players[p].Hand = append(g.S.Players[p].Hand, realCard(t, "B1a-003"))
	onBoard(t, g, p, 1, "B1a-001")
	if g.WhyCantPlay(p, "A4b-314") == nil {
		t.Error("Rare Candy is playable on the first turn")
	}
}

func TestRedCard(t *testing.T) {
	g, p := itemGame(t, "P-A-006")
	opp := &g.S.Players[1-p]
	total := len(opp.Hand) + len(opp.Deck)

	playItem(t, g, p, "P-A-006", engine.NoSpot)

	if len(opp.Hand) != 3 || len(opp.Hand)+len(opp.Deck) != total {
		t.Errorf("opponent hand %d, total %d→%d; want 3 and unchanged", len(opp.Hand), total, len(opp.Hand)+len(opp.Deck))
	}
}

func TestRepel(t *testing.T) {
	g, p := itemGame(t, "A3a-064")
	blocked(t, g, p, "A3a-064") // nothing on their Bench
	opp := &g.S.Players[1-p]
	old := opp.Active
	old.Inflict(engine.Poison)
	onBoard(t, g, 1-p, 1, "B1a-017")

	playItem(t, g, p, "A3a-064", engine.NoSpot)

	if opp.Active.Card.Name != "Squirtle" || opp.Bench[0] != old || old.Has(engine.Poison) {
		t.Errorf("opponent Active %s, bench %+v; want Squirtle in, cured Torkoal benched", opp.Active.Card.Name, opp.Bench[0])
	}
}

func TestRepelNeedsBasic(t *testing.T) {
	g, p := itemGame(t, "A3a-064")
	onBoard(t, g, 1-p, 0, "B3a-008") // Vaporeon
	onBoard(t, g, 1-p, 1, "B1a-017")
	blocked(t, g, p, "A3a-064")
}

func TestRotomDex(t *testing.T) {
	g, p := itemGame(t, "A3-145")
	pl := &g.S.Players[p]
	pl.Deck = []engine.Card{realCard(t, "B4-070"), realCard(t, "B1a-017")}

	playItem(t, g, p, "A3-145", engine.NoSpot)
	if pr := g.S.Pending; pr == nil || !strings.HasSuffix(pr.Title, "Mewtwo") || len(pr.Labels) != 2 {
		t.Fatalf("prompt %+v, want the top card and a shuffle choice", pr)
	}
	choose(t, g, p, 1)

	if joinCardNames(pl.Deck) != "Mewtwo,Squirtle" {
		t.Errorf("deck %s, want unshuffled", joinCardNames(pl.Deck))
	}
}

func TestSquirtBottle(t *testing.T) {
	g, p := itemGame(t, "A4-152")
	opp := &g.S.Players[1-p]
	opp.Active.Energy[engine.Water] = 1
	blocked(t, g, p, "A4-152")

	opp.Active.Energy[engine.Fire] = 1
	playItem(t, g, p, "A4-152", engine.NoSpot)

	if opp.Active.Energy[engine.Fire] != 0 || opp.Active.Energy[engine.Water] != 1 || len(opp.DiscardedEnergy) != 1 {
		t.Errorf("energy %v, discarded %v; want the Fire discarded", opp.Active.Energy, opp.DiscardedEnergy)
	}
}

func TestGooZooka(t *testing.T) {
	g, p := itemGame(t, "B4a-068")
	base := g.RetreatCost(1 - p)

	playItem(t, g, p, "B4a-068", engine.NoSpot)
	endTurn(t, g)
	if got := g.RetreatCost(1 - p); got != base+1 {
		t.Errorf("opponent's retreat cost %d during their turn, want %d", got, base+1)
	}
	endTurn(t, g)
	if got := g.RetreatCost(1 - p); got != base {
		t.Errorf("retreat cost %d after their turn, want %d", got, base)
	}
}

func TestThievingMachine(t *testing.T) {
	g, p := itemGame(t, "B4a-067")
	opp := &g.S.Players[1-p]
	opp.Discard = []engine.Card{realCard(t, "B4a-067"), realCard(t, "B1a-017")}
	blocked(t, g, p, "B4a-067")

	opp.Discard = append(opp.Discard, realCard(t, "P-A-001"))
	playItem(t, g, p, "B4a-067", engine.NoSpot)

	if !handHas(g.S.Players[p].Hand, "Potion") || len(opp.Discard) != 2 {
		t.Errorf("Potion not taken; opponent discard %d", len(opp.Discard))
	}
}

func TestOrderPad(t *testing.T) {
	g, p := itemGame(t, "B4-145")
	pl := &g.S.Players[p]
	pl.Deck = []engine.Card{realCard(t, "B1a-017"), realCard(t, "P-A-001")}

	playItem(t, g, p, "B4-145", engine.NoSpot)

	if heads := lastFlips(g)[0].Heads; handHas(pl.Hand, "Potion") != heads {
		t.Errorf("heads %v but Potion in hand is %v", heads, !heads)
	}
	if handHas(pl.Hand, "Squirtle") {
		t.Error("a Pokémon came to hand")
	}
}

func TestPokemonCommunication(t *testing.T) {
	g, p := itemGame(t, "A4b-316")
	pl := &g.S.Players[p]
	pl.Hand = []engine.Card{realCard(t, "B1a-017"), realCard(t, "A4b-316")}
	pl.Deck = []engine.Card{realCard(t, "P-A-001")}
	blocked(t, g, p, "A4b-316") // no Pokémon in the deck

	pl.Deck = append(pl.Deck, realCard(t, "B4-070"))
	playItem(t, g, p, "A4b-316", engine.NoSpot)

	if !handHas(pl.Hand, "Mewtwo") || handHas(pl.Hand, "Squirtle") || !handHas(pl.Deck, "Squirtle") {
		t.Errorf("hand %s, deck %s; want Squirtle and Mewtwo switched", joinCardNames(pl.Hand), joinCardNames(pl.Deck))
	}
}

func TestQuickGrowExtract(t *testing.T) {
	g, p := itemGame(t, "B1a-067")
	pl := &g.S.Players[p]
	pl.Deck = append([]engine.Card{realCard(t, "B1a-002")}, pl.Deck...) // Ivysaur
	bulba := onBoard(t, g, p, 1, "B1a-001")
	if g.WhyCantPlay(p, "B1a-067") == nil {
		t.Error("Quick-Grow Extract is playable on the first turn")
	}
	g.S.Turn, bulba.PlacedTurn = 3, 3
	blocked(t, g, p, "B1a-067") // put into play this turn

	bulba.PlacedTurn = 1
	playItem(t, g, p, "B1a-067", engine.Spot{Player: p, Pos: 1})

	if bulba.Card.Name != "Ivysaur" || handHas(pl.Deck, "Ivysaur") {
		t.Errorf("Bulbasaur is %s, want Ivysaur from the deck", bulba.Card.Name)
	}
}

func TestXSpeed(t *testing.T) {
	g, p := itemGame(t, "P-A-002")
	base := g.RetreatCost(p)

	playItem(t, g, p, "P-A-002", engine.NoSpot)
	if got := g.RetreatCost(p); got != base-1 {
		t.Errorf("retreat cost %d, want %d", got, base-1)
	}
	endTurn(t, g)
	if got := g.RetreatCost(p); got != base {
		t.Errorf("X Speed outlasted the turn: %d", got)
	}
}
