package main

import (
	"testing"
)

// deckStats is what one drafted bot deck looks like.
type deckStats struct {
	size, types, pairs, trainers int
	foreign                      []string
}

// lineNames is every picked name plus all its under-evolutions.
func lineNames(picked map[string]bool) map[string]bool {
	allowed := map[string]bool{}
	for n := range picked {
		allowed[n] = true
	}
	byName := nameIndex()
	for grew := true; grew; {
		grew = false
		for n := range allowed {
			for _, rc := range byName[n] {
				if rc.EvolvesFrom != "" && !allowed[rc.EvolvesFrom] {
					allowed[rc.EvolvesFrom], grew = true, true
				}
			}
		}
	}
	return allowed
}

func statsForSeed(t *testing.T, seed uint64) deckStats {
	t.Helper()
	picks, ids, err := draftBotIDs(seed)
	if err != nil {
		t.Fatalf("seed %d: %v", seed, err)
	}
	picked := map[string]bool{}
	for _, id := range picks {
		picked[rawByID[id].Name] = true
	}
	allowed := lineNames(picked)
	s := deckStats{size: len(ids)}
	copies := map[string]int{}
	types := typeSet{}
	isTrainer := map[string]bool{}
	for _, id := range ids {
		rc := rawByID[id]
		copies[rc.Name]++
		isTrainer[rc.Name] = rc.CardType == "Trainer"
		types.add(rc)
		if !allowed[rc.Name] {
			s.foreign = append(s.foreign, rc.Name)
		}
	}
	for n, c := range copies {
		if picked[n] && c >= 2 {
			s.pairs++
		}
		if picked[n] && isTrainer[n] {
			s.trainers++
		}
	}
	s.types = len(types)
	return s
}

func TestBotPickerBuildsARealDeck(t *testing.T) {
	useRealCards(t)
	le2, withPair, withTrainers := 0, 0, 0
	for seed := uint64(1); seed <= botSeeds; seed++ {
		s := statsForSeed(t, seed)
		if s.size != 20 {
			t.Errorf("seed %d: deck has %d cards, want 20", seed, s.size)
		}
		if s.types > 3 {
			t.Errorf("seed %d: %d attack types, want at most 3", seed, s.types)
		}
		if len(s.foreign) > 0 {
			t.Errorf("seed %d: cards not from picks or their lines: %v", seed, s.foreign)
		}
		if s.types <= 2 {
			le2++
		}
		if s.pairs > 0 {
			withPair++
		}
		if s.trainers >= 2 {
			withTrainers++
		}
	}
	t.Logf("<=2 types: %d/%d; with a picked 2-of: %d/%d; >=2 picked Trainers: %d/%d", le2, botSeeds, withPair, botSeeds, withTrainers, botSeeds)
	// Thresholds are 90%: "most", with room for an odd offer sequence.
	if withPair*10 < botSeeds*9 {
		t.Errorf("only %d/%d decks have a picked 2-of", withPair, botSeeds)
	}
	if withTrainers*10 < botSeeds*9 {
		t.Errorf("only %d/%d decks have >=2 picked Trainers", withTrainers, botSeeds)
	}
	if le2*10 < botSeeds*8 {
		t.Errorf("only %d/%d decks are <=2 types", le2, botSeeds)
	}
}
