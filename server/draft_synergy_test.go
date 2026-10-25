package main

import "testing"

func pokemon(name, typ string) draftCard {
	return draftCard{rawCard: rawCard{ID: name + typ, Name: name, CardType: typ}}
}

func trainerText(name, text string) draftCard {
	return draftCard{rawCard: rawCard{ID: name + text, Name: name, CardType: "Trainer", TrainerText: text}}
}

func withAttackEffect(c draftCard, effect string) draftCard {
	c.ID += effect
	c.Attacks = []rawAttack{{Cost: "W", Effect: effect}}
	return c
}

func TestSynergyNamedTrainerOnly(t *testing.T) {
	syn := newSynergy([]draftCard{pokemon("Onix", "Fighting")})
	tally := typeTally{}
	brock := trainerText("Brock", "Attach an Energy to Golem or Onix.")
	misty := trainerText("Misty", "Attach an Energy to your Pokémon.")
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
	syn := newSynergy([]draftCard{pokemon("Mewtwo ex", "Psychic")})
	if w := weightOf(trainerText("X", "Heal your Mewtwo."), typeTally{}, syn); w != synergyMult {
		t.Errorf("trainer naming base name = %v, want %v", w, synergyMult)
	}
}

// Rule 3a: candidate text names a type you picked.
func TestSynergyTextNamesPickedType(t *testing.T) {
	syn := newSynergy([]draftCard{pokemon("Squirtle", "Water")})
	tally := tallyTypes([]draftCard{pokemon("Squirtle", "Water")}) // Water 1
	cases := []struct {
		name string
		card draftCard
		want float64
	}{
		{"symbol in brackets", trainerText("Misty", "Choose 1 of your [W] Pokémon."), synergyMult},
		{"symbol in braces", trainerText("M2", "Attach a {W} Energy."), synergyMult},
		{"Type Pokémon words", trainerText("M3", "Heal your Water Pokémon."), synergyMult},
		{"Type Energy words", trainerText("M4", "Discard a Water Energy."), synergyMult},
		{"other type", trainerText("Blaine", "Choose 1 of your [R] Pokémon."), 1},
		{"colorless never counts", trainerText("C", "Attach a [C] Energy."), 1},
		{"attack effect counts", withAttackEffect(pokemon("Psyduck", "Psychic"), "Take a [W] Energy."), synergyMult}, // Psychic, no Psychic picks: base 1
		{"no text", pokemon("Lapras", "Colorless"), 1},
	}
	for _, c := range cases {
		if got := weightOf(c.card, tally, syn); got != c.want {
			t.Errorf("%s: weight %v, want %v", c.name, got, c.want)
		}
	}
}

func TestSynergyAttackCostIsNotText(t *testing.T) {
	syn := newSynergy([]draftCard{pokemon("Squirtle", "Water")})
	c := pokemon("Psyduck", "Psychic")
	c.ID = "cost-only"
	c.Attacks = []rawAttack{{Cost: "[W]", Name: "[W] Splash", Effect: ""}}
	if w := weightOf(c, typeTally{}, syn); w != 1 {
		t.Errorf("symbol only in cost/name = %v, want 1", w)
	}
}

// Rule 3b: a pick's text names a type, so candidates of that type are boosted.
func TestSynergyPickTextBoostsThatType(t *testing.T) {
	pick := trainerText("Misty", "Choose 1 of your [W] Pokémon.")
	syn := newSynergy([]draftCard{pick})
	tally := typeTally{}
	if w := weightOf(pokemon("Squirtle", "Water"), tally, syn); w != synergyMult {
		t.Errorf("Water Pokémon after picking Misty = %v, want %v", w, synergyMult)
	}
	if w := weightOf(pokemon("Charmander", "Fire"), tally, syn); w != 1 {
		t.Errorf("Fire Pokémon = %v, want 1", w)
	}
	colorless := newSynergy([]draftCard{trainerText("Poké Ball", "Put a [C] Pokémon in hand.")})
	if w := weightOf(pokemon("Meowth", "Colorless"), tally, colorless); w != 1 {
		t.Errorf("Colorless reference = %v, want 1", w)
	}
	// A Trainer is not a "candidate Pokémon": rule 3b does not apply to it.
	if w := weightOf(trainerText("Other", "Draw a card."), tally, syn); w != 1 {
		t.Errorf("Trainer under rule 3b = %v, want 1", w)
	}
}

// A card's text naming its own type is the type weight's job, not synergy's.
func TestSynergyOwnTypeNeverCounts(t *testing.T) {
	// A Water pick whose text says [W] must not boost other Water Pokémon (3b).
	pick := withAttackEffect(pokemon("Vaporeon", "Water"), "Take a [W] Energy.")
	if w := weightOf(pokemon("Squirtle", "Water"), typeTally{}, newSynergy([]draftCard{pick})); w != 1 {
		t.Errorf("Water pick naming [W], Water candidate = %v, want 1", w)
	}
	// ...but it still boosts a Fire Pokémon if its text names [R].
	fireText := withAttackEffect(pokemon("Flareon", "Fire"), "Take a [W] Energy.")
	if w := weightOf(pokemon("Squirtle", "Water"), typeTally{}, newSynergy([]draftCard{fireText})); w != synergyMult {
		t.Errorf("Fire pick naming [W], Water candidate = %v, want %v", w, synergyMult)
	}
	// A Water candidate whose own text says [W] is not boosted by Water picks (3a).
	candidate := withAttackEffect(pokemon("Kyogre", "Water"), "Take a [W] Energy.")
	if w := weightOf(candidate, typeTally{}, newSynergy([]draftCard{pokemon("Squirtle", "Water")})); w != 1 {
		t.Errorf("Water candidate naming its own [W] = %v, want 1", w)
	}
}

