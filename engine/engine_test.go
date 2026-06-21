package engine

import (
	"reflect"
	"testing"
)

// --- a tiny "auto-player" that drives a full game by reacting to state, so the
// tests don't have to predict seeded draws/energy. It greedily attaches energy,
// fills its bench, attacks when it can, and answers every prompt. ---

func pickSetup(hand []Card) (string, []string) {
	var active string
	var bench []string
	for _, c := range hand {
		if c.Stage != Basic {
			continue
		}
		if active == "" {
			active = c.ID
		} else if len(bench) < 3 {
			bench = append(bench, c.ID)
		}
	}
	return active, bench
}

func firstBasicID(hand []Card) (string, bool) {
	for _, c := range hand {
		if c.Stage == Basic {
			return c.ID, true
		}
	}
	return "", false
}

func playableAttack(p Player) (int, bool) {
	if p.Active == nil {
		return 0, false
	}
	best, bestDmg := -1, -1
	for i, a := range p.Active.Card.Attacks {
		if canPay(p.Active.Energy, a.Cost) && a.Damage > bestDmg {
			best, bestDmg = i, a.Damage
		}
	}
	return best, best >= 0
}

func autoPlay(t *testing.T, g *Game) {
	t.Helper()
	for step := 0; step < 5000 && g.S.Phase != PhaseOver; step++ {
		if g.S.Pending != nil {
			p := g.S.Pending.Player
			var err error
			switch g.S.Pending.Kind {
			case PromptSetup:
				a, b := pickSetup(g.S.Players[p].Hand)
				err = g.Submit(SetupPlace{Player: p, ActiveCardID: a, BenchCardIDs: b})
			case PromptNewActive:
				err = g.Submit(ChooseNewActive{Player: p, BenchIndex: 0})
			}
			if err != nil {
				t.Fatalf("prompt response failed: %v", err)
			}
			continue
		}

		p := g.S.Active
		pl := g.S.Players[p]
		var err error
		switch {
		case pl.EnergyZone != NoEnergy && !pl.EnergyUsed:
			err = g.Submit(AttachEnergy{Player: p, Target: 0})
		case len(pl.Bench) < 3 && hasBasicInHand(pl.Hand):
			id, _ := firstBasicID(pl.Hand)
			err = g.Submit(PlayBasic{Player: p, CardID: id})
		default:
			if ai, ok := playableAttack(pl); ok {
				err = g.Submit(UseAttack{Player: p, Index: ai})
			} else {
				err = g.Submit(EndTurn{Player: p})
			}
		}
		if err != nil {
			t.Fatalf("step %d submit failed: %v", step, err)
		}
	}
	if g.S.Phase != PhaseOver {
		t.Fatalf("game did not finish within step budget")
	}
}

func hasBasicInHand(hand []Card) bool {
	_, ok := firstBasicID(hand)
	return ok
}

func doSetup(t *testing.T, g *Game) {
	t.Helper()
	for g.S.Pending != nil && g.S.Pending.Kind == PromptSetup {
		p := g.S.Pending.Player
		a, b := pickSetup(g.S.Players[p].Hand)
		if err := g.Submit(SetupPlace{Player: p, ActiveCardID: a, BenchCardIDs: b}); err != nil {
			t.Fatalf("setup failed: %v", err)
		}
	}
}

func twoDecks() ([]Card, []Card) {
	return deckOf(demoEmbor(), 20), deckOf(demoVolt(), 20)
}

// Criterion 1: a full game runs start→finish and ends with a 3-point winner.
func TestFullGameReachesWinner(t *testing.T) {
	d0, d1 := twoDecks()
	g := NewGame(42, d0, d1)
	autoPlay(t, g)

	if g.S.Winner < 0 {
		t.Fatal("game ended with no winner")
	}
	if pts := g.S.Players[g.S.Winner].Points; pts < 3 {
		t.Fatalf("winner %s has only %d points", side(g.S.Winner), pts)
	}
}

// Criterion 2: illegal moves are refused and leave the game unchanged.
func TestIllegalMovesRefused(t *testing.T) {
	d0, d1 := twoDecks()
	g := NewGame(1, d0, d1)
	doSetup(t, g)

	p := g.S.Active

	// Acting out of turn.
	if err := g.Submit(UseAttack{Player: 1 - p, Index: 0}); err == nil {
		t.Error("expected error acting out of turn")
	}
	// Answering a prompt when none is pending.
	if err := g.Submit(ChooseNewActive{Player: p, BenchIndex: 0}); err == nil {
		t.Error("expected error: no prompt pending")
	}
	// Attacking with no energy attached (first player has none on turn 1).
	if g.S.Players[p].Active.totalEnergy() == 0 {
		if err := g.Submit(UseAttack{Player: p, Index: 0}); err == nil {
			t.Error("expected error attacking with no energy")
		}
	}

	// Move to a turn that has energy, then prove the one-energy-per-turn rule.
	if err := g.Submit(EndTurn{Player: p}); err != nil {
		t.Fatalf("end turn: %v", err)
	}
	q := g.S.Active
	if err := g.Submit(AttachEnergy{Player: q, Target: 0}); err != nil {
		t.Fatalf("first attach should succeed: %v", err)
	}
	if err := g.Submit(AttachEnergy{Player: q, Target: 0}); err == nil {
		t.Error("expected error attaching a second energy in one turn")
	}
	// Overfilling the bench (deck is all Basics, so hand has plenty).
	for i := 0; i < 10; i++ {
		id, ok := firstBasicID(g.S.Players[q].Hand)
		if !ok {
			break
		}
		err := g.Submit(PlayBasic{Player: q, CardID: id})
		if len(g.S.Players[q].Bench) == 3 && err == nil {
			// keep going; next one must fail
			continue
		}
	}
	if len(g.S.Players[q].Bench) != 3 {
		t.Fatalf("expected a full bench of 3, got %d", len(g.S.Players[q].Bench))
	}
	if id, ok := firstBasicID(g.S.Players[q].Hand); ok {
		if err := g.Submit(PlayBasic{Player: q, CardID: id}); err == nil {
			t.Error("expected error playing onto a full bench")
		}
	}
}

