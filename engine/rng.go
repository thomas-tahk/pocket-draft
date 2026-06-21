package engine

import "math/rand/v2"

// rng is a seeded, reproducible source. Every "random" decision in the engine
// (shuffle, coin flip, which energy the zone generates) draws from here so that
// re-folding the same log yields the same game (ADR-0005). Never use the global
// rand or any ambient randomness in engine code.
type rng struct {
	r *rand.Rand
}

func newRng(seed uint64) *rng {
	return &rng{r: rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15))}
}

func (g *rng) intn(n int) int { return g.r.IntN(n) }

func (g *rng) shuffle(n int, swap func(i, j int)) { g.r.Shuffle(n, swap) }
