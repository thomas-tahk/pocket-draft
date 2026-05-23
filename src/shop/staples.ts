// Fixed staple-shelf: 12 high-impact Trainers always available at shop, regardless
// of draft. Per rules.md, "ensures no player is bricked without key Supporters due
// to bad draft luck." Same contents every session — tunable list.
export const STAPLE_SHELF_IDS: ReadonlyArray<string> = [
  'A2-150',   // Cyrus       — Supporter, move opponent damage to active
  'A1-225',   // Sabrina     — Supporter, force opponent switch
  'A3a-064',  // Repel       — Item, force opponent's active switch
  'B1-225',   // Copycat     — Supporter, copy opponent's last Supporter
  'A1-223',   // Giovanni    — Supporter, +10 damage this turn
  'A2-146',   // Pokémon Communication — Item, search for Pokémon
  'A1-219',   // Erika       — Supporter, heal Grass
  'A2-155',   // Mars        — Supporter, hand disruption
  'B2a-089',  // Iono        — Supporter, shuffle-redraw
  'A2-152',   // Cynthia     — Supporter, signature damage boost
  'A2b-070',  // Pokémon Center Lady — Supporter, heal + status clear
  'A3-155',   // Lillie      — Supporter, draw
];
