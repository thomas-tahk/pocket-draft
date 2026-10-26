package main

import "testing"

// leanCounts is what a run of seeded offers contained.
type leanCounts struct {
	offers      int
	slots       int
	water       int // Water Pokémon
	otherTyped  int // Pokémon of a type that is neither Water nor Colorless
	support     int // Water-support Trainers/Items
	otherTypeBy map[string]int
}

func countOffers(byTier map[tier][]draftCard, picks []draftCard, support map[string]bool) leanCounts {
	lc := leanCounts{otherTypeBy: map[string]int{}}
	for seed := uint64(1); seed <= 20; seed++ {
		rng := newDraftRng(seed)
		for i := 0; i < 250; i++ {
			offer, err := generateOffer(byTier, picks, rng.Float64)
			if err != nil {
				panic(err)
			}
			lc.offers++
			for _, c := range offer {
				lc.slots++
				switch {
				case support[c.ID]:
					lc.support++
				case c.CardType == "Water":
					lc.water++
				case c.CardType != "Trainer" && c.CardType != "Colorless":
					lc.otherTyped++
					lc.otherTypeBy[c.CardType]++
				}
			}
		}
	}
	return lc
}

// pickTypeCards returns n distinct non-ex final-evolution Pokémon of typ whose
// text names no type and that no Trainer in the pool names, so the only thing
// they can do to an offer is lean the type.
func pickTypeCards(pool []draftCard, typ string, n int) []draftCard {
	var out []draftCard
	for _, c := range pool {
		if c.CardType != typ || c.IsEx || c.IsMegaEx || len(namedTypes(c)) > 0 {
			continue
		}
		named := false
		for _, tr := range pool {
			if tr.CardType == "Trainer" && newSynergy([]draftCard{c}).namesPickedIn(tr.TrainerText) {
				named = true
				break
			}
		}
		if !named {
			out = append(out, c)
			if len(out) == n {
				break
			}
		}
	}
	return out
}

// After 3 Water picks, Water Pokémon and Water-support Trainers/Items are each
// clearly more common than before the lean, and nothing is forced: Pokémon of
// every other type still show up.
func TestThreeWaterPicksLeanPokemonAndSupport(t *testing.T) {
	useRealCards(t)
	pool := toDraftablePool()
	byTier := indexByTier(pool)

	// Water support = Trainers whose text names only Water, plus two that name
	// Water Pokémon without saying "Water" (Marlon, Lana; hand-judged).
	support := map[string]bool{}
	for _, c := range pool {
		if c.CardType != "Trainer" {
			continue
		}
		named := typesNamedIn(effectText(c))
		if (len(named) == 1 && named["Water"]) || c.Name == "Marlon" || c.Name == "Lana" {
			support[c.ID] = true
		}
	}
	if len(support) < 5 {
		t.Fatalf("only %d Water-support Trainers in the pool", len(support))
	}

	water := pickTypeCards(pool, "Water", 3)
	colorless := pickTypeCards(pool, "Colorless", 3)
	if len(water) != 3 || len(colorless) != 3 {
		t.Fatalf("need 3 Water and 3 Colorless picks, got %d and %d", len(water), len(colorless))
	}
	base := countOffers(byTier, colorless, support) // no type lean at all
	lean := countOffers(byTier, water, support)

	pokeRatio := float64(lean.water) / float64(base.water)
	supRatio := float64(lean.support) / float64(base.support)
	t.Logf("%d Water-support cards in pool; %d offers each", len(support), base.offers)
	t.Logf("Water Pokémon slots: baseline %d (%.1f%%), after 3 Water picks %d (%.1f%%), x%.2f",
		base.water, 100*float64(base.water)/float64(base.slots), lean.water, 100*float64(lean.water)/float64(lean.slots), pokeRatio)
	t.Logf("Water-support slots: baseline %d (%.2f%%), after 3 Water picks %d (%.2f%%), x%.2f",
		base.support, 100*float64(base.support)/float64(base.slots), lean.support, 100*float64(lean.support)/float64(lean.slots), supRatio)
	t.Logf("other-typed Pokémon slots after the lean: %d (%.1f%%), by type %v",
		lean.otherTyped, 100*float64(lean.otherTyped)/float64(lean.slots), lean.otherTypeBy)

	if pokeRatio < leanMinPokemonRatio {
		t.Errorf("Water Pokémon only x%.2f more common after 3 Water picks, want >= x%.1f", pokeRatio, leanMinPokemonRatio)
	}
	if supRatio < leanMinSupportRatio {
		t.Errorf("Water-support cards only x%.2f more common after 3 Water picks, want >= x%.1f", supRatio, leanMinSupportRatio)
	}
	// Raised odds, never guarantees: other types stay in play.
	if share := float64(lean.otherTyped) / float64(lean.slots); share < 0.15 {
		t.Errorf("other-typed Pokémon are only %.1f%% of slots after the lean, want >= 15%%", 100*share)
	}
	for _, typ := range []string{"Fire", "Grass", "Lightning", "Psychic", "Fighting", "Darkness", "Metal"} {
		if lean.otherTypeBy[typ] == 0 {
			t.Errorf("no %s Pokémon offered in %d offers after the lean", typ, lean.offers)
		}
	}
}

// What "clearly" means: the lean must at least this many times the baseline.
const (
	leanMinPokemonRatio = 3.0
	leanMinSupportRatio = 3.0
)

// A Trainer that works against a type must not get that type's lean.
func TestOpposingTrainerGetsNoLean(t *testing.T) {
	useRealCards(t)
	var squirtBottle draftCard
	for _, c := range toDraftablePool() {
		if c.Name == "Squirt Bottle" {
			squirtBottle = c
		}
	}
	if squirtBottle.ID == "" {
		t.Skip("Squirt Bottle is not in the draft pool")
	}
	if w := weightOf(squirtBottle, typeTally{"Fire": 3}, synergy{}); w != 1 {
		t.Errorf("Squirt Bottle after 3 Fire picks = %v, want 1", w)
	}
}

// A table Trainer (text names Pokémon, not a type) leans like a text-named one.
func TestTableTrainerLeansWithItsType(t *testing.T) {
	marlon := trainerText("Marlon", "Heal 70 damage from 1 of your Carracosta or Jellicent.")
	if w := weightOf(marlon, typeTally{"Water": 3}, synergy{}); w != 1+typeSupportCoef*3 {
		t.Errorf("Marlon after 3 Water picks = %v, want %v", w, 1+typeSupportCoef*3)
	}
	if w := weightOf(marlon, typeTally{"Fire": 3}, synergy{}); w != 1 {
		t.Errorf("Marlon after 3 Fire picks = %v, want 1", w)
	}
}
