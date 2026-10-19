package engine

import "errors"

// Fossils (spec A5): an Item played to an empty Bench slot as a Basic with
// its printed HP that can't retreat. On your turn you may discard it from
// play; that scores nothing. It is never a Basic for the opening hand, setup
// or deck building (Card.IsBasicPokemon is false). It evolves like a Basic
// named after it (Kabuto from Helix Fossil).

// fossilTargets is each of p's empty Bench slots.
func fossilTargets(g *Game, p int) []Spot {
	var out []Spot
	for slot, mon := range g.S.Players[p].Bench {
		if mon == nil {
			out = append(out, Spot{p, slot + 1})
		}
	}
	return out
}

func (g *Game) playFossil(p int, c Card, at Spot) {
	g.S.Players[p].Bench[at.Pos-1] = newInPlay(c, g.S.Turn)
	g.narrate("%s put %s on the Bench.", side(p), c.Name)
}

// DiscardFossil discards a Fossil from play on its owner's turn. Pos is 0 for
// the Active, n for bench slot n-1.
type DiscardFossil struct {
	Player int
	Pos    int
}

func (DiscardFossil) event() {}

// ActDiscardFossil is the Action kind for a Fossil discarded from play.
const ActDiscardFossil ActionKind = "discard_fossil"

func listDiscardFossil(g *Game, p int) []DiscardFossil {
	var out []DiscardFossil
	for _, pos := range targets() {
		out = append(out, DiscardFossil{Player: p, Pos: pos})
	}
	return out
}

func (g *Game) checkDiscardFossil(e DiscardFossil) error {
	if err := g.onTurn(e.Player); err != nil {
		return err
	}
	if mon := targetMon(&g.S.Players[e.Player], e.Pos); mon == nil || !mon.Card.Fossil {
		return errors.New("only a Fossil can be discarded from play")
	}
	return nil
}

// applyDiscardFossil discards the Fossil (and any Tool on it). An Active
// Fossil leaves the spot empty: its owner promotes, or loses with no Pokémon.
func (g *Game) applyDiscardFossil(e DiscardFossil) {
	pl := &g.S.Players[e.Player]
	mon := targetMon(pl, e.Pos)
	pl.Discard = append(pl.Discard, mon.Card)
	if mon.Tool != nil {
		pl.Discard = append(pl.Discard, *mon.Tool)
	}
	g.narrate("%s discarded %s from play.", side(e.Player), mon.Card.Name)
	if e.Pos > 0 {
		pl.Bench[e.Pos-1] = nil
		return
	}
	pl.Active = nil
	if !g.checkResult() {
		g.promptPromotion()
	}
}
