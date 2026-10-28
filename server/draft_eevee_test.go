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
	if got := eeveeBoost(c, synergy{eeveelution: true}, true); got != 1 {
		t.Errorf("boost on an already boosted card = %v, want 1", got)
	}
	if got := eeveeBoost(draftCard{rawCard: rawCard{Name: "Eevee"}}, synergy{eeveelution: true}, false); got != 1 {
		t.Errorf("plain Eevee boosted by %v", got)
	}
}

// The other direction (owner 2026-10-28, decision 4): once Eevee ex is picked,
// every Eeveelution gets the same x3, never stacked with other synergy.
func TestPickingEeveeExRaisesEeveelutions(t *testing.T) {
	// Arrange: Sylveon ex vs. Cresselia ex, both Psychic ex, so only the
	// Eeveelution differs; Eevee ex is Colorless, so it leans no type.
	useRealCards(t)
	eevee := rawDraftCard(t, eeveeExID)
	sylveon := rawDraftCard(t, "A3b-034")
	other := rawDraftCard(t, "P-A-037")
	bag := trainerText("Eevee Bag", "Heal each Pokémon that evolves from Eevee.")
	w := func(c draftCard, picks ...draftCard) float64 {
		return weightOf(c, tallyTypes(picks), newSynergy(picks))
	}

	// Act / Assert
	if got, base := w(sylveon, eevee), w(sylveon); got != eeveeMult*base {
		t.Errorf("Sylveon ex after Eevee ex = %v, without = %v, want x%v", got, base, eeveeMult)
	}
	if got, base := w(other, eevee), w(other); got != base {
		t.Errorf("Cresselia ex after Eevee ex = %v, want %v (not an Eeveelution)", got, base)
	}
	if got, base := w(sylveon, eevee, bag), w(sylveon); got != synergyMult*base {
		t.Errorf("Sylveon ex after Eevee ex and Eevee Bag = %v, want %v (one multiplication)", got, synergyMult*base)
	}

	// And in the offers.
	byTier := indexByTier(toDraftablePool())
	count := func(picks ...draftCard) int {
		n := 0
		for seed := uint64(1); seed <= 4000; seed++ {
			rng := newDraftRng(seed)
			offer, err := generateOffer(byTier, picks, rng.Float64)
			if err != nil {
				t.Fatal(err)
			}
			for _, c := range offer {
				if isEeveelution(c) {
					n++
				}
			}
		}
		return n
	}
	colorlessEx := draftCard{}
	for _, c := range toDraftablePool() {
		if c.CardType == "Colorless" && c.Name != eeveeEx && !isEeveelution(c) && len(namedTypes(c)) == 0 && len(nicheGroupsOf(c)) == 0 {
			colorlessEx = c
			break
		}
	}
	with, without := count(eevee), count(colorlessEx)
	t.Logf("Eeveelutions offered %d times after Eevee ex, %d after %s (4000 offers each)", with, without, colorlessEx.Name)
	if without == 0 || float64(with) < 1.8*float64(without) {
		t.Errorf("Eeveelutions offered %d times after Eevee ex vs %d; want a clear rise", with, without)
	}
}
