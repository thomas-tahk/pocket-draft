package engine

import "testing"

func testTool(id string) Card {
	onHolder := func(c PassiveCheck) bool { return c.Source.Spot == c.Target.Spot }
	return Card{ID: id, Name: "Boots " + id, Kind: KindTool,
		Passive: []Passive{{Hook: HookRetreatCost, Value: 1, When: onHolder}}}
}

// testBlower discards a Tool from any Pokémon (Field Blower's Tool half).
func testBlower(id string) Card {
	return Card{ID: id, Name: "Blower " + id, Kind: KindItem, Trainer: []EffectOp{DiscardTool{}}}
}

var active = func(p int) Spot { return Spot{Player: p, Pos: 0} }

// A3: a Tool attaches to one of your Pokémon and stays there; its text applies.
func TestToolAttachesAndApplies(t *testing.T) {
	g, p := mainPhaseWith(t, testTool("T1"))
	before := g.RetreatCost(p)

	mustSubmit(t, g, PlayTrainer{Player: p, CardID: "T1", Target: active(p)})

	pl := &g.S.Players[p]
	if pl.Active.Tool == nil || pl.Active.Tool.ID != "T1" {
		t.Fatalf("Active's Tool = %v, want T1", pl.Active.Tool)
	}
	if len(pl.Discard) != 0 {
		t.Errorf("an attached Tool went to the discard: %v", pl.Discard)
	}
	if got := g.RetreatCost(p); got != before+1 {
		t.Errorf("retreat cost %d, want %d with the Tool", got, before+1)
	}
}

// A3: a Tool needs a target of your own with no Tool.
func TestToolOnlyOnOwnPokemonWithoutOne(t *testing.T) {
	g, p := mainPhaseWith(t, testTool("T1"), testTool("T2"))
	if isLegal(g, p, PlayTrainer{Player: p, CardID: "T1", Target: active(1 - p)}) {
		t.Error("a Tool was legal on the opponent's Pokémon")
	}
	if isLegal(g, p, PlayTrainer{Player: p, CardID: "T1", Target: NoSpot}) {
		t.Error("a Tool was legal with no target")
	}
	mustSubmit(t, g, PlayTrainer{Player: p, CardID: "T1", Target: active(p)})
	if isLegal(g, p, PlayTrainer{Player: p, CardID: "T2", Target: active(p)}) {
		t.Error("a second Tool was legal on the same Pokémon")
	}
}

// A3 done-gate: after a card removes the Tool, a new Tool is legal there.
func TestBlowerThenNewToolLegal(t *testing.T) {
	g, p := mainPhaseWith(t, testTool("T1"), testTool("T2"), testBlower("B1"))
	mustSubmit(t, g, PlayTrainer{Player: p, CardID: "T1", Target: active(p)})

	mustSubmit(t, g, PlayTrainer{Player: p, CardID: "B1", Target: active(p)})

	pl := &g.S.Players[p]
	if pl.Active.Tool != nil {
		t.Fatalf("Tool still attached after the Blower")
	}
	if !isLegal(g, p, PlayTrainer{Player: p, CardID: "T2", Target: active(p)}) {
		t.Error("a new Tool isn't legal after the old one was removed")
	}
	if n := len(pl.Discard); n != 2 {
		t.Errorf("discard has %d cards, want the Blower and the Tool", n)
	}
}

// The Blower has nothing to target when no Pokémon has a Tool.
func TestBlowerNeedsATool(t *testing.T) {
	g, p := mainPhaseWith(t, testBlower("B1"))
	for _, m := range g.LegalMoves(p) {
		if pt, ok := m.(PlayTrainer); ok && pt.CardID == "B1" {
			t.Errorf("Blower legal with no Tool in play: %+v", pt)
		}
	}
}

// A3: a Tool is discarded with its Pokémon when it is Knocked Out.
func TestToolDiscardedWithKnockedOutPokemon(t *testing.T) {
	g, p := mainPhaseWith(t, testTool("T1"))
	mustSubmit(t, g, PlayTrainer{Player: p, CardID: "T1", Target: active(p)})
	pl := &g.S.Players[p]
	pl.Active.Damage = pl.Active.Card.HP

	g.removeIfKnockedOut(p)

	if n := len(pl.Discard); n != 2 {
		t.Errorf("discard has %d cards, want the Pokémon and its Tool", n)
	}
}

// A Tool with no built text can't be played.
func TestUnbuiltToolNotPlayable(t *testing.T) {
	g, p := mainPhaseWith(t, Card{ID: "T9", Name: "Blank", Kind: KindTool})
	if g.Submit(PlayTrainer{Player: p, CardID: "T9", Target: active(p)}) == nil {
		t.Error("a Tool with no text was played")
	}
}
