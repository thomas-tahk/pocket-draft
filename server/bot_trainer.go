package main

import (
	"strings"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// trainerRank is the bot's judgment on a Trainer, read off its printed text:
// a few plain checks so it doesn't waste a card. Everything else it plays.
//   - heals: only a damaged target, the most damaged first
//   - switching its own Active, or cutting the Retreat Cost: only a stuck Active
//   - "+damage this turn": only when the Active can attack now
//   - shuffling its hand away to draw: only from a small hand
func trainerRank(g *engine.Game, e engine.PlayTrainer) (int, bool) {
	text := rawByID[e.CardID].TrainerText
	pl := &g.S.Players[e.Player]
	switch {
	case strings.Contains(text, "Heal"):
		return healRank(g, e)
	case strings.HasPrefix(text, "Switch your Active"), strings.Contains(text, "Retreat Cost of your Active"):
		return rankTrainer, missingEnergy(g, e.Player, pl.Active) == unreachable
	case strings.Contains(text, "During this turn, attacks used by"):
		return rankTrainer, missingEnergy(g, e.Player, pl.Active) == 0
	case strings.HasPrefix(text, "Shuffle your hand"), strings.HasPrefix(text, "Each player shuffles"):
		return rankTrainer, len(pl.Hand) <= smallHand
	}
	return rankTrainer, true
}

// smallHand counts the Trainer being played.
const smallHand = 3

// healRank heals the target with the most damage on it. A heal with no target
// waits until one of the bot's Pokémon is damaged.
func healRank(g *engine.Game, e engine.PlayTrainer) (int, bool) {
	if e.Target.Player == e.Player {
		dmg := target(g, e.Player, e.Target.Pos).Damage
		return rankTrainer + min(dmg/10, 9), dmg > 0
	}
	return rankTrainer, mostDamage(g, e.Player) > 0
}

func mostDamage(g *engine.Game, player int) int {
	pl := &g.S.Players[player]
	most := 0
	for _, m := range append([]*engine.InPlay{pl.Active}, pl.Bench[:]...) {
		if m != nil {
			most = max(most, m.Damage)
		}
	}
	return most
}
