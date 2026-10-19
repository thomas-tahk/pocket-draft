package engine

import (
	"errors"
	"fmt"
)

// Trainer cards (spec A1–A3). A Trainer's text is a list of EffectOps on
// Card.Trainer — the same verbs attacks use, run against an EffectContext
// whose acting player is the one who played the card.
//
// HOW TO ADD A TRAINER TEXT (one wording):
//  1. Write (or reuse) an EffectOp in effects.go. Apply does the effect.
//  2. Optionally give it either of these methods; PlayTrainer asks every op:
//     - CanPlay(ctx *EffectContext) error — why the card can't be played right
//       now (nil when it can). Nothing is played when any op refuses.
//     - Targets(ctx *EffectContext) []Spot — the spots the card is dropped on
//       (Potion: your damaged Pokémon). Without it the card has no target and
//       is dropped on the board. Apply reads the chosen spot from ctx.Target.
//  3. Map the printed wording to the ops in trainerEffects (server/carddata.go).
//     Every printing with that wording then plays it.
//  4. Test it on a real card from public/data/cards.json.
//
// The played Item or Supporter is in the discard pile before its ops run.
// A Tool is attached instead (tool.go); its text is Card.Passive.

// Gated is an optional EffectOp method for Trainers: why it can't be played now.
type Gated interface {
	CanPlay(ctx *EffectContext) error
}

// Targeted is an optional EffectOp method for Trainers: where it may be played.
type Targeted interface {
	Targets(ctx *EffectContext) []Spot
}

// PlayTrainer plays a Trainer from hand. Target is the spot it is dropped on
// (a Tool: the Pokémon it attaches to), or NoSpot for a Trainer with no target.
type PlayTrainer struct {
	Player int
	CardID string
	Target Spot
}

func (PlayTrainer) event() {}

// ActPlayTrainer is the Action kind for a played Trainer.
const ActPlayTrainer ActionKind = "play_trainer"

func (g *Game) trainerCtx(p int, c Card, target Spot) *EffectContext {
	return &EffectContext{g: g, attacker: p, Card: c, Target: target}
}

// listPlayTrainer offers each Trainer in hand on each spot it targets, or with
// no target; the check filters them.
func listPlayTrainer(g *Game, p int) []PlayTrainer {
	var out []PlayTrainer
	for _, c := range g.S.Players[p].Hand {
		if !c.IsTrainer() {
			continue
		}
		for _, t := range trainerTargets(g.trainerCtx(p, c, NoSpot)) {
			out = append(out, PlayTrainer{Player: p, CardID: c.ID, Target: t})
		}
	}
	return out
}

// trainerTargets is the union of the card's ops' targets, or just NoSpot when
// no op takes a target.
func trainerTargets(ctx *EffectContext) []Spot {
	if ctx.Card.Kind == KindTool {
		return toolTargets(ctx.g, ctx.attacker)
	}
	var out []Spot
	targeted := false
	for _, op := range ctx.Card.Trainer {
		if t, ok := op.(Targeted); ok {
			targeted = true
			out = append(out, t.Targets(ctx)...)
		}
	}
	if !targeted {
		return []Spot{NoSpot}
	}
	return out
}

// trainerGate is every PlayTrainer check that doesn't depend on the target.
func (g *Game) trainerGate(p int, cardID string) (Card, error) {
	if err := g.onTurn(p); err != nil {
		return Card{}, err
	}
	pl := &g.S.Players[p]
	idx := indexInHand(pl.Hand, cardID)
	if idx < 0 {
		return Card{}, errors.New("card is not in hand")
	}
	c := pl.Hand[idx]
	switch {
	case !c.IsTrainer() || c.Kind == KindStadium:
		return Card{}, errors.New("only Items, Supporters and Tools can be played this way")
	case !c.TextBuilt():
		return Card{}, fmt.Errorf("%s's text isn't built yet", c.Name)
	case c.Kind == KindSupporter && pl.SupporterUsed:
		return Card{}, errors.New("you already played a Supporter this turn")
	}
	ctx := g.trainerCtx(p, c, NoSpot)
	for _, op := range c.Trainer {
		if gate, ok := op.(Gated); ok {
			if err := gate.CanPlay(ctx); err != nil {
				return Card{}, err
			}
		}
	}
	return c, nil
}

func (g *Game) checkPlayTrainer(e PlayTrainer) error {
	c, err := g.trainerGate(e.Player, e.CardID)
	if err != nil {
		return err
	}
	for _, t := range trainerTargets(g.trainerCtx(e.Player, c, NoSpot)) {
		if t == e.Target {
			return nil
		}
	}
	return fmt.Errorf("%s can't be played there", c.Name)
}

// applyPlayTrainer attaches a Tool, or discards the card and runs its text.
func (g *Game) applyPlayTrainer(e PlayTrainer) {
	pl := &g.S.Players[e.Player]
	idx := indexInHand(pl.Hand, e.CardID)
	c := pl.Hand[idx]
	pl.Hand = removeAt(pl.Hand, idx)
	if c.Kind == KindTool {
		g.attachTool(e.Player, c, e.Target)
		return
	}
	pl.Discard = append(pl.Discard, c)
	if c.Kind == KindSupporter {
		pl.SupporterUsed = true
	}
	g.narrate("%s played %s.", side(e.Player), c.Name)
	ctx := g.trainerCtx(e.Player, c, e.Target)
	for _, op := range c.Trainer {
		op.Apply(ctx)
	}
}

// whyNoTrainer: the target-independent reason, else that nothing can be
// targeted.
func (g *Game) whyNoTrainer(p int, cardID string) error {
	c, err := g.trainerGate(p, cardID)
	if err != nil {
		return err
	}
	return fmt.Errorf("%s has nothing to target", c.Name)
}

// errHandFull is why a card that puts cards into your hand can't be played
// with 10 in hand (spec T3). The card counts: it is still in hand.
var errHandFull = errors.New("your hand is full")

func handFull(ctx *EffectContext) error {
	if len(ctx.Player().Hand) >= handMax {
		return errHandFull
	}
	return nil
}
