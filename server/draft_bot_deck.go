package main

import (
	"sort"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// The bot's deck is built to be playable, not strong: it keeps to two attack
// types, drafts lines it can finish, and fills in the free under-evolutions
// the way a human does at deckbuild. The bot is only a
// practice opponent (docs/VISION.md), so these are simple rules, no lookahead.

const (
	botMaxAttackTypes = 2
	botTrainerPicks   = 4 // Trainers the draft takes before preferring Pokémon again
	botSecondCopy     = 4 // score bonus for a card the draft already holds one copy of
)

type typeSet map[engine.Energy]bool

// costTypes is the typed (non-Colorless) Energy the card's attacks ask for.
func costTypes(rc rawCard) typeSet {
	types := typeSet{}
	for _, a := range rc.Attacks {
		for _, e := range parseCost(a.Cost) {
			if e != engine.Colorless {
				types[e] = true
			}
		}
	}
	return types
}

// fits reports whether adding the card's types keeps the deck within the cap.
func (have typeSet) fits(rc rawCard) bool {
	return have.fitsWithin(rc, botMaxAttackTypes)
}

func (have typeSet) fitsWithin(rc rawCard, limit int) bool {
	n := len(have)
	for e := range costTypes(rc) {
		if !have[e] {
			n++
		}
	}
	return n <= limit
}

// has reports whether every typed Energy the card asks for is already drafted.
func (have typeSet) has(rc rawCard) bool {
	for e := range costTypes(rc) {
		if !have[e] {
			return false
		}
	}
	return true
}

func (have typeSet) add(rc rawCard) {
	for e := range costTypes(rc) {
		have[e] = true
	}
}

// botPick takes the best card of the offer: Pokémon that keep the deck to two
// attack types, the shortest evolution lines first, and a few Trainers. A card
// that would add a third type scores 0 and is dropped when the deck is built.
// Ties go to the earliest card in the offer.
func botPick(offer, picks []draftCard, byName map[string][]rawCard) draftCard {
	have, trainers, copies := typeSet{}, 0, map[string]int{}
	for _, p := range picks {
		copies[p.Name]++
		if p.CardType == "Trainer" {
			trainers++
		} else {
			have.add(p.rawCard)
		}
	}
	best, bestScore := offer[0], -1
	for _, c := range offer {
		if copies[c.Name] >= 2 {
			continue
		}
		s := botPickScore(c, have, trainers, byName)
		if s > 0 && copies[c.Name] == 1 {
			s += botSecondCopy
		}
		if s > bestScore {
			best, bestScore = c, s
		}
	}
	return best
}

func botPickScore(c draftCard, have typeSet, trainers int, byName map[string][]rawCard) int {
	if c.CardType == "Trainer" {
		if trainers < botTrainerPicks {
			return 11
		}
		return 2
	}
	if !have.fits(c.rawCard) {
		return 0
	}
	line := lineFrom(c.rawCard, have, byName)
	if line == nil {
		return 0 // an evolution whose earlier stage can't be had
	}
	score := 12 - len(line)
	if c.IsEx {
		score++
	}
	if !have.has(c.rawCard) && len(have) > 0 {
		score-- // prefer staying inside the types already drafted
	}
	return score
}

// nameIndex maps a card name to every printing, lowest id first.
func nameIndex() map[string][]rawCard {
	ids := make([]string, 0, len(rawByID))
	for id := range rawByID {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	byName := map[string][]rawCard{}
	for _, id := range ids {
		byName[rawByID[id].Name] = append(byName[rawByID[id].Name], rawByID[id])
	}
	return byName
}

// lineFrom is the card's chain from its Basic up to itself, using for each
// earlier stage the first working printing whose types fit. Nil when a stage
// has no such printing.
func lineFrom(rc rawCard, have typeSet, byName map[string][]rawCard) []rawCard {
	line := []rawCard{rc}
	for rc.EvolvesFrom != "" {
		prev, ok := firstFitting(byName[rc.EvolvesFrom], have)
		if !ok {
			return nil
		}
		line = append([]rawCard{prev}, line...)
		rc = prev
	}
	return line
}

func firstFitting(printings []rawCard, have typeSet) (rawCard, bool) {
	for _, rc := range printings {
		if textWorks(rc) && have.fits(rc) {
			return rc, true
		}
	}
	return rawCard{}, false
}

// botDeckBuilder accumulates a legal 20-card id list.
type botDeckBuilder struct {
	ids    []string
	count  map[string]int // copies by card name
	types  typeSet
	byName map[string][]rawCard
	limit  int // attack types allowed; raised only when the picks can't fill 20
}

// addCopies raises each card of the line to target copies, or adds nothing if
// that would overflow 20 cards or add a third attack type.
func (b *botDeckBuilder) addCopies(line []rawCard, target int) {
	if line == nil {
		return
	}
	need, trial := 0, typeSet{}
	for e := range b.types {
		trial[e] = true
	}
	for _, rc := range line {
		if !trial.fitsWithin(rc, b.limit) {
			return
		}
		trial.add(rc)
		need += max(0, target-b.count[rc.Name])
	}
	if len(b.ids)+need > draftDeckSize {
		return
	}
	b.types = trial
	for _, rc := range line {
		for b.count[rc.Name] < target {
			b.ids = append(b.ids, rc.ID)
			b.count[rc.Name]++
		}
	}
}

// pickGroup is one distinct picked card and how many times it was picked.
type pickGroup struct {
	card  draftCard
	times int
}

// groupPicks collapses the picks by name, keeping first-pick order.
func groupPicks(picks []draftCard) []pickGroup {
	var groups []pickGroup
	index := map[string]int{}
	for _, p := range picks {
		if i, ok := index[p.Name]; ok {
			groups[i].times++
			continue
		}
		index[p.Name] = len(groups)
		groups = append(groups, pickGroup{p, 1})
	}
	return groups
}

// addGroups raises each group's card (with its line, for Pokémon) to target
// copies where it fits.
func (b *botDeckBuilder) addGroups(groups []pickGroup, target int) {
	for _, g := range groups {
		if g.card.CardType == "Trainer" {
			b.addCopies([]rawCard{g.card.rawCard}, target)
		} else {
			b.addCopies(lineFrom(g.card.rawCard, b.types, b.byName), target)
		}
	}
}

// rankGroups splits the picks into Pokémon (picked twice first, then the
// cheapest line) and Trainers (pick order, at most botTrainerPicks).
func (b *botDeckBuilder) rankGroups(picks []draftCard) (pokemon, trainers []pickGroup) {
	for _, g := range groupPicks(picks) {
		if g.card.CardType == "Trainer" {
			trainers = append(trainers, g)
		} else {
			pokemon = append(pokemon, g)
		}
	}
	cost := func(g pickGroup) int { return len(lineFrom(g.card.rawCard, b.types, b.byName)) }
	sort.SliceStable(pokemon, func(i, j int) bool {
		if pokemon[i].times != pokemon[j].times {
			return pokemon[i].times > pokemon[j].times
		}
		return cost(pokemon[i]) < cost(pokemon[j])
	})
	return pokemon, trainers[:min(len(trainers), botTrainerPicks)]
}

// fillFrom builds the deck around pairs: Trainers picked twice at two, the
// other Trainers, then Pokémon lines two copies at a time (cheapest first).
// Singles and second Trainer copies only fill what is left.
func (b *botDeckBuilder) fillFrom(picks []draftCard) {
	pokemon, trainers := b.rankGroups(picks)
	var pairs []pickGroup
	for _, g := range trainers {
		if g.times >= 2 {
			pairs = append(pairs, g)
		}
	}
	b.addGroups(pairs, 2)
	b.addGroups(trainers, 1)
	b.addGroups(pokemon, 2)
	b.addGroups(pokemon, 1)
	b.addGroups(trainers, 2)
}

// botDeckIDs turns the bot's draft picks into 20 card ids: each Pokémon pick with its under-evolutions, then Trainers,
// then second copies. The always-legal fallback is the plain top-up.
func botDeckIDs(picks []draftCard, byName map[string][]rawCard) []string {
	b := &botDeckBuilder{count: map[string]int{}, types: typeSet{}, byName: byName, limit: botMaxAttackTypes}
	b.fillFrom(picks)
	if len(b.ids) < draftDeckSize {
		b.limit = int(engine.Metal) + 1
		b.fillFrom(picks)
	}
	if len(b.ids) != draftDeckSize {
		return topUpToDeckSize(picks)
	}
	return b.ids
}
