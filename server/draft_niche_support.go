package main

import (
	"sort"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// Niche support for draft offers (#126): archetype Trainers whose text does
// not name the Pokémon they work with, paired with those Pokémon, so picking
// either raises the other's odds (owner, 2026-10-25, point 2). Mirrored by
// src/draft/nicheSupport.ts; edit both together.
//
// A Trainer that names a Pokémon in its text needs no entry here: the synergy
// rules read the text (rule 1 / rule 4 in draft_synergy.go). This file is only
// for what the text cannot say. Hand-made from public/data/cards.json; the
// owner has not reviewed the groups, so they are guesses.
//
// Names are exact, without " ex": a Pokémon matches by its name minus a
// trailing " ex" (the same stripping rule 1 applies to picked names), so
// "Aerodactyl" covers "Aerodactyl ex" but never a longer name. Pokémon not in
// today's draft pool are listed too: the pool grows as card text lands, and an
// entry with no card in the pool does nothing. Eevee ex is not listed (#128).

// nicheGroup pairs Trainers with the Pokémon they support. Cross-kind only: a
// picked Trainer raises the group's Pokémon and a picked Pokémon raises the
// group's Trainers; two Pokémon, or two Trainers, of one group do not raise
// each other.
type nicheGroup struct {
	label       string
	trainers    []string               // exact Trainer names
	pokemon     []string               // Pokémon names without " ex"
	evolvesFrom []string               // any Pokémon whose card says it evolves from one of these
	fits        func(engine.Card) bool // Pokémon the engine recognises by rule rather than by list
}

var nicheGroups = []nicheGroup{
	// Fossils: each Fossil/Old Amber supports the final evolution of its own line.
	{label: "Helix Fossil", trainers: []string{"Helix Fossil"}, pokemon: []string{"Omastar"}},
	{label: "Dome Fossil", trainers: []string{"Dome Fossil"}, pokemon: []string{"Kabutops"}},
	{label: "Old Amber", trainers: []string{"Old Amber"}, pokemon: []string{"Aerodactyl"}},
	{label: "Skull Fossil", trainers: []string{"Skull Fossil"}, pokemon: []string{"Rampardos"}},
	{label: "Armor Fossil", trainers: []string{"Armor Fossil"}, pokemon: []string{"Bastiodon"}},
	{label: "Plume Fossil", trainers: []string{"Plume Fossil"}, pokemon: []string{"Archeops"}},
	{label: "Cover Fossil", trainers: []string{"Cover Fossil"}, pokemon: []string{"Carracosta"}},
	{label: "Root Fossil", trainers: []string{"Root Fossil"}, pokemon: []string{"Cradily"}},
	{label: "Claw Fossil", trainers: []string{"Claw Fossil"}, pokemon: []string{"Armaldo"}},
	{label: "Jaw Fossil", trainers: []string{"Jaw Fossil"}, pokemon: []string{"Tyrantrum"}},
	{label: "Sail Fossil", trainers: []string{"Sail Fossil"}, pokemon: []string{"Aurorus"}},
	// Eevee Bag's text says "Eevee"; the Pokémon it helps are the ones that
	// evolve from it (Eevee ex itself is #128's).
	{label: "Eeveelutions", trainers: []string{"Eevee Bag"}, evolvesFrom: []string{"Eevee"}},
	// Ancient / Future / Ultra Beast: cards.json has no such tag, so the members
	// are the engine's own lists (texts_3_13.go, texts_3_12.go), not a third
	// copy. TestNicheTSTableMatchesTheEngine keeps the TS table equal to them.
	{
		label:    "Ancient",
		trainers: []string{"Ancient Booster Energy Capsule", "Professor Sada"},
		pokemon:  ancientPokemon,
	},
	{
		label:    "Future",
		trainers: []string{"Future Booster Energy Capsule", "Professor Turo"},
		fits:     func(c engine.Card) bool { return isFuture(c) }, // Miraidon and every "Iron ..."
	},
	{
		label:    "Ultra Beast",
		trainers: []string{"Lusamine", "Beast Wall", "Beastite"},
		pokemon:  ultraBeastNames(),
	},
}

// ultraBeastNames is the engine's Ultra Beast list (texts_3_12.go), sorted.
func ultraBeastNames() []string {
	out := make([]string, 0, len(ultraBeasts))
	for n := range ultraBeasts {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

// Lookup tables over nicheGroups: name -> indexes of the groups it is in.
var (
	nicheByTrainer     = map[string][]int{}
	nicheByPokemon     = map[string][]int{}
	nicheByEvolvesFrom = map[string][]int{}
)

func init() {
	for i, g := range nicheGroups {
		for _, n := range g.trainers {
			nicheByTrainer[n] = append(nicheByTrainer[n], i)
		}
		for _, n := range g.pokemon {
			nicheByPokemon[n] = append(nicheByPokemon[n], i)
		}
		for _, n := range g.evolvesFrom {
			nicheByEvolvesFrom[n] = append(nicheByEvolvesFrom[n], i)
		}
	}
}

// nicheGroupsOf is the niche groups c belongs to: a Trainer by exact name, a
// Pokémon by name minus " ex" or by what it evolves from.
func nicheGroupsOf(c draftCard) []int {
	if c.CardType == "Trainer" {
		return nicheByTrainer[c.Name]
	}
	out := append([]int{}, nicheByPokemon[pokemonBaseName(c.Name)]...)
	for i, g := range nicheGroups {
		if g.fits != nil && g.fits(engine.Card{Name: c.Name}) {
			out = append(out, i)
		}
	}
	if c.EvolvesFrom != "" {
		out = append(out, nicheByEvolvesFrom[c.EvolvesFrom]...)
	}
	return out
}
