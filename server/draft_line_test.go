package main

import "testing"

// hasWorkingLine reports, independently of the code under test, whether some
// printing of every earlier stage of rc works (a card with no earlier stage
// has a working line).
func hasWorkingLine(rc rawCard) bool {
	if rc.EvolvesFrom == "" {
		return true
	}
	for _, prev := range rawByID {
		if prev.Name == rc.EvolvesFrom && modelable(prev) && hasWorkingLine(prev) {
			return true
		}
	}
	return false
}

// An evolution whose earlier stage cannot be played can never reach the board,
// so the pool and the working set must not offer it.
func TestDraftPoolEvolutionsHaveAWorkingLine(t *testing.T) {
	useRealCards(t)
	for _, c := range toDraftablePool() {
		if !hasWorkingLine(c.rawCard) {
			t.Errorf("pool offers %s (%s), whose earlier stage %q does not work", c.ID, c.Name, c.EvolvesFrom)
		}
	}
	for _, id := range workingIDs() {
		if !hasWorkingLine(rawByID[id]) {
			t.Errorf("working set has %s (%s), whose earlier stage does not work", id, rawByID[id].Name)
		}
	}
}
