package engine

// SearchDeck puts a random card matching Match from the acting player's deck
// into their hand (spec A7: there is no deck search, only a random pick). It
// can't be played with 10 cards in hand (T3). With no match it does nothing.
// E.g. Poké Ball: SearchDeck{Match: Card.IsBasicPokemon}.
type SearchDeck struct {
	Match func(Card) bool
}

func (SearchDeck) CanPlay(ctx *EffectContext) error { return handFull(ctx) }

func (s SearchDeck) Apply(ctx *EffectContext) {
	pl := ctx.Player()
	var matches []int
	for i, c := range pl.Deck {
		if s.Match(c) {
			matches = append(matches, i)
		}
	}
	if len(matches) == 0 || len(pl.Hand) >= handMax {
		ctx.g.narrate("%s found nothing in the deck.", side(ctx.attacker))
		return
	}
	i := matches[ctx.g.rng.intn(len(matches))]
	pl.Hand = append(pl.Hand, pl.Deck[i])
	pl.Deck = removeAt(pl.Deck, i)
	ctx.g.narrate("%s put a card from the deck into their hand.", side(ctx.attacker))
}
