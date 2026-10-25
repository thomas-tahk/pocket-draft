package main

import (
	"sort"
	"testing"
)

func pokemon(name, typ, evolvesFrom string) draftCard {
	return draftCard{rawCard: rawCard{ID: name + typ, Name: name, CardType: typ, EvolvesFrom: evolvesFrom}}
}

func trainerText(name, text string) draftCard {
	return draftCard{rawCard: rawCard{ID: name + text, Name: name, CardType: "Trainer", TrainerText: text}}
}

func TestSynergyEvolutionLineBothDirections(t *testing.T) {
	picks := []draftCard{pokemon("Charmeleon", "Fire", "Charmander")}
	syn := newSynergy(picks)
	tally := tallyTypes(picks) // Fire 1 -> base weight 2
	cases := []struct {
		card draftCard
		want float64
	}{
		{pokemon("Charizard", "Fire", "Charmeleon"), 2 * synergyMult},
		{pokemon("Charmander", "Fire", ""), 2 * synergyMult},
		{pokemon("Charmander", "Colorless", ""), 1 * synergyMult},
		{pokemon("Squirtle", "Fire", ""), 2},
		{pokemon("Charizard", "Fire", "Charmeleon X"), 2},
	}
	for _, c := range cases {
		if got := weightOf(c.card, tally, syn); got != c.want {
			t.Errorf("weightOf(%s from %q) = %v, want %v", c.card.Name, c.card.EvolvesFrom, got, c.want)
		}
	}
}

func TestSynergyNamedTrainerOnly(t *testing.T) {
	syn := newSynergy([]draftCard{pokemon("Onix", "Fighting", "")})
	tally := typeTally{}
	brock := trainerText("Brock", "Attach a [F] Energy to Golem or Onix.")
	misty := trainerText("Misty", "Flip a coin until you get tails. Attach Water Energy.")
	longer := trainerText("Other", "Attach to Onixia.")
	if w := weightOf(brock, tally, syn); w != synergyMult {
		t.Errorf("Brock naming Onix = %v, want %v", w, synergyMult)
	}
	if w := weightOf(misty, tally, syn); w != 1 {
		t.Errorf("Misty = %v, want 1", w)
	}
	if w := weightOf(longer, tally, syn); w != 1 {
		t.Errorf("text with only a longer word = %v, want 1", w)
	}
}

func TestSynergyExPickNamedByBaseName(t *testing.T) {
	syn := newSynergy([]draftCard{pokemon("Mewtwo ex", "Psychic", "")})
	if w := weightOf(trainerText("X", "Heal your Mewtwo."), typeTally{}, syn); w != synergyMult {
		t.Errorf("trainer naming base name = %v, want %v", w, synergyMult)
	}
}

func TestSynergyNeverStacks(t *testing.T) {
	picks := []draftCard{pokemon("Charmeleon", "Fire", "Charmander"), pokemon("Charizard", "Fire", "Charmeleon")}
	syn := newSynergy(picks)
	// Charmeleon is both the parent of a pick and the child of... itself picked.
	w := weightOf(pokemon("Charmeleon", "Fire", "Charmander"), tallyTypes(picks), syn)
	if w != 3*synergyMult {
		t.Errorf("both directions = %v, want a single x%v on base 3", w, synergyMult)
	}
}

// poolWithEvolutions is toDraftablePool minus the final-evolutions-only filter.
// The shipped pool drops every Pokémon something evolves into by name, so a
// pick and a candidate can never be one stage apart there (see
// TestShippedPoolHasNoEvolutionPairs); this is the only way to exercise the
// evolution rule on real card data.
func poolWithEvolutions() []draftCard {
	enriched := []draftCard{}
	for _, rc := range rawByID {
		if !isDraftCandidate(rc) {
			continue
		}
		if dc, ok := enrichForDraft(rc); ok {
			enriched = append(enriched, dc)
		}
	}
	sort.Slice(enriched, func(i, j int) bool { return enriched[i].ID < enriched[j].ID })
	pool := keepWorkingText(functionalDedupe(dedupePromoRayquaza(enriched)))
	sort.Slice(pool, func(i, j int) bool { return pool[i].ID < pool[j].ID })
	return pool
}

// Documents why the rule is quiet in the shipped draft: no pick is ever one
// stage from a candidate by name, because the pool is final evolutions only.
func TestShippedPoolHasNoEvolutionPairs(t *testing.T) {
	useRealCards(t)
	names := map[string]bool{}
	pool := toDraftablePool()
	for _, c := range pool {
		names[c.Name] = true
	}
	for _, c := range pool {
		if c.EvolvesFrom != "" && names[c.EvolvesFrom] {
			t.Errorf("%s evolves from %s and both are in the shipped pool", c.Name, c.EvolvesFrom)
		}
	}
}

