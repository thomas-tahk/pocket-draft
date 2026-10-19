package engine

import (
	"errors"
	"fmt"
)

// Stadiums (spec A6): one in play, shared by both players; its Card.Passive
// applies while it is (passive.go). A new one replaces it and the old goes to
// its owner's discard. One Stadium per turn per player, and never one with the
// same name as the Stadium in play.

// StadiumSpot is the Stadium in play, as a target (Field Blower).
var StadiumSpot = Spot{Player: -1, Pos: -2}

func (g *Game) stadiumGate(p int, c Card) error {
	if g.S.Players[p].StadiumUsed {
		return errors.New("you already played a Stadium this turn")
	}
	if st := g.S.Stadium; st != nil && st.Name == c.Name {
		return fmt.Errorf("%s is already in play", c.Name)
	}
	return nil
}

func (g *Game) playStadium(p int, c Card) {
	g.discardStadium()
	g.S.Stadium = &c
	g.S.StadiumOwner = p
	g.S.Players[p].StadiumUsed = true
	g.narrate("%s played %s.", side(p), c.Name)
}

// discardStadium puts the Stadium in play, if any, in its owner's discard.
func (g *Game) discardStadium() {
	st := g.S.Stadium
	if st == nil {
		return
	}
	owner := &g.S.Players[g.S.StadiumOwner]
	owner.Discard = append(owner.Discard, *st)
	g.S.Stadium = nil
	g.narrate("%s was discarded.", st.Name)
}

// DiscardToolOrStadium discards the Tool from the targeted Pokémon, or the
// Stadium in play when StadiumSpot is targeted (Field Blower).
type DiscardToolOrStadium struct{}

func (DiscardToolOrStadium) Targets(ctx *EffectContext) []Spot {
	out := DiscardTool{}.Targets(ctx)
	if ctx.g.S.Stadium != nil {
		out = append(out, StadiumSpot)
	}
	return out
}

func (DiscardToolOrStadium) Apply(ctx *EffectContext) {
	if ctx.Target == StadiumSpot {
		ctx.g.discardStadium()
		return
	}
	DiscardTool{}.Apply(ctx)
}
