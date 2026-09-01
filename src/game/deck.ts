// Flattens a finalized deck (cardId → count) into a flat list of card ids, the
// shape the sim's /api/new expects.
export function expandDeck(deck: Record<string, number>): string[] {
  const ids: string[] = [];
  for (const [id, count] of Object.entries(deck)) {
    for (let i = 0; i < count; i++) ids.push(id);
  }
  return ids;
}
