package main

import "testing"

// Eevee ex becomes draftable (owner 2026-10-25, issue #128).

func rawDraftCard(t *testing.T, id string) draftCard {
	t.Helper()
	rc, ok := rawByID[id]
	if !ok {
		t.Fatalf("no card %s", id)
	}
	dc, ok := enrichForDraft(rc)
	if !ok {
		t.Fatalf("%s has no draft rarity", id)
	}
	return dc
}

func TestEeveeExIsInTheDraftPool(t *testing.T) {
	// Arrange
	useRealCards(t)

	// Act
	pool := toDraftablePool()

	// Assert
	n := 0
	for _, c := range pool {
		if c.Name == eeveeEx {
			n++
		}
	}
	if n == 0 {
		t.Fatal("Eevee ex is not in the draft pool")
	}
}

// eeveeExOffers counts how often Eevee ex shows up in offers over many seeds
// with the given picks held.
func eeveeExOffers(t *testing.T, picks []draftCard) int {
	t.Helper()
	byTier := indexByTier(toDraftablePool())
	n := 0
	for seed := uint64(1); seed <= 4000; seed++ {
		rng := newDraftRng(seed)
		offer, err := generateOffer(byTier, picks, rng.Float64)
		if err != nil {
			t.Fatal(err)
		}
		for _, c := range offer {
			if c.Name == eeveeEx {
				n++
			}
		}
	}
	return n
}

func TestPickingAnEeveelutionRaisesEeveeExOdds(t *testing.T) {
	// Arrange: Sylveon ex vs. Cresselia ex, both Psychic ex, so the type lean is
	// the same and only the Eeveelution differs.
	useRealCards(t)
	sylveon := rawDraftCard(t, "A3b-034")
	other := rawDraftCard(t, "P-A-037")
	if !isEeveelution(sylveon) || isEeveelution(other) || sylveon.CardType != other.CardType {
		t.Fatalf("bad fixtures: %s %s / %s %s", sylveon.Name, sylveon.CardType, other.Name, other.CardType)
	}

	// Act
	with := eeveeExOffers(t, []draftCard{sylveon})
	without := eeveeExOffers(t, []draftCard{other})

	// Assert
	t.Logf("Eevee ex offered %d times with Sylveon ex picked, %d with Cresselia ex (4000 offers each)", with, without)
	if without == 0 || float64(with) < 1.8*float64(without) {
		t.Errorf("Eevee ex offered %d times with an Eeveelution picked vs %d without; want a clear rise", with, without)
	}
}

func TestEeveeBoostDoesNotStack(t *testing.T) {
	c := draftCard{rawCard: rawCard{Name: eeveeEx}}
	if got := eeveeBoost(c, true, true); got != 1 {
		t.Errorf("boost on an already boosted card = %v, want 1", got)
	}
	if got := eeveeBoost(draftCard{rawCard: rawCard{Name: "Eevee"}}, true, false); got != 1 {
		t.Errorf("plain Eevee boosted by %v", got)
	}
}
