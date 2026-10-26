package main

import (
	"os"
	"regexp"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// Niche odds (#126): a Trainer and the Pokémon it supports raise each other's
// odds, whichever is picked first. Mirrored in e2e/draft-offer.spec.ts.

func poolCard(t *testing.T, pool []draftCard, name string) draftCard {
	t.Helper()
	for _, c := range pool {
		if c.Name == name {
			return c
		}
	}
	t.Fatalf("%q is not in the draft pool", name)
	return draftCard{}
}

// nicheOdds counts how often target is offered with the given picks, over
// 20 seeds x 400 offers.
func nicheOdds(t *testing.T, byTier map[tier][]draftCard, picks []draftCard, target draftCard) (n, slots int) {
	t.Helper()
	for seed := uint64(1); seed <= 20; seed++ {
		rng := newDraftRng(seed)
		for i := 0; i < 400; i++ {
			offer, err := generateOffer(byTier, picks, rng.Float64)
			if err != nil {
				t.Fatal(err)
			}
			slots += len(offer)
			for _, c := range offer {
				if c.ID == target.ID {
					n++
				}
			}
		}
	}
	return n, slots
}

// Picking a Trainer whose text names a Pokémon raises that Pokémon's odds.
// Control: a Trainer that names nothing and has no type (Poké Ball), so the
// only difference between the two runs is the naming.
func TestPickedTrainerRaisesThePokemonItNames(t *testing.T) {
	useRealCards(t)
	pool := toDraftablePool()
	byTier := indexByTier(pool)
	nemona, pawmot := poolCard(t, pool, "Nemona"), poolCard(t, pool, "Pawmot")
	control := poolCard(t, pool, "Poké Ball")
	if !containsWord(nemona.TrainerText, "Pawmot") {
		t.Fatalf("Nemona no longer names Pawmot: %q", nemona.TrainerText)
	}
	with, slots := nicheOdds(t, byTier, []draftCard{nemona}, pawmot)
	without, _ := nicheOdds(t, byTier, []draftCard{control}, pawmot)
	t.Logf("Pawmot offered %d of %d slots after Nemona, %d after the control Poké Ball (%.2fx)",
		with, slots, without, float64(with)/float64(without))
	if without == 0 || float64(with) < 2*float64(without) {
		t.Errorf("Pawmot offered %d vs %d, want >= 2x", with, without)
	}
}

// The existing direction, kept: picking the Pokémon raises the Trainer.
func TestPickedPokemonRaisesTheTrainerNamingIt(t *testing.T) {
	useRealCards(t)
	pool := toDraftablePool()
	byTier := indexByTier(pool)
	nemona, pawmot := poolCard(t, pool, "Nemona"), poolCard(t, pool, "Pawmot")
	// Control: a Lightning Pokémon that no Trainer names, so the type lean is equal.
	control := draftCard{}
	for _, c := range pool {
		if c.CardType == "Lightning" && c.Name != "Pawmot" && !c.IsEx && len(namedTypes(c)) == 0 &&
			!newSynergy([]draftCard{c}).namesPickedIn(nemona.TrainerText) {
			control = c
			break
		}
	}
	with, slots := nicheOdds(t, byTier, []draftCard{pawmot}, nemona)
	without, _ := nicheOdds(t, byTier, []draftCard{control}, nemona)
	t.Logf("Nemona offered %d of %d slots after Pawmot, %d after %s (%.2fx)",
		with, slots, without, control.Name, float64(with)/float64(without))
	if without == 0 || float64(with) < 2*float64(without) {
		t.Errorf("Nemona offered %d vs %d, want >= 2x", with, without)
	}
}

// Archetype cards, both directions, on real cards: Eevee Bag <-> an
// Eeveelution, and Professor Sada <-> an Ancient Pokémon.
func TestNicheTableBothDirections(t *testing.T) {
	useRealCards(t)
	pool := toDraftablePool()
	byTier := indexByTier(pool)
	control := poolCard(t, pool, "Poké Ball")
	for _, tc := range []struct{ trainer, pokemon string }{
		{"Eevee Bag", "Vaporeon"},
		{"Professor Sada", "Roaring Moon"},
		{"Professor Turo", "Iron Valiant"},
	} {
		tr, mon := poolCard(t, pool, tc.trainer), poolCard(t, pool, tc.pokemon)
		// Control for the Pokémon-first direction: another Pokémon of the same
		// type that no table entry or Trainer text ties to the Trainer.
		var sameType draftCard
		for _, c := range pool {
			if c.CardType == mon.CardType && c.Name != mon.Name && !c.IsEx && len(namedTypes(c)) == 0 &&
				len(nicheGroupsOf(c)) == 0 {
				sameType = c
				break
			}
		}
		a, slots := nicheOdds(t, byTier, []draftCard{tr}, mon)
		b, _ := nicheOdds(t, byTier, []draftCard{control}, mon)
		t.Logf("%s offered %d of %d after %s, %d after Poké Ball (%.2fx)", mon.Name, a, slots, tr.Name, b, float64(a)/float64(b))
		if b == 0 || float64(a) < 1.5*float64(b) {
			t.Errorf("%s after %s: %d vs %d, want >= 1.5x", mon.Name, tr.Name, a, b)
		}
		c, _ := nicheOdds(t, byTier, []draftCard{mon}, tr)
		d, _ := nicheOdds(t, byTier, []draftCard{sameType}, tr)
		t.Logf("%s offered %d of %d after %s, %d after %s (%.2fx)", tr.Name, c, slots, mon.Name, d, sameType.Name, float64(c)/float64(d))
		// The offer counts are small and noisy, so the assertion is on the weight:
		// exactly the synergy multiplier over the same-type control.
		wMon := weightOf(tr, tallyTypes([]draftCard{mon}), newSynergy([]draftCard{mon}))
		wCtl := weightOf(tr, tallyTypes([]draftCard{sameType}), newSynergy([]draftCard{sameType}))
		if wMon != synergyMult*wCtl {
			t.Errorf("%s weight after %s = %v, after %s = %v, want exactly %vx", tr.Name, mon.Name, wMon, sameType.Name, wCtl, synergyMult)
		}
		if d > 0 && c <= d {
			t.Errorf("%s after %s: %d offers vs %d, want more", tr.Name, mon.Name, c, d)
		}
	}
}

// Raised, never guaranteed; applied once, never stacked; no cross-talk.
func TestNicheRulesWeightsAndTraps(t *testing.T) {
	tally := typeTally{}
	// Text rule: "Snorlax ex" is named by text saying "Snorlax"; "Snorlax" is
	// not named by text saying only "Snorlax ex"; longer words never match.
	say := func(text string) synergy { return newSynergy([]draftCard{trainerText("T", text)}) }
	if w := weightOf(pokemon("Snorlax ex", "Colorless"), tally, say("Heal your Snorlax.")); w != synergyMult {
		t.Errorf("Snorlax ex named as Snorlax = %v, want %v", w, synergyMult)
	}
	if w := weightOf(pokemon("Snorlax", "Colorless"), tally, say("Heal your Snorlax ex.")); w != 1 {
		t.Errorf("Snorlax named only as Snorlax ex = %v, want 1", w)
	}
	if w := weightOf(pokemon("Snorlax", "Colorless"), tally, say("Heal your Snorlaxes.")); w != 1 {
		t.Errorf("longer word = %v, want 1", w)
	}
	// Once only: named by the text and in a group of the same Trainer.
	sada := trainerText("Professor Sada", "Heal your Roaring Moon.")
	if w := weightOf(pokemon("Roaring Moon", "Darkness"), tally, newSynergy([]draftCard{sada})); w != synergyMult {
		t.Errorf("text and table together = %v, want %v (never stacked)", w, synergyMult)
	}
	// Table is exact and cross-kind: a similar name, a sibling Pokémon, and a
	// sibling Trainer are not raised.
	syn := newSynergy([]draftCard{pokemon("Roaring Moon", "Darkness")})
	for name, c := range map[string]draftCard{
		"Professor Sada":    trainerText("Professor Sada", "x"),
		"Ancient Booster":   trainerText("Ancient Booster Energy Capsule", "x"),
		"unrelated Trainer": trainerText("Professor Turo", "x"),
		"sibling Pokémon":   pokemon("Great Tusk", "Fighting"),
		"Eeveelution":       {rawCard: rawCard{ID: "v", Name: "Vaporeon", CardType: "Water", EvolvesFrom: "Eevee"}},
	} {
		want := 1.0
		if name == "Professor Sada" || name == "Ancient Booster" {
			want = synergyMult
		}
		if w := weightOf(c, tally, syn); w != want {
			t.Errorf("%s after Roaring Moon = %v, want %v", name, w, want)
		}
	}
	// Eevee ex is not an Eeveelution (it is a Basic); #128 owns it.
	bag := newSynergy([]draftCard{trainerText("Eevee Bag", "Heal each Pokémon that evolves from Eevee.")})
	vap := draftCard{rawCard: rawCard{ID: "v", Name: "Vaporeon", CardType: "Water", EvolvesFrom: "Eevee"}}
	if w := weightOf(vap, tally, bag); w != synergyMult {
		t.Errorf("Vaporeon after Eevee Bag = %v, want %v", w, synergyMult)
	}
	if got := nicheGroupsOf(pokemon("Eevee ex", "Colorless")); len(got) != 0 {
		t.Errorf("Eevee ex is in niche groups %v, want none", got)
	}
}

// Every name in the table must be spelled like a card in cards.json, or the
// entry silently does nothing.
func TestNicheTableNamesExistInCards(t *testing.T) {
	useRealCards(t)
	have := map[string]bool{}
	for _, rc := range rawByID {
		have[pokemonBaseName(rc.Name)] = true
		have[rc.Name] = true
	}
	for _, g := range nicheGroups {
		for _, n := range append(append([]string{}, g.trainers...), g.pokemon...) {
			if !have[n] {
				t.Errorf("%s: %q is not a card name in cards.json", g.label, n)
			}
		}
		for _, n := range g.evolvesFrom {
			if !have[n] {
				t.Errorf("%s: evolvesFrom %q is not a card name", g.label, n)
			}
		}
	}
}

// Kiawe names "Alolan Marowak". It raises that card and not the Fighting
// Marowak, and picking Marowak does not raise Kiawe.
func TestKiaweNamesAlolanMarowakNotMarowak(t *testing.T) {
	useRealCards(t)
	pool := toDraftablePool()
	kiawe := poolCard(t, pool, "Kiawe")
	var marowak, alolan draftCard
	for _, c := range pool {
		switch c.Name {
		case "Marowak":
			if marowak.ID == "" {
				marowak = c
			}
		case "Alolan Marowak":
			alolan = c
		}
	}
	if marowak.ID != "A1-152" {
		t.Fatalf("Marowak in the pool is %q, want A1-152", marowak.ID)
	}
	tally := typeTally{}
	syn := newSynergy([]draftCard{kiawe})
	if w := weightOf(marowak, tally, syn); w != 1 {
		t.Errorf("Marowak after Kiawe = %v, want 1", w)
	}
	if alolan.ID == "" {
		t.Log("Alolan Marowak is not in the draft pool today; checked on a synthetic card")
		alolan = pokemon("Alolan Marowak", "Fire")
	}
	if w := weightOf(alolan, tally, syn); w != synergyMult {
		t.Errorf("Alolan Marowak after Kiawe = %v, want %v", w, synergyMult)
	}
	// Rule 1's direction: picking Marowak must not raise Kiawe; picking the Alolan one does.
	base := weightOf(kiawe, tally, synergy{})
	if w := weightOf(kiawe, tally, newSynergy([]draftCard{marowak})); w != base {
		t.Errorf("Kiawe after Marowak = %v, want unchanged %v", w, base)
	}
	if w := weightOf(kiawe, tally, newSynergy([]draftCard{alolan})); w != synergyMult*base {
		t.Errorf("Kiawe after Alolan Marowak = %v, want %vx", w, synergyMult)
	}
}

// Eevee Bag says "evolve from Eevee": it names no Eevee ex.
func TestEeveeBagDoesNotRaiseEeveeEx(t *testing.T) {
	useRealCards(t)
	pool := toDraftablePool()
	bag := poolCard(t, pool, "Eevee Bag")
	eevee := draftCard{rawCard: rawCard{ID: "ex", Name: "Eevee ex", CardType: "Colorless"}}
	tally := typeTally{}
	if got, want := weightOf(eevee, tally, newSynergy([]draftCard{bag})), weightOf(eevee, tally, synergy{}); got != want {
		t.Errorf("Eevee ex after Eevee Bag = %v, want unchanged %v", got, want)
	}
}

// Rule 1 and rule 4 read " ex" the same way: text naming only "Snorlax ex"
// names only the ex card, in both directions; text naming "Snorlax" names both.
func TestExIsReadTheSameBothDirections(t *testing.T) {
	tally := typeTally{}
	snorlax, snorlaxEx := pokemon("Snorlax", "Colorless"), pokemon("Snorlax ex", "Colorless")
	onlyEx, plain := trainerText("A", "Heal your Snorlax ex."), trainerText("B", "Heal your Snorlax.")
	one := func(c draftCard, picks ...draftCard) float64 { return weightOf(c, tally, newSynergy(picks)) }
	cases := []struct {
		name string
		got  float64
		want float64
	}{
		{"rule 4: Snorlax after text 'Snorlax ex'", one(snorlax, onlyEx), 1},
		{"rule 4: Snorlax ex after text 'Snorlax ex'", one(snorlaxEx, onlyEx), synergyMult},
		{"rule 4: Snorlax after text 'Snorlax'", one(snorlax, plain), synergyMult},
		{"rule 4: Snorlax ex after text 'Snorlax'", one(snorlaxEx, plain), synergyMult},
		{"rule 1: text 'Snorlax ex' after picked Snorlax", one(onlyEx, snorlax), 1},
		{"rule 1: text 'Snorlax ex' after picked Snorlax ex", one(onlyEx, snorlaxEx), synergyMult},
		{"rule 1: text 'Snorlax' after picked Snorlax", one(plain, snorlax), synergyMult},
		{"rule 1: text 'Snorlax' after picked Snorlax ex", one(plain, snorlaxEx), synergyMult},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s = %v, want %v", c.name, c.got, c.want)
		}
	}
}

