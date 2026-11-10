package engine

import (
	"errors"
	"fmt"
)

// WhyCantPlay is the one reason hand card cardID has nowhere to go right now,
// or nil when LegalMoves lists a move with it. The screen flashes it when the
// card is picked up (spec F17).
func (g *Game) WhyCantPlay(p int, cardID string) error {
	for _, mv := range g.LegalMoves(p) {
		switch m := mv.(type) {
		case SetupActive:
			if m.CardID == cardID {
				return nil
			}
		case SetupBench:
			if m.CardID == cardID {
				return nil
			}
		case PlayBasic:
			if m.CardID == cardID {
				return nil
			}
		case Evolve:
			if m.HandCardID == cardID {
				return nil
			}
		case PlayTrainer:
			if m.CardID == cardID {
				return nil
			}
		}
	}
	pl := &g.S.Players[p]
	idx := indexInHand(pl.Hand, cardID)
	if idx < 0 {
		return errors.New("card is not in hand")
	}
	if pl.Hand[idx].IsTrainer() {
		return g.whyNoTrainer(p, cardID)
	}
	if pl.Hand[idx].Stage == Basic {
		return g.whyNoBench(p, cardID)
	}
	return g.whyNoEvolve(p, cardID)
}

// whyNoBench: a Basic with no legal slot. If slot 0 is refused only for being
// taken, every slot is (any free one would be legal), so the Bench is full.
func (g *Game) whyNoBench(p int, cardID string) error {
	err := g.checkPlayBasic(PlayBasic{Player: p, CardID: cardID, Slot: 0})
	if errors.Is(err, errSlotTaken) {
		return errors.New("your Bench is full")
	}
	return err
}

// whyNoEvolve: a target-independent reason first, then why the first Pokémon
// it evolves from can't take it, else that no such Pokémon is in play.
func (g *Game) whyNoEvolve(p int, cardID string) error {
	ev, err := g.evolveGate(p, cardID)
	if err != nil {
		return err
	}
	pl := &g.S.Players[p]
	for _, t := range targets() {
		if mon := targetMon(pl, t); mon != nil && (mon.Card.Name == ev.EvolvesFrom || g.evolvesAs(g.holder(p, t), ev.EvolvesFrom)) {
			return g.checkEvolve(Evolve{Player: p, HandCardID: cardID, Target: t})
		}
	}
	if g.S.Turn <= 2 {
		return errFirstTurnEvolve
	}
	return fmt.Errorf("no %s in play to evolve into %s", ev.EvolvesFrom, ev.Name)
}

// WhyCantAttach is the reason this turn's energy can't go on any Pokémon, or
// nil when LegalMoves lists an AttachEnergy.
func (g *Game) WhyCantAttach(p int) error {
	for _, mv := range g.LegalMoves(p) {
		if _, ok := mv.(AttachEnergy); ok {
			return nil
		}
	}
	return g.checkAttach(AttachEnergy{Player: p, Target: 0})
}