func TestSynergyDragonSymbol(t *testing.T) {
	syn := newSynergy([]draftCard{pokemon("Dratini", "Dragon")})
	if w := weightOf(trainerText("Lance", "Choose 1 of your [N] Pokémon."), typeTally{}, syn); w != synergyMult {
		t.Errorf("[N] Trainer after a Dragon pick = %v, want %v", w, synergyMult)
	}
	if w := weightOf(trainerText("Blaine", "Choose 1 of your [R] Pokémon."), typeTally{}, syn); w != 1 {
		t.Errorf("[R] Trainer after a Dragon pick = %v, want 1", w)
	}
}

func TestSynergyNeverStacks(t *testing.T) {
	// Squirtle-named Trainer text that also names Water: matches rules 1 and 2.
	picks := []draftCard{pokemon("Squirtle", "Water")}
	syn := newSynergy(picks)
	both := trainerText("X", "Heal your Squirtle and your [W] Pokémon.")
	if w := weightOf(both, tallyTypes(picks), syn); w != synergyMult {
		t.Errorf("two rules = %v, want one x%v", w, synergyMult)
	}
}

func idsOf(cards []draftCard) map[string]bool {
	out := map[string]bool{}
	for _, c := range cards {
		out[c.ID] = true
	}
	return out
}

// offeredCount counts cards in idSet offered over 20 seeds x 100 offers.
func offeredCount(byTier map[tier][]draftCard, picks []draftCard, ids map[string]bool) int {
	n := 0
	for seed := uint64(1); seed <= 20; seed++ {
		rng := newDraftRng(seed)
		for i := 0; i < 50; i++ {
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

func firstOfType(pool []draftCard, typ string) draftCard {
	for _, c := range pool {
		if c.CardType == typ && len(namedTypes(c)) == 0 {
			return c
		}
	}
	return draftCard{}
}

// Rule 3a on the shipped pool: Trainers that name only Water are offered more
// after a Water pick than after a Fire pick.
func TestOffersFavourTrainersNamingPickedType(t *testing.T) {
	useRealCards(t)
	pool := toDraftablePool()
	byTier := indexByTier(pool)
	var waterOnly []draftCard
	for _, c := range pool {
		named := typesNamedIn(effectText(c))
		if c.CardType == "Trainer" && len(named) == 1 && named["Water"] {
			waterOnly = append(waterOnly, c)
		}
	}
	if len(waterOnly) == 0 {
		t.Fatal("no Trainer in the pool names only Water")
	}
	ids := idsOf(waterOnly)
	water, fire := firstOfType(pool, "Water"), firstOfType(pool, "Fire")
	with := offeredCount(byTier, []draftCard{water}, ids)
	without := offeredCount(byTier, []draftCard{fire}, ids)
	t.Logf("%d Water-only Trainers: with Water pick %d, with Fire pick %d", len(waterOnly), with, without)
	if without == 0 || float64(with) < 1.5*float64(without) {
		t.Errorf("Water-naming Trainers offered %d vs %d, want >= 1.5x", with, without)
	}
}

// Rule 2 on the shipped pool: Trainers naming a picked Pokémon.
func TestOffersFavourTrainerNamingPick(t *testing.T) {
	useRealCards(t)
	pool := toDraftablePool()
	byTier := indexByTier(pool)
	var pick draftCard
	var named []draftCard
	for _, c := range pool {
		if c.CardType == "Trainer" || c.CardType == "Colorless" {
			continue
		}
		var n []draftCard
		for _, tr := range pool {
			if tr.CardType == "Trainer" && newSynergy([]draftCard{c}).nameText.MatchString(tr.TrainerText) {
				n = append(n, tr)
			}
		}
		if len(n) > len(named) {
			pick, named = c, n
		}
	}
	if len(named) == 0 {
		t.Fatal("no Trainer in the pool names a pool Pokémon")
	}
	control := firstOfType(pool, pick.CardType)
	with := offeredCount(byTier, []draftCard{pick}, idsOf(named))
	without := offeredCount(byTier, []draftCard{control}, idsOf(named))
	t.Logf("pick %s, %d naming Trainers: with %d, control (%s) %d", pick.Name, len(named), with, control.Name, without)
	if without == 0 || float64(with) < 1.5*float64(without) {
		t.Errorf("naming Trainers offered %d vs %d, want >= 1.5x", with, without)
	}
}

// How many shipped-pool cards each rule can fire on, so we know it is alive.
func TestSynergyRulesAreAliveInShippedPool(t *testing.T) {
	useRealCards(t)
	pool := toDraftablePool()
	allNames := newSynergy(pool)
	rule1, rule2, rule3, textTypes := 0, 0, 0, map[string]bool{}
	for _, c := range pool {
		if c.CardType == "Trainer" && allNames.nameText.MatchString(c.TrainerText) {
			rule1++
		}
		if len(namedTypes(c)) > 0 {
			rule2++
		}
		for _, typ := range namedTypes(c) {
			textTypes[typ] = true
		}
	}
	for _, c := range pool {
		if c.CardType != "Trainer" && textTypes[c.CardType] {
			rule3++
		}
	}
	t.Logf("pool %d cards: rule 1 (Trainer names a pool Pokémon) %d; rule 2 (text names a type) %d; rule 3 (Pokémon of a type some card's text names) %d",
		len(pool), rule1, rule2, rule3)
	if rule1 == 0 || rule2 == 0 || rule3 == 0 {
		t.Errorf("a rule can fire on nothing: %d %d %d", rule1, rule2, rule3)
	}
}