// The TS table (src/draft/nicheSupport.ts) must hold the same groups as the Go
// one, and the Ancient / Future / Ultra Beast members must be the engine's.
func TestNicheTSTableMatchesTheEngine(t *testing.T) {
	useRealCards(t)
	src, err := os.ReadFile("../src/draft/nicheSupport.ts")
	if err != nil {
		t.Fatal(err)
	}
	body := string(src)
	body = body[strings.Index(body, "export const NICHE_GROUPS"):]
	body = body[:strings.Index(body, "];\n")]
	quoted := regexp.MustCompile(`'([^']+)'`)
	names := func(block, field string) []string {
		m := regexp.MustCompile(field + `: \[([^\]]*)\]`).FindStringSubmatch(block)
		var out []string
		if m != nil {
			for _, q := range quoted.FindAllStringSubmatch(m[1], -1) {
				out = append(out, q[1])
			}
		}
		sort.Strings(out)
		return out
	}
	blocks := regexp.MustCompile(`label: '([^']+)'`).FindAllStringSubmatchIndex(body, -1)
	if len(blocks) != len(nicheGroups) {
		t.Fatalf("TS has %d groups, Go has %d", len(blocks), len(nicheGroups))
	}
	// Every Pokémon the engine's rule calls Future, by name before " ex".
	var future []string
	seen := map[string]bool{}
	for _, rc := range rawByID {
		n := pokemonBaseName(rc.Name)
		if rc.CardType != "Trainer" && isFuture(engine.Card{Name: rc.Name}) && !seen[n] {
			seen[n] = true
			future = append(future, n)
		}
	}
	for i, g := range nicheGroups {
		end := len(body)
		if i+1 < len(blocks) {
			end = blocks[i+1][0]
		}
		block := body[blocks[i][0]:end]
		if label := body[blocks[i][2]:blocks[i][3]]; label != g.label {
			t.Fatalf("group %d: TS %q, Go %q", i, label, g.label)
		}
		want := g.pokemon
		if g.fits != nil {
			want = future
		}
		for field, w := range map[string][]string{"trainers": g.trainers, "pokemon": want, "evolvesFrom": g.evolvesFrom} {
			if got := names(block, field); !slices.Equal(got, sortedCopy(w)) {
				t.Errorf("%s %s: TS %v, Go %v", g.label, field, got, sortedCopy(w))
			}
		}
	}
}

func sortedCopy(in []string) []string {
	out := append([]string{}, in...)
	sort.Strings(out)
	return out
}
