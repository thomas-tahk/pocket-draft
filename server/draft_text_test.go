package main

import (
	"encoding/json"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// draftSeeds is how many seeded drafts each test plays through.
const draftSeeds = 400

// modelable reports, independently of the code under test, whether the engine
// can play rc as printed: it converts to an engine card, has an Ability only if its
// wording is authored, and every attack with text has an authored effect.
func modelable(rc rawCard) bool {
	if _, err := toEngineCard(rc); err != nil {
		return false
	}
	if rc.CardType == "Trainer" && trainerEffects[rc.TrainerText] == nil && trainerPassives[rc.TrainerText] == nil && trainerBuilt[rc.TrainerText] == nil &&
		!strings.HasPrefix(rc.TrainerText, "Play this card as if it were a 40-HP Basic") {
		return false
	}
	if _, ok := abilityTexts[abilityWording(rc)]; rc.Ability != nil && !ok {
		return false
	}
	for _, a := range rc.Attacks {
		if a.Effect != "" && effectsByID[rc.ID][a.Name] == nil {
			return false
		}
	}
	return true
}

func TestDraftPoolOffersOnlyCardsWhoseTextWorks(t *testing.T) {
	useRealCards(t)

	pool := toDraftablePool()

	if len(pool) == 0 {
		t.Fatal("draft pool is empty")
	}
	for _, c := range pool {
		if !modelable(c.rawCard) {
			t.Errorf("pool offers %s (%s), whose text the engine can't model", c.ID, c.Name)
		}
	}
}

// draftIDs plays a 16-pack draft, choosing from each offer with choose, and
// tops the picks up to 20 ids exactly as the bot's draft does.
func draftIDs(t *testing.T, seed uint64, choose func([]draftCard) draftCard) []string {
	t.Helper()
	byTier := indexByTier(toDraftablePool())
	rng := newDraftRng(seed)
	picks := make([]draftCard, 0, totalDraftPacks)
	for round := 0; round < totalDraftPacks; round++ {
		offer, err := generateOffer(byTier, picks, rng.Float64)
		if err != nil {
			t.Fatalf("seed %d round %d: %v", seed, round, err)
		}
		picks = append(picks, choose(offer))
	}
	return topUpToDeckSize(picks)
}

func cardIDs(deck []engine.Card) []string {
	ids := make([]string, len(deck))
	for i, c := range deck {
		ids[i] = c.ID
	}
	return ids
}

func TestEveryBotDraftIsALegal20TheServerAccepts(t *testing.T) {
	useRealCards(t)

	for seed := uint64(1_000_000); seed < 1_000_000+draftSeeds; seed++ {
		deck, err := draftBotDeck(seed)
		if err != nil {
			t.Fatalf("seed %d: %v", seed, err)
		}
		if len(deck) != draftDeckSize {
			t.Fatalf("seed %d: bot drafted %d cards, want %d", seed, len(deck), draftDeckSize)
		}
		if _, serr := startGame(newReq{You: cardIDs(deck), Seed: &seed}); serr != nil {
			t.Fatalf("seed %d: /api/new refused the bot's drafted deck: %v", seed, serr)
		}
	}
}

// A human picks any card from each offer. Drafts that end with no Basic are
// skipped: a human adds Basics at deckbuild (free under-evolutions), which is
// outside the draft this test plays.
func TestEveryHumanDraftIsALegal20TheServerAccepts(t *testing.T) {
	useRealCards(t)

	played := 0
	for seed := uint64(1); seed <= draftSeeds; seed++ {
		pickRng := newDraftRng(seed ^ 0xabcdef)
		hasBasic := false
		ids := draftIDs(t, seed, func(offer []draftCard) draftCard {
			c := offer[pickRng.IntN(len(offer))]
			hasBasic = hasBasic || c.Stage == "Basic"
			return c
		})
		if !hasBasic {
			continue
		}
		played++
		deck, err := deckFromIDs(ids)
		if err != nil {
			t.Fatalf("seed %d: %v", seed, err)
		}
		if len(deck) != draftDeckSize {
			t.Fatalf("seed %d: drafted deck plays as %d cards, want %d", seed, len(deck), draftDeckSize)
		}
		if _, serr := startGame(newReq{You: ids, Seed: &seed}); serr != nil {
			t.Fatalf("seed %d: /api/new refused a drafted deck: %v", seed, serr)
		}
	}
	if played < draftSeeds*9/10 {
		t.Fatalf("only %d of %d drafts had a Basic", played, draftSeeds)
	}
}

// The SPA draws its packs from the same pool by reading draftPoolFile. Set
// PD_WRITE_DRAFT_POOL=1 to regenerate it after the pool changes.
func TestDraftPoolFileMatchesTheEngine(t *testing.T) {
	useRealCards(t)
	want := draftPoolFileContents()

	if os.Getenv("PD_WRITE_DRAFT_POOL") == "1" {
		b, _ := json.MarshalIndent(want, "", "  ")
		if err := os.WriteFile(draftPoolFile, append(b, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	b, err := os.ReadFile(draftPoolFile)
	if err != nil {
		t.Fatalf("%v (regenerate with PD_WRITE_DRAFT_POOL=1 go test -run TestDraftPoolFile)", err)
	}
	var got draftPoolFileJSON
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got.OnlyWorkingText != want.OnlyWorkingText || !slices.Equal(got.IDs, want.IDs) || !slices.Equal(got.Working, want.Working) {
		t.Fatalf("%s is stale (regenerate with PD_WRITE_DRAFT_POOL=1 go test -run TestDraftPoolFile)", draftPoolFile)
	}
}

// Deckbuild offers only cards the engine can play (owner 2026-10-19 (1)). The
// SPA reads that set from the pool file's working list, which must be every
// card whose text works (under-evolutions and universals included) and
// nothing else.
func TestDraftPoolFileListsEveryWorkingCard(t *testing.T) {
	useRealCards(t)

	working := draftPoolFileContents().Working

	inList := map[string]bool{}
	for _, id := range working {
		inList[id] = true
	}
	for id, rc := range rawByID {
		if textWorks(rc) != inList[id] {
			t.Fatalf("%s (%s): textWorks %v, in working list %v", id, rc.Name, textWorks(rc), inList[id])
		}
	}
	if !slices.IsSorted(working) {
		t.Fatal("working list is not sorted")
	}
}

func abilityWording(rc rawCard) string {
	if rc.Ability == nil {
		return ""
	}
	return rc.Ability.Effect
}
