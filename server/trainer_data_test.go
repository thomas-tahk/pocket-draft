package main

import (
	"fmt"
	"testing"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// realCard converts the real printing id from public/data/cards.json.
func realCard(t *testing.T, id string) engine.Card {
	t.Helper()
	rc, ok := rawByID[id]
	if !ok {
		t.Fatalf("card %s not in cards.json", id)
	}
	c, err := toEngineCard(rc)
	if err != nil {
		t.Fatalf("convert %s: %v", id, err)
	}
	return c
}

// A1–A3: every Item and Supporter in the data loads as a Trainer of its kind.
func TestAllItemsAndSupportersLoad(t *testing.T) {
	useRealCards(t)
	want := map[string]engine.CardKind{"Item": engine.KindItem, "Supporter": engine.KindSupporter}
	n := 0
	for _, rc := range rawByID {
		kind, ok := want[rc.TrainerKind]
		if !ok {
			continue
		}
		n++
		c, err := toEngineCard(rc)
		if err != nil {
			t.Errorf("%s (%s) doesn't load: %v", rc.ID, rc.Name, err)
			continue
		}
		if c.Kind != kind || c.IsBasicPokemon() {
			t.Errorf("%s (%s) kind = %v, want %v", rc.ID, rc.Name, c.Kind, kind)
		}
	}
	if n != 263 {
		t.Errorf("found %d Items and Supporters, want 263", n)
	}
}

// Tools and Stadiums have no trainerKind in the scrape; they load by text.
func TestToolsAndStadiumsLoad(t *testing.T) {
	useRealCards(t)
	cases := map[string]engine.CardKind{
		"A4b-322": engine.KindTool,    // Rocky Helmet
		"A2-149":  engine.KindTool,    // Lum Berry
		"B3-155":  engine.KindStadium, // Bounded Field
		"B2b-069": engine.KindStadium, // Hiking Trail
	}
	for id, kind := range cases {
		if c := realCard(t, id); c.Kind != kind {
			t.Errorf("%s (%s) kind = %v, want %v", id, c.Name, c.Kind, kind)
		}
	}
}

// supporterGame is a game past setup where the player on turn holds trainer.
func trainerGame(t *testing.T, trainer engine.Card) (*engine.Game, int) {
	t.Helper()
	basic := realCard(t, "A4b-069") // Torkoal
	g := engine.NewGame(7, deckOfCard(basic, 20), deckOfCard(basic, 20))
	doSetup(t, g)
	p := g.S.Active
	g.S.Players[p].Hand = append(g.S.Players[p].Hand, trainer)
	return g, p
}

// "Draw 2 cards." — Professor's Research (A4b-373).
func TestProfessorsResearchDrawsTwo(t *testing.T) {
	useRealCards(t)
	g, p := trainerGame(t, realCard(t, "A4b-373"))
	pl := &g.S.Players[p]
	hand, deck := len(pl.Hand), len(pl.Deck)

	if err := g.Submit(engine.PlayTrainer{Player: p, CardID: "A4b-373", Target: engine.NoSpot}); err != nil {
		t.Fatalf("play Professor's Research: %v", err)
	}

	if len(pl.Hand) != hand-1+2 || len(pl.Deck) != deck-2 {
		t.Errorf("hand %d→%d, deck %d→%d; want +1 net and deck −2", hand, len(pl.Hand), deck, len(pl.Deck))
	}
}

// T3: with 10 cards in hand, Professor's Research can't be played and stays.
func TestProfessorsResearchBlockedAtTen(t *testing.T) {
	useRealCards(t)
	g, p := trainerGame(t, realCard(t, "A4b-373"))
	pl := &g.S.Players[p]
	for len(pl.Hand) < 10 {
		pl.Hand = append(pl.Hand, pl.Deck[0])
		pl.Deck = pl.Deck[1:]
	}

	if g.Submit(engine.PlayTrainer{Player: p, CardID: "A4b-373", Target: engine.NoSpot}) == nil {
		t.Error("Professor's Research was played with 10 cards in hand")
	}
	if g.WhyCantPlay(p, "A4b-373") == nil {
		t.Error("no reason given for the blocked card")
	}
}

// Every printing of a wording plays the same text.
func TestTrainerTextMapsEveryPrinting(t *testing.T) {
	useRealCards(t)
	for _, rc := range rawByID {
		if rc.CardType != "Trainer" || trainerEffects[rc.TrainerText] == nil {
			continue
		}
		if c := realCard(t, rc.ID); c.Trainer == nil {
			t.Errorf("%s (%s) has built text but no effect", rc.ID, rc.Name)
		}
	}
}

// A deck keeps its Trainers, and they don't count as its Basic.
func TestDeckFromIDsKeepsTrainers(t *testing.T) {
	rawByID = map[string]rawCard{}
	ids := make([]string, 0, 20)
	for i := 0; i < 9; i++ {
		b := fmt.Sprintf("B%d", i)
		rawByID[b] = rawCard{ID: b, Name: "Basic" + b, CardType: "Fire", Stage: "Basic", HP: 60,
			Attacks: []rawAttack{{Cost: "R", Name: "Hit", Damage: "20"}}}
		ids = append(ids, b, b)
	}
	rawByID["PR"] = rawCard{ID: "PR", Name: "Professor's Research", CardType: "Trainer", TrainerKind: "Supporter",
		TrainerText: "Draw 2 cards."}
	ids = append(ids, "PR", "PR")

	deck, err := deckFromIDs(ids)
	if err != nil {
		t.Fatalf("deck with Trainers: %v", err)
	}
	if len(deck) != 20 {
		t.Fatalf("got %d cards, want 20", len(deck))
	}

	noBasic := make([]string, 0, 20)
	for i := 0; i < 10; i++ {
		id := fmt.Sprintf("T%d", i)
		rawByID[id] = rawCard{ID: id, Name: "Item" + id, CardType: "Trainer", TrainerKind: "Item",
			TrainerText: "Draw 2 cards."}
		noBasic = append(noBasic, id, id)
	}
	if _, err := deckFromIDs(noBasic); err == nil {
		t.Error("an all-Trainer deck passed as having a Basic")
	}
}

// A PlayTrainer move survives the trip through the move request, with and
// without a target.
func TestPlayTrainerRequestRoundTrips(t *testing.T) {
	for _, ev := range []engine.PlayTrainer{
		{Player: 0, CardID: "A4b-373", Target: engine.NoSpot},
		{Player: 1, CardID: "P-A-001", Target: engine.Spot{Player: 1, Pos: 2}},
	} {
		got, err := toEvent(fromEvent(ev))
		if err != nil || got != ev {
			t.Errorf("round trip of %+v = %+v, %v", ev, got, err)
		}
	}
}

// "Discard a Pokémon Tool card from a Pokémon (yours or your opponent's), or
// discard a Stadium card in play." — Field Blower (B3-147): either target, and
// a new Tool is legal once the old one is gone (spec A3).
func TestFieldBlower(t *testing.T) {
	useRealCards(t)
	boots := engine.Card{ID: "T1", Name: "Boots", Kind: engine.KindTool,
		Passive: []engine.Passive{{Hook: engine.HookRetreatCost, Value: 1}}}
	stadium := engine.Card{ID: "S1", Name: "Field", Kind: engine.KindStadium,
		Passive: []engine.Passive{{Hook: engine.HookRetreatCost, Value: 1}}}
	blower := realCard(t, "B3-147")
	g, p := trainerGame(t, blower)
	pl := &g.S.Players[p]
	pl.Hand = append(pl.Hand, boots, stadium, blower, boots)
	active := engine.Spot{Player: p, Pos: 0}
	for _, ev := range []engine.PlayTrainer{
		{Player: p, CardID: "T1", Target: active},
		{Player: p, CardID: "S1", Target: engine.NoSpot},
		{Player: p, CardID: "B3-147", Target: active},
		{Player: p, CardID: "B3-147", Target: engine.StadiumSpot},
		{Player: p, CardID: "T1", Target: active},
	} {
		if err := g.Submit(ev); err != nil {
			t.Fatalf("%+v: %v", ev, err)
		}
	}
	if g.S.Stadium != nil || pl.Active.Tool == nil {
		t.Errorf("Stadium %v, Tool %v; want no Stadium and the new Tool", g.S.Stadium, pl.Active.Tool)
	}
}
