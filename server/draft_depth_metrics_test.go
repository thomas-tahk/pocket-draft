package main

import (
	"fmt"

	"github.com/thomas-tahk/pocket-draft/engine"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// Evidence for the next-milestone draft issues (owner 2026-10-28, "Draft design:
// next milestone"). It measures; it asserts nothing about balance. Numbers go
// to docs/chief/research/2026-10-28-draft-depth-evidence.md.
//
//	PD_METRICS=1 go test -run TestDraftDepthMetrics -v
//
// Skipped otherwise so the normal suite stays fast. Seeds are fixed, so a re-run
// on the same code prints the same numbers.

const metricsDrafts = 500

// The bot drafts with the game seed; Auto-draft (POST /api/autodeck, the Home
// "Auto-draft (test)" button) runs the same draftBotIDs with seed^yourDraftSalt.
func metricsSeeds(salt uint64) []uint64 {
	seeds := make([]uint64, metricsDrafts)
	for i := range seeds {
		seeds[i] = (uint64(i) + 1) ^ salt
	}
	return seeds
}

func TestDraftDepthMetrics(t *testing.T) {
	if os.Getenv("PD_METRICS") != "1" {
		t.Skip("set PD_METRICS=1 to print the draft-depth numbers")
	}
	useRealCards(t)
	pool := toDraftablePool()
	poolByID := map[string]draftCard{}
	for _, c := range pool {
		poolByID[c.ID] = c
	}
	t.Run("a_picks_in_deck", func(t *testing.T) { metricsPicksInDeck(t, poolByID) })
	t.Run("b_pool_by_type", func(t *testing.T) { metricsPoolByType(t, pool) })
	t.Run("c_mega_ex", func(t *testing.T) { metricsMegaEx(t, pool) })
	t.Run("d_shop", func(t *testing.T) { metricsShop(t, pool) })
	t.Run("e_trainer_share", func(t *testing.T) { metricsTrainerShare(t, pool) })
}

type pickFate struct {
	inDeck, trainerDropped, offType, noRoom int
}

func metricsPicksInDeck(t *testing.T, poolByID map[string]draftCard) {
	for _, run := range []struct {
		label string
		salt  uint64
	}{{"draft bot (seed n)", 0}, {"Auto-draft (seed n ^ yourDraftSalt)", yourDraftSalt}} {
		var shares []float64
		hist := make([]int, totalDraftPacks+1)
		var fate pickFate
		slotPicked, slotFree, slotOther := 0, 0, 0
		picksAt2, picksAt1 := 0, 0
		decksWithStage2 := 0
		for _, seed := range metricsSeeds(run.salt) {
			pickIDs, deck, err := draftBotIDs(seed)
			if err != nil {
				t.Fatalf("seed %d: %v", seed, err)
			}
			pickNames := map[string]bool{}
			for _, id := range pickIDs {
				pickNames[poolByID[id].Name] = true
			}
			copies := map[string]int{}
			deckTypes := typeSet{}
			stage2 := false
			for _, id := range deck {
				rc := rawByID[id]
				copies[rc.Name]++
				if rc.CardType != "Trainer" {
					deckTypes.add(rc)
				}
				if rc.Stage == "Stage 2" {
					stage2 = true
				}
				switch {
				case pickNames[rc.Name]:
					slotPicked++
				case rc.CardType != "Trainer":
					slotFree++ // an under-evolution granted free at deckbuild
				default:
					slotOther++ // a Trainer not picked: only the shop could give it
				}
			}
			if stage2 {
				decksWithStage2++
			}
			in := 0
			for _, id := range pickIDs {
				c := poolByID[id]
				switch {
				case copies[c.Name] > 0:
					in++
					if copies[c.Name] >= 2 {
						picksAt2++
					} else {
						picksAt1++
					}
				case c.CardType == "Trainer":
					fate.trainerDropped++
				case !deckTypes.fits(c.rawCard):
					fate.offType++
				default:
					fate.noRoom++
				}
			}
			fate.inDeck += in
			hist[in]++
			shares = append(shares, float64(in)/float64(len(pickIDs)))
		}
		sort.Float64s(shares)
		mean := 0.0
		for _, s := range shares {
			mean += s
		}
		mean /= float64(len(shares))
		totalPicks := float64(metricsDrafts * totalDraftPacks)
		totalSlots := float64(metricsDrafts * draftDeckSize)
		t.Logf("(a) %s, %d drafts: share of the 20 picks with >=1 copy in the final deck", run.label, metricsDrafts)
		t.Logf("| mean | median | min | max |\n|---|---|---|---|\n| %.1f%% | %.1f%% | %.0f%% | %.0f%% |",
			100*mean, 100*median(shares), 100*shares[0], 100*shares[len(shares)-1])
		var b strings.Builder
		b.WriteString("| picks in deck (of 20) | drafts |\n|---|---|\n")
		for n, k := range hist {
			if k > 0 {
				fmt.Fprintf(&b, "| %d | %d |\n", n, k)
			}
		}
		t.Logf("distribution:\n%s", b.String())
		t.Logf("| pick fate | share of all picks |\n|---|---|\n| in deck at 2 copies | %.1f%% |\n| in deck at 1 copy | %.1f%% |\n| Trainer left out | %.1f%% |\n| Pokémon off the deck's attack types | %.1f%% |\n| Pokémon on-type, no room | %.1f%% |",
			100*float64(picksAt2)/totalPicks, 100*float64(picksAt1)/totalPicks,
			100*float64(fate.trainerDropped)/totalPicks, 100*float64(fate.offType)/totalPicks, 100*float64(fate.noRoom)/totalPicks)
		t.Logf("| deck slot source | share of 20 slots |\n|---|---|\n| picked card | %.1f%% |\n| free under-evolution | %.1f%% |\n| shop / other | %.1f%% |",
			100*float64(slotPicked)/totalSlots, 100*float64(slotFree)/totalSlots, 100*float64(slotOther)/totalSlots)
		t.Logf("decks holding a Stage 2 card: %d of %d (%.1f%%)", decksWithStage2, metricsDrafts, 100*float64(decksWithStage2)/metricsDrafts)
	}
	metricsOneCopyPerPick(t, poolByID)
}

// metricsOneCopyPerPick is a what-if, not a rule: the same bot drafts, but the
// deck may hold only 1 copy of each picked name (under-evolutions still free,
// at 1 copy each), every pick tried in pick order.
func metricsOneCopyPerPick(t *testing.T, poolByID map[string]draftCard) {
	byName := nameIndex()
	var shares []float64
	full := 0
	for _, seed := range metricsSeeds(0) {
		pickIDs, _, err := draftBotIDs(seed)
		if err != nil {
			t.Fatal(err)
		}
		picks := make([]draftCard, len(pickIDs))
		for i, id := range pickIDs {
			picks[i] = poolByID[id]
		}
		b := &botDeckBuilder{count: map[string]int{}, types: typeSet{}, byName: byName, limit: botMaxAttackTypes}
		b.addGroups(groupPicks(picks), 1)
		if len(b.ids) < draftDeckSize {
			b.limit = int(engine.Metal) + 1
			b.addGroups(groupPicks(picks), 1)
		}
		if len(b.ids) == draftDeckSize {
			full++
		}
		in := 0
		for _, p := range picks {
			if b.count[p.Name] > 0 {
				in++
			}
		}
		shares = append(shares, float64(in)/float64(len(picks)))
	}
	sort.Float64s(shares)
	mean := 0.0
	for _, s := range shares {
		mean += s
	}
	mean /= float64(len(shares))
	t.Logf("(a) what-if 1 copy per pick (same bot drafts, seed n): picks in deck mean %.1f%%, median %.1f%%, min %.0f%%, max %.0f%%; decks that reach 20 cards from picks + free under-evolutions: %d of %d",
		100*mean, 100*median(shares), 100*shares[0], 100*shares[len(shares)-1], full, metricsDrafts)
}

func median(sorted []float64) float64 {
	n := len(sorted)
	if n%2 == 1 {
		return sorted[n/2]
	}
	return (sorted[n/2-1] + sorted[n/2]) / 2
}

var metricsTypes = []string{"Grass", "Fire", "Water", "Lightning", "Psychic", "Fighting", "Darkness", "Metal", "Dragon", "Colorless", "Trainer"}

func metricsPoolByType(t *testing.T, pool []draftCard) {
	type row struct{ all, ex, mega, t3, t4, t5, stage2, stage2ex int }
	rows := map[string]*row{}
	for _, typ := range metricsTypes {
		rows[typ] = &row{}
	}
	for _, c := range pool {
		r := rows[c.CardType]
		if r == nil {
			r = &row{}
			rows[c.CardType] = r
		}
		r.all++
		switch tierOf(c) {
		case tier1:
			r.ex++
		case tier3:
			r.t3++
		case tier4:
			r.t4++
		case tier5:
			r.t5++
		}
		if c.IsMegaEx {
			r.mega++
		}
		if c.Stage == "Stage 2" {
			r.stage2++
			if c.IsEx {
				r.stage2ex++
			}
		}
	}
	var b strings.Builder
	b.WriteString("| type | pool cards | ex (tier 1) | of which Mega ex | ◊◊◊ (tier 3) | ◊◊ | ◊ | Stage 2 (of which ex) |\n|---|---|---|---|---|---|---|---|\n")
	tot := row{}
	for _, typ := range metricsTypes {
		r := rows[typ]
		fmt.Fprintf(&b, "| %s | %d | %d | %d | %d | %d | %d | %d (%d) |\n", typ, r.all, r.ex, r.mega, r.t3, r.t4, r.t5, r.stage2, r.stage2ex)
		tot.all += r.all
		tot.ex += r.ex
		tot.mega += r.mega
		tot.t3 += r.t3
		tot.t4 += r.t4
		tot.t5 += r.t5
		tot.stage2 += r.stage2
		tot.stage2ex += r.stage2ex
	}
	fmt.Fprintf(&b, "| **all** | %d | %d | %d | %d | %d | %d | %d (%d) |\n", tot.all, tot.ex, tot.mega, tot.t3, tot.t4, tot.t5, tot.stage2, tot.stage2ex)
	t.Logf("(b) draft pool (%d cards, onlyWorkingText=%v) by card type and tier:\n%s", len(pool), onlyWorkingTextInDraft, b.String())

	// ex names per type in all of cards.json (any text state), for comparison.
	exNames, megaNames := map[string]map[string]bool{}, map[string]map[string]bool{}
	for _, rc := range rawByID {
		if rc.ExKind == "regular" || rc.CardType == "Trainer" || !isDraftCandidate(rc) {
			continue
		}
		m := exNames
		if rc.ExKind == "mega-ex" {
			m = megaNames
		}
		if m[rc.CardType] == nil {
			m[rc.CardType] = map[string]bool{}
		}
		m[rc.CardType][rc.Name] = true
	}
	b.Reset()
	b.WriteString("| type | ex names in pool | ex names in cards.json | Mega ex names in cards.json |\n|---|---|---|---|\n")
	for _, typ := range metricsTypes[:len(metricsTypes)-1] {
		fmt.Fprintf(&b, "| %s | %d | %d | %d |\n", typ, rows[typ].ex, len(exNames[typ]), len(megaNames[typ]))
	}
	t.Logf("(b) ex by type, pool vs all of cards.json (non-cosmetic printings, distinct names):\n%s", b.String())
}

func metricsMegaEx(t *testing.T, pool []draftCard) {
	inPool := 0
	for _, c := range pool {
		if c.IsMegaEx {
			inPool++
		}
	}
	var ids []string
	for id, rc := range rawByID {
		if rc.ExKind == "mega-ex" {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	names := map[string]bool{}
	candidates, working := 0, 0
	reasons := map[string]int{}
	stages := map[string]int{}
	missingAttack := map[string]string{} // effect text -> first card name
	missingAbility := map[string]string{}
	seenName := map[string]bool{}
	for _, id := range ids {
		rc := rawByID[id]
		names[rc.Name] = true
		if !isDraftCandidate(rc) {
			continue
		}
		candidates++
		if textWorks(rc) {
			working++
		}
		if seenName[rc.Name] {
			continue // reasons counted once per name
		}
		seenName[rc.Name] = true
		stages[rc.Stage]++
		c, err := toEngineCard(rc)
		if err != nil {
			reasons["does not convert to an engine card"]++
			continue
		}
		if rc.Ability != nil && c.Ability == nil {
			reasons["Ability has no wording in the engine"]++
			missingAbility[rc.Ability.Effect] = rc.Name
		}
		attackMissing := false
		for _, a := range rc.Attacks {
			if a.Effect != "" && effectsByID[rc.ID][a.Name] == nil {
				attackMissing = true
				missingAttack[a.Effect] = rc.Name
			}
		}
		if attackMissing {
			reasons["an attack's text has no authored effect"]++
		}
		if !lineWorks(rc) {
			reasons["its earlier stage has no working printing"]++
		}
		if ownTextWorks(rc) && lineWorks(rc) {
			reasons["works (would be in pool)"]++
		}
	}
	t.Logf("(c) Mega ex: cards.json %d ids / %d names; non-cosmetic draft candidates %d; working text %d; in the draft pool %d (onlyWorkingText=%v)",
		len(ids), len(names), candidates, working, inPool, onlyWorkingTextInDraft)
	t.Logf("by distinct name (%d): stages %v", len(seenName), stages)
	var b strings.Builder
	for _, k := range sortedKeys(reasons) {
		fmt.Fprintf(&b, "| %s | %d |\n", k, reasons[k])
	}
	t.Logf("| reason (a name can have several) | names |\n|---|---|\n%s", b.String())
	n := 0
	for _, txt := range sortedKeys(missingAttack) {
		if n >= 8 {
			break
		}
		t.Logf("missing attack text: %s: %q", missingAttack[txt], txt)
		n++
	}
	t.Logf("distinct missing attack texts: %d; distinct missing Ability texts: %d", len(missingAttack), len(missingAbility))
	for _, txt := range sortedKeys(missingAbility) {
		t.Logf("missing Ability text: %s: %q", missingAbility[txt], txt)
	}
	// Do non-Mega ex cards with the same texts also miss them? (are the texts Mega-only?)
	otherUsers := 0
	for _, txt := range sortedKeys(missingAttack) {
		for _, rc := range rawByID {
			if rc.ExKind != "mega-ex" && hasAttackText(rc, txt) {
				otherUsers++
				break
			}
		}
	}
	t.Logf("of those attack texts, %d also appear on a non-Mega card", otherUsers)
}

func hasAttackText(rc rawCard, txt string) bool {
	for _, a := range rc.Attacks {
		if a.Effect == txt {
			return true
		}
	}
	return false
}

func sortedKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// staplesShelf mirrors src/shop/staples.ts (there is no Go twin of the shop).
var staplesShelf = []string{"A2-150", "A1-225", "A3a-064", "B1-225", "A1-223", "A2-146", "A1-219", "A2-155", "B2a-089", "A2-152", "A2b-070", "A3-155"}

var searchText = regexp.MustCompile(`(?i)from your deck into your hand|search your deck|from your deck onto your bench|look at the top`)

func metricsShop(t *testing.T, pool []draftCard) {
	inPoolName := map[string]bool{}
	poolTrainers := map[string]bool{}
	for _, c := range pool {
		inPoolName[c.Name] = true
		if c.CardType == "Trainer" {
			poolTrainers[c.Name] = true
		}
	}
	var b strings.Builder
	b.WriteString("| staple | id | text works | also in draft pool |\n|---|---|---|---|\n")
	shopName := map[string]bool{}
	stapleWorking := 0
	for _, id := range staplesShelf {
		rc := rawByID[id]
		w := textWorks(rc)
		if w {
			stapleWorking++
			shopName[rc.Name] = true // shopPool keeps working cards only (App.tsx)
		}
		fmt.Fprintf(&b, "| %s (%s) | %s | %v | %v |\n", rc.Name, rc.TrainerKind, id, w, inPoolName[rc.Name])
	}
	t.Logf("(d) staple shelf, %d of %d playable:\n%s", stapleWorking, len(staplesShelf), b.String())

	// Every Trainer name in cards.json (non-cosmetic printings) and where it can come from.
	allTrainer, workingTrainer := map[string]bool{}, map[string]bool{}
	for _, rc := range rawByID {
		if rc.CardType != "Trainer" || !isDraftCandidate(rc) {
			continue
		}
		allTrainer[rc.Name] = true
		if textWorks(rc) {
			workingTrainer[rc.Name] = true
		}
	}
	reach := map[string]bool{}
	for n := range poolTrainers {
		reach[n] = true
	}
	for n := range shopName {
		reach[n] = true
	}
	shopOnly := 0
	for n := range shopName {
		if !poolTrainers[n] {
			shopOnly++
		}
	}
	t.Logf("Trainer names: cards.json %d, working %d, draft pool %d, shop staples (playable) %d (of them not in the draft pool: %d), reachable draft ∪ shop %d",
		len(allTrainer), len(workingTrainer), len(poolTrainers), len(shopName), shopOnly, len(reach))
	t.Logf("same-type ex shelf: Pokémon only (ex / Mega ex of a drafted type); it never offers a Trainer")

	// Rare Candy and search / tutor Trainers.
	want := map[string]bool{"Rare Candy": true, "Poké Ball": true, "Pokémon Communication": true, "Clemont": true, "Team Galactic Grunt": true}
	for _, rc := range rawByID {
		if rc.CardType == "Trainer" && searchText.MatchString(rc.TrainerText) {
			want[rc.Name] = true
		}
	}
	b.Reset()
	poolTier := map[string]tier{}
	for _, c := range pool {
		poolTier[c.Name] = tierOf(c)
	}
	b.WriteString("| card | kind | printings in cards.json | text works | in draft pool (tier) | on shop shelf |\n|---|---|---|---|---|---|\n")
	for _, name := range sortedKeys(want) {
		printings, kind, works := 0, "", false
		text := ""
		for _, rc := range rawByID {
			if rc.Name != name {
				continue
			}
			printings++
			kind = rc.TrainerKind
			text = rc.TrainerText
			works = works || textWorks(rc)
		}
		_ = text
		inPool := "no"
		if poolTrainers[name] {
			inPool = fmt.Sprintf("yes (%d)", poolTier[name])
		}
		fmt.Fprintf(&b, "| %s | %s | %d | %v | %s | %v |\n", name, kind, printings, works, inPool, shopName[name])
	}
	t.Logf("Rare Candy and search / tutor Trainers (text matches %q, plus the names asked about):\n%s", searchText.String(), b.String())
	for _, name := range sortedKeys(want) {
		for _, id := range sortedKeys(rawByID) {
			if rawByID[id].Name == name {
				t.Logf("text %s (%s): %q", name, id, rawByID[id].TrainerText)
				break
			}
		}
	}

	// Stage 2 lines in the pool.
	s2, s2ex := 0, 0
	byType := map[string]int{}
	for _, c := range pool {
		if c.Stage == "Stage 2" {
			s2++
			byType[c.CardType]++
			if c.IsEx {
				s2ex++
			}
		}
	}
	pokemon := 0
	for _, c := range pool {
		if c.CardType != "Trainer" {
			pokemon++
		}
	}
	t.Logf("Stage 2 lines in the pool: %d of %d Pokémon (%d ex); by type %v", s2, pokemon, s2ex, byType)

	// How often a draft even shows these cards, with a picker that ignores type
	// (no Trainer lean from picks) and with the bot.
	watch := []string{"Rare Candy", "Poké Ball", "Pokémon Communication", "Clemont", "Team Galactic Grunt", "Pokédex", "Arven"}
	byTier := indexByTier(pool)
	byName := nameIndex()
	for _, run := range []struct {
		label  string
		random bool
	}{{"random picker", true}, {"draft bot", false}} {
		seen := map[string]int{}
		stage2Picks := 0
		for _, seed := range metricsSeeds(0) {
			rng, pickRng := newDraftRng(seed), newDraftRng(seed^0x1234)
			var picks []draftCard
			shown := map[string]bool{}
			for round := 0; round < totalDraftPacks; round++ {
				offer, err := generateOffer(byTier, picks, rng.Float64)
				if err != nil {
					t.Fatal(err)
				}
				for _, c := range offer {
					shown[c.Name] = true
				}
				var p draftCard
				if run.random {
					p = offer[pickRng.IntN(len(offer))]
				} else {
					p = botPick(offer, picks, byName)
				}
				if p.Stage == "Stage 2" {
					stage2Picks++
				}
				picks = append(picks, p)
			}
			for _, n := range watch {
				if shown[n] {
					seen[n]++
				}
			}
		}
		b.Reset()
		b.WriteString("| card | drafts that offer it at least once |\n|---|---|\n")
		for _, n := range watch {
			fmt.Fprintf(&b, "| %s | %.1f%% |\n", n, 100*float64(seen[n])/metricsDrafts)
		}
		t.Logf("(d) %s, %d drafts: Stage 2 picks per draft %.2f\n%s", run.label, metricsDrafts, float64(stage2Picks)/metricsDrafts, b.String())
	}
}

func metricsTrainerShare(t *testing.T, pool []draftCard) {
	byTier := indexByTier(pool)
	poolT, poolTS := 0, 0
	for _, c := range pool {
		if c.CardType == "Trainer" {
			poolT++
			if len(supportTypes(c)) > 0 {
				poolTS++
			}
		}
	}
	// Trainer share of each tier, and the share an unbiased slot would show.
	expected := 0.0
	var b strings.Builder
	b.WriteString("| tier | slot weight | cards | Trainers | Trainer share |\n|---|---|---|---|---|\n")
	totalW := 0.0
	for _, tr := range tierOrder {
		totalW += slotTierWeights[tr]
	}
	for _, tr := range tierOrder {
		n, k := len(byTier[tr]), 0
		for _, c := range byTier[tr] {
			if c.CardType == "Trainer" {
				k++
			}
		}
		share := 0.0
		if n > 0 {
			share = float64(k) / float64(n)
		}
		expected += slotTierWeights[tr] / totalW * share
		fmt.Fprintf(&b, "| %d | %.2f | %d | %d | %.1f%% |\n", tr, slotTierWeights[tr], n, k, 100*share)
	}
	t.Logf("(e) pool: %d Trainers of %d cards (%.1f%%); %d of the Trainers support a type (%.1f%%). Tier-weighted Trainer share of a slot with no lean: %.1f%%\n%s",
		poolT, len(pool), 100*float64(poolT)/float64(len(pool)), poolTS, 100*float64(poolTS)/float64(poolT), 100*expected, b.String())

	type phase struct{ slots, trainers, support, supportOn int }
	phaseOf := func(round int) int {
		switch {
		case round < 7:
			return 0
		case round < 14:
			return 1
		}
		return 2
	}
	phaseName := []string{"early (picks 1-7)", "mid (8-14)", "late (15-20)"}
	byName := nameIndex()
	for _, run := range []struct {
		label  string
		salt   uint64
		random bool
	}{
		{"draft bot", 0, false},
		{"Auto-draft", yourDraftSalt, false},
		{"random picker (stand-in for a human who ignores types)", 0xabcdef, true},
	} {
		phases := make([]phase, 3)
		trainerPicks, deckTrainers := 0, 0
		for _, seed := range metricsSeeds(run.salt) {
			rng := newDraftRng(seed)
			pickRng := newDraftRng(seed ^ 0x1234)
			picks := make([]draftCard, 0, totalDraftPacks)
			for round := 0; round < totalDraftPacks; round++ {
				offer, err := generateOffer(byTier, picks, rng.Float64)
				if err != nil {
					t.Fatal(err)
				}
				tally := tallyTypes(picks)
				p := &phases[phaseOf(round)]
				for _, c := range offer {
					p.slots++
					if c.CardType != "Trainer" {
						continue
					}
					p.trainers++
					if st := supportTypes(c); len(st) > 0 {
						p.support++
						for _, typ := range st {
							if tally[typ] > 0 {
								p.supportOn++ // supports a type already picked
								break
							}
						}
					}
				}
				var pick draftCard
				if run.random {
					pick = offer[pickRng.IntN(len(offer))]
				} else {
					pick = botPick(offer, picks, byName)
				}
				if pick.CardType == "Trainer" {
					trainerPicks++
				}
				picks = append(picks, pick)
			}
			if !run.random {
				for _, id := range botDeckIDs(picks, byName) {
					if rawByID[id].CardType == "Trainer" {
						deckTrainers++
					}
				}
			}
		}
		b.Reset()
		b.WriteString("| phase | Trainer share of offer slots | type-support share of offered Trainers | of those, supporting a type already picked |\n|---|---|---|---|\n")
		for i, p := range phases {
			fmt.Fprintf(&b, "| %s | %.1f%% | %.1f%% | %.1f%% |\n", phaseName[i], 100*float64(p.trainers)/float64(p.slots),
				100*float64(p.support)/float64(max(1, p.trainers)), 100*float64(p.supportOn)/float64(max(1, p.support)))
		}
		deck := "n/a (no deckbuild in Go for this picker)"
		if !run.random {
			deck = fmt.Sprintf("%.1f%% (%.2f of 20 cards)", 100*float64(deckTrainers)/float64(metricsDrafts*draftDeckSize), float64(deckTrainers)/metricsDrafts)
		}
		t.Logf("%s, %d drafts: Trainer share of picks %.1f%%; of final-deck slots %s\n%s", run.label, metricsDrafts,
			100*float64(trainerPicks)/float64(metricsDrafts*totalDraftPacks), deck, b.String())
	}
}