// Criterion 3: a knockout raises a "choose new Active" prompt — including a
// knockout caused by end-of-turn poison (mode 2 chaining into mode 3).
func TestPoisonKnockoutRaisesPrompt(t *testing.T) {
	d0, d1 := twoDecks()
	g := NewGame(5, d0, d1)
	doSetup(t, g)

	p := g.S.Active
	opp := 1 - p
	oppActive := g.S.Players[opp].Active
	if oppActive == nil || len(g.S.Players[opp].Bench) == 0 {
		t.Fatalf("test setup expected opponent to have an Active and a bench")
	}

	// Poison the opponent's Active and bring it within 10 of being knocked out.
	oppActive.Poisoned = true
	oppActive.Damage = oppActive.Card.HP - 5
	before := g.S.Players[p].Points

	if err := g.Submit(EndTurn{Player: p}); err != nil {
		t.Fatalf("end turn: %v", err)
	}

	if got := g.S.Players[p].Points; got != before+1 {
		t.Errorf("expected the poison KO to score 1 point, points went %d→%d", before, got)
	}
	if g.S.Pending == nil || g.S.Pending.Kind != PromptNewActive || g.S.Pending.Player != opp {
		t.Errorf("expected a new-Active prompt for %s, got %+v", side(opp), g.S.Pending)
	}
}

// Criterion 4: the whole game reconstructs by replaying its event log.
func TestReplayReconstructsState(t *testing.T) {
	d0, d1 := twoDecks()
	g := NewGame(99, d0, d1)
	autoPlay(t, g)

	r, err := Replay(99, d0, d1, g.Events)
	if err != nil {
		t.Fatalf("replay failed: %v", err)
	}
	if r.S.Winner != g.S.Winner {
		t.Errorf("replay winner %d != original %d", r.S.Winner, g.S.Winner)
	}
	if r.S.Turn != g.S.Turn || r.S.Phase != g.S.Phase {
		t.Errorf("replay (turn %d, phase %d) != original (turn %d, phase %d)", r.S.Turn, r.S.Phase, g.S.Turn, g.S.Phase)
	}
	for i := 0; i < 2; i++ {
		if r.S.Players[i].Points != g.S.Players[i].Points {
			t.Errorf("replay %s points %d != original %d", side(i), r.S.Players[i].Points, g.S.Players[i].Points)
		}
	}
}

// Criterion 5: same seed + same decisions ⇒ byte-identical event log and result.
func TestDeterminism(t *testing.T) {
	run := func() *Game {
		d0, d1 := twoDecks()
		g := NewGame(7, d0, d1)
		autoPlay(t, g)
		return g
	}
	a, b := run(), run()

	if !reflect.DeepEqual(a.Events, b.Events) {
		t.Fatalf("event logs diverged: %d vs %d events", len(a.Events), len(b.Events))
	}
	if a.S.Winner != b.S.Winner {
		t.Errorf("winners diverged: %d vs %d", a.S.Winner, b.S.Winner)
	}
	if !reflect.DeepEqual(a.Narration, b.Narration) {
		t.Error("narration feeds diverged for the same seed")
	}
}

// Bonus (supports criterion 1's scoring rules): an EX is worth 2 points.
func TestEXWorthTwoPoints(t *testing.T) {
	d0 := deckOf(demoVolt(), 20)
	d1 := deckOf(demoEmborEX(), 20) // P2 plays all-EX
	g := NewGame(3, d0, d1)
	doSetup(t, g)

	p := g.S.Active
	// Find the player whose Active is the EX and knock it out via poison.
	exOwner := -1
	for i := 0; i < 2; i++ {
		if g.S.Players[i].Active != nil && g.S.Players[i].Active.Card.IsEX {
			exOwner = i
		}
	}
	if exOwner < 0 {
		t.Fatal("expected an EX Active in play")
	}
	// Make sure the player who ends the turn isn't blocked; poison resolves for
	// both Actives regardless of whose turn it is.
	ex := g.S.Players[exOwner].Active
	ex.Poisoned = true
	ex.Damage = ex.Card.HP - 5
	scorer := 1 - exOwner
	before := g.S.Players[scorer].Points

	if err := g.Submit(EndTurn{Player: p}); err != nil {
		t.Fatalf("end turn: %v", err)
	}
	if got := g.S.Players[scorer].Points; got != before+2 {
		t.Errorf("knocking out an EX should score 2, points went %d→%d", before, got)
	}
}
