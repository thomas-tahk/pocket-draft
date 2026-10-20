package engine

import "errors"

// Before Ready a placed card is not final (owner 2026-10-19 (8)): it can go
// back to the hand, or move to another spot. Spots are numbered like
// AttachEnergy.Target: 0 is the Active, n is Bench slot n-1.

// SetupReturn takes the Pokémon at Pos back to the hand.
type SetupReturn struct {
	Player int
	Pos    int
}

// SetupMove moves the Pokémon at From to To; if To holds one too, they swap.
type SetupMove struct {
	Player   int
	From, To int
}

func (SetupReturn) event() {}
func (SetupMove) event()   {}

func setupSpot(pl *Player, pos int) **InPlay {
	if pos == 0 {
		return &pl.Active
	}
	return &pl.Bench[pos-1]
}

func validSpot(pos int) bool { return pos >= 0 && pos <= BenchSlots }

func listSetupReturn(g *Game, p int) []SetupReturn {
	var out []SetupReturn
	for pos := 0; pos <= BenchSlots; pos++ {
		out = append(out, SetupReturn{Player: p, Pos: pos})
	}
	return out
}

func listSetupMove(g *Game, p int) []SetupMove {
	var out []SetupMove
	for from := 0; from <= BenchSlots; from++ {
		for to := 0; to <= BenchSlots; to++ {
			out = append(out, SetupMove{Player: p, From: from, To: to})
		}
	}
	return out
}

func (g *Game) checkSetupReturn(e SetupReturn) error {
	if err := g.onSetup(e.Player); err != nil {
		return err
	}
	if !validSpot(e.Pos) || *setupSpot(&g.S.Players[e.Player], e.Pos) == nil {
		return errors.New("no Pokémon there to take back")
	}
	return nil
}

func (g *Game) checkSetupMove(e SetupMove) error {
	if err := g.onSetup(e.Player); err != nil {
		return err
	}
	if !validSpot(e.From) || !validSpot(e.To) || e.From == e.To {
		return errors.New("invalid spot to move between")
	}
	if *setupSpot(&g.S.Players[e.Player], e.From) == nil {
		return errors.New("no Pokémon there to move")
	}
	return nil
}

func (g *Game) applySetupReturn(e SetupReturn) {
	pl := &g.S.Players[e.Player]
	spot := setupSpot(pl, e.Pos)
	pl.Hand = append(pl.Hand, (*spot).Card)
	*spot = nil
	g.narrate("%s took a Pokémon back.", side(e.Player))
}

func (g *Game) applySetupMove(e SetupMove) {
	pl := &g.S.Players[e.Player]
	from, to := setupSpot(pl, e.From), setupSpot(pl, e.To)
	*from, *to = *to, *from
	g.narrate("%s moved a Pokémon.", side(e.Player))
}
