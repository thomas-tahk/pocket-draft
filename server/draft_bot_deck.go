package main

import (
	"sort"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// The bot's deck is built to be playable, not strong: it keeps to two attack
// types, drafts lines it can finish, and fills in the free under-evolutions
// and universal Trainers the way a human does at deckbuild. The bot is only a
// practice opponent (docs/VISION.md), so these are simple rules, no lookahead.

const (
	botMaxAttackTypes = 2
	botTrainerPicks   = 4 // Trainers the draft takes before preferring Pokémon again
)

var botUniversals = []string{"Professor's Research", "Poké Ball"}

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
		if s := botPickScore(c, have, trainers, byName) - copies[c.Name]; copies[c.Name] < 2 && s > bestScore {
			best, bestScore = c, s
		}
	}
	return best
}

func botPickScore(c draftCard, have typeSet, trainers int, byName map[string][]rawCard) int {
	if c.CardType == "Trainer" {
		if trainers < botTrainerPicks {
			return 6
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
	return 12 - len(line)
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

func (b *botDeckBuilder) addUniversals() {
	for _, name := range botUniversals {
		for _, rc := range b.byName[name] {
			if rc.CardType == "Trainer" && textWorks(rc) {
				b.addCopies([]rawCard{rc}, 2)
				break
			}
		}
	}
}

func (b *botDeckBuilder) addPicks(picks []draftCard, trainers bool, target int) {
	for _, p := range picks {
		if (p.CardType == "Trainer") != trainers {
			continue
		}
		if trainers {
			b.addCopies([]rawCard{p.rawCard}, target)
		} else {
			b.addCopies(lineFrom(p.rawCard, b.types, b.byName), target)
		}
	}
}

func (b *botDeckBuilder) fillFrom(picks []draftCard) {
	b.addPicks(picks, false, 1)
	b.addPicks(picks, true, 1)
	b.addPicks(picks, false, 2)
	b.addPicks(picks, true, 2)
}

// botDeckIDs turns the bot's draft picks into 20 card ids: the universal
// Trainers, then each Pokémon pick with its under-evolutions, then Trainers,
// then second copies. The always-legal fallback is the plain top-up.
func botDeckIDs(picks []draftCard, byName map[string][]rawCard) []string {
	b := &botDeckBuilder{count: map[string]int{}, types: typeSet{}, byName: byName, limit: botMaxAttackTypes}
	b.addUniversals()
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
