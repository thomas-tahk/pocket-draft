package engine

// flipCoin flips one coin for player p from the seeded rng — never a second
// RNG source, so folding the same log reproduces every flip (ADR-0005) — and
// records it on the move being applied (spec F11). first marks the first flip
// of an effect: it waits for p's tap unless p is the bot. Inside an attack's
// effect use EffectContext.Flip, which tracks first for you.
func (g *Game) flipCoin(p int, first bool) bool {
	return g.recordFlip(p, g.rng.intn(2) == 0, first)
}

// recordFlip records a flip of p's that landed heads or not.
func (g *Game) recordFlip(p int, heads, first bool) bool {
	g.flips = append(g.flips, CoinFlip{Flipper: p, Heads: heads, Waits: first && !g.S.Mode.isBot(p)})
	return heads
}

// isBot reports whether player p is the bot: P2 in a game against the bot.
func (m Mode) isBot(p int) bool { return m == VsBot && p == 1 }
