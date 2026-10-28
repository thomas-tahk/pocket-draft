package main

import (
	"reflect"
	"sort"
	"sync"
)

// Type support for draft offers: which Trainers/Items help one type's Pokémon,
// so leaning a type raises their odds too (owner, 2026-10-25, point 3).
//
// The card text decides (owner 2026-10-28, decision 2): a Trainer supports
// the types its text names ("Water Pokémon", "[W]", ...) plus the
// non-Colorless card types of the Pokémon its text names (matched by
// textNames, looked up across every card in cards.json, since many named
// Pokémon are not in the draft pool). Only when the text gives neither does
// typeSupportTable decide.
//
// This Go code is the source of truth: draftPoolFileContents writes every
// pool Trainer's answer into public/data/draft-pool.json ("supportTypes"),
// which src/draft/weights.ts reads. src/draft/typeSupport.ts keeps the same
// fallback table and exclusions (TestTypeSupportTSTableMatchesGo).
//
// It is NOT the niche Trainer <-> Pokémon table (#126): that one pairs a
// Trainer with specific Pokémon; this one only says which type it favours.

// typeSupportTable maps a Trainer's exact name to the types it helps, for
// Trainers whose text names no type and no typed Pokémon. Hand-made.
var typeSupportTable = map[string][]string{
	"Drayden": {"Dragon"}, // names the Draco Meteor attack, not a Pokémon
}

// notTypeSupport lists Trainers whose text names a type but works against it,
// so the text rule must not count them.
var notTypeSupport = map[string]bool{
	"Squirt Bottle": true, // discards a [R] Energy from the OPPONENT's Pokémon
}

// supportTypes is the sorted types a Trainer/Item supports; Pokémon and
// notTypeSupport Trainers support nothing.
func supportTypes(c draftCard) []string {
	if c.CardType != "Trainer" || notTypeSupport[c.Name] {
		return nil
	}
	return currentSupportIndex().typesOf(c)
}

// supportTypesFromText is what the text alone says: the types it names plus
// the types of the Pokémon it names, sorted. No fallback, no exclusions.
func supportTypesFromText(c draftCard) []string {
	return currentSupportIndex().fromText(c)
}

// supportIndex is the Pokémon names and types of one rawByID, plus the
// answers computed over it, by card id.
type supportIndex struct {
	cards map[string]rawCard // the rawByID it was built from
	size  int
	names []string            // every Pokémon name, sorted
	types map[string][]string // Pokémon name -> its non-Colorless card types
	mu    sync.Mutex
	byID  map[string][]string
}

var (
	supportMu  sync.Mutex
	supportIdx *supportIndex
)

// currentSupportIndex returns the index over today's rawByID, rebuilding it
// when tests swap or reload the card data.
func currentSupportIndex() *supportIndex {
	supportMu.Lock()
	defer supportMu.Unlock()
	if supportIdx != nil && supportIdx.size == len(rawByID) &&
		reflect.ValueOf(supportIdx.cards).UnsafePointer() == reflect.ValueOf(rawByID).UnsafePointer() {
		return supportIdx
	}
	idx := &supportIndex{cards: rawByID, size: len(rawByID), types: map[string][]string{}, byID: map[string][]string{}}
	seen := map[string]map[string]bool{}
	for _, rc := range rawByID {
		if rc.CardType == "Trainer" {
			continue
		}
		if seen[rc.Name] == nil {
			seen[rc.Name] = map[string]bool{}
			idx.names = append(idx.names, rc.Name)
		}
		if rc.CardType != "Colorless" && !seen[rc.Name][rc.CardType] {
			seen[rc.Name][rc.CardType] = true
			idx.types[rc.Name] = append(idx.types[rc.Name], rc.CardType)
		}
	}
	sort.Strings(idx.names)
	supportIdx = idx
	return idx
}

func (idx *supportIndex) fromText(c draftCard) []string {
	set := map[string]bool{}
	for _, typ := range namedTypes(c) {
		set[typ] = true
	}
	for _, name := range idx.names {
		if textNames(c.TrainerText, name) {
			for _, typ := range idx.types[name] {
				set[typ] = true
			}
		}
	}
	out := make([]string, 0, len(set))
	for typ := range set {
		out = append(out, typ)
	}
	sort.Strings(out)
	return out
}

func (idx *supportIndex) typesOf(c draftCard) []string {
	idx.mu.Lock()
	cached, ok := idx.byID[c.ID]
	idx.mu.Unlock()
	if ok {
		return cached
	}
	out := idx.fromText(c)
	if len(out) == 0 {
		out = append(out, typeSupportTable[c.Name]...)
		sort.Strings(out)
	}
	idx.mu.Lock()
	idx.byID[c.ID] = out
	idx.mu.Unlock()
	return out
}

// supportTypesByName is what the pool file carries: every pool Trainer that
// supports a type, by name, with its sorted types.
func supportTypesByName(pool []draftCard) map[string][]string {
	out := map[string][]string{}
	for _, c := range pool {
		if types := supportTypes(c); len(types) > 0 {
			out[c.Name] = types
		}
	}
	return out
}