// Real Trainers that name a pool Pokémon get offered more once you pick it.
func TestOffersFavourTrainerNamingPick(t *testing.T) {
	useRealCards(t)
	pool := toDraftablePool()
	byTier := indexByTier(pool)
	var pick draftCard
	var named map[string]bool
	for _, c := range pool {
		if c.CardType == "Trainer" || c.CardType == "Colorless" {
			continue
		}
		syn := newSynergy([]draftCard{c})
		n := map[string]bool{}
		for _, tr := range pool {
			if tr.CardType == "Trainer" && syn.matches(tr) {
				n[tr.ID] = true
			}
		}
		if len(n) > len(named) {
			pick, named = c, n
		}
	}
	if len(named) == 0 {
		t.Fatal("no Trainer in the pool names a pool Pokémon")
	}
	var control draftCard
	for _, c := range pool {
		if c.CardType == pick.CardType && c.Name != pick.Name && !newSynergy([]draftCard{c}).matchesAny(pool, named) {
			control = c
			break
		}
	}
	with := countIDsOffered(byTier, []draftCard{pick}, named)
	without := countIDsOffered(byTier, []draftCard{control}, named)
	t.Logf("pick %s, %d naming Trainers: with %d, control (%s) %d", pick.Name, len(named), with, control.Name, without)
	if without == 0 || float64(with) < 1.5*float64(without) {
		t.Errorf("naming Trainers offered %d with the pick vs %d without, want >= 1.5x", with, without)
	}
}

func (s synergy) matchesAny(pool []draftCard, ids map[string]bool) bool {
	for _, c := range pool {
		if ids[c.ID] && s.matches(c) {
			return true
		}
	}
	return false
}

func countIDsOffered(byTier map[tier][]draftCard, picks []draftCard, ids map[string]bool) int {
	n := 0
	for seed := uint64(1); seed <= 20; seed++ {
		rng := newDraftRng(seed)
		for i := 0; i < 100; i++ {
			offer, err := generateOffer(byTier, picks, rng.Float64)
			if err != nil {
				panic(err)
			}
			for _, c := range offer {
				if ids[c.ID] {
					n++
				}
			}
		}
	}
	return n
}

// Picks the Basic with the most evolution printings, then compares how often
// those evolutions are offered against picking an unrelated Basic of the same
// type (same type tally, no line link). Runs on poolWithEvolutions.
func TestOffersFavourEvolutionLineOfPick(t *testing.T) {
	useRealCards(t)
	pool := poolWithEvolutions()
	byTier := indexByTier(pool)
	basic, line := bestBasicWithLine(pool)
	if basic.ID == "" {
		t.Fatal("no Basic with an evolution in the pool")
	}
	var control draftCard
	for _, c := range pool {
		if c.Stage == "Basic" && c.CardType == basic.CardType && c.Name != basic.Name && !line[c.Name] {
			control = c
			break
		}
	}
	with := lineMembersOffered(byTier, []draftCard{basic}, line)
	without := lineMembersOffered(byTier, []draftCard{control}, line)
	t.Logf("%s line %v: with pick %d, control (%s) %d", basic.Name, sortedKeys(line), with, control.Name, without)
	if without == 0 || float64(with) < 1.5*float64(without) {
		t.Errorf("line members offered %d with the pick vs %d without, want >= 1.5x", with, without)
	}
}

func sortedKeys(m map[string]bool) []string {
	out := []string{}
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// bestBasicWithLine returns the non-Colorless Basic whose name has the most
// evolution printings in pool (ties: lowest ID), and the names evolving from it.
func bestBasicWithLine(pool []draftCard) (draftCard, map[string]bool) {
	count := map[string]int{}
	for _, c := range pool {
		if c.EvolvesFrom != "" {
			count[c.EvolvesFrom]++
		}
	}
	var best draftCard
	for _, c := range pool {
		if c.Stage != "Basic" || c.CardType == "Colorless" || c.CardType == "Trainer" {
			continue
		}
		if best.ID == "" || count[c.Name] > count[best.Name] || (count[c.Name] == count[best.Name] && c.ID < best.ID) {
			best = c
		}
	}
	line := map[string]bool{}
	for _, c := range pool {
		if c.EvolvesFrom == best.Name {
			line[c.Name] = true
		}
	}
	return best, line
}

func lineMembersOffered(byTier map[tier][]draftCard, picks []draftCard, line map[string]bool) int {
	n := 0
	for seed := uint64(1); seed <= 20; seed++ {
		rng := newDraftRng(seed)
		for i := 0; i < 100; i++ {
			offer, err := generateOffer(byTier, picks, rng.Float64)
			if err != nil {
				panic(err)
			}
			for _, c := range offer {
				if line[c.Name] {
					n++
				}
			}
		}
	}
	return n
}
