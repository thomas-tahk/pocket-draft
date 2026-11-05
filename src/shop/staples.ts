// Fixed staple-shelf: 13 high-impact generic Trainers. Always available at shop
// regardless of draft, except deduped against what the player already drafted
// (since drafted Trainers are already in collection at 2x cap).
//
// Per rules.md §2, the shop "ensures no player is bricked without key Supporters
// due to bad draft luck." Tunable list — currently the 12 from the previous shop
// iteration plus Rare Candy. Future curated additions belong in this file.
export const STAPLE_SHELF_IDS: ReadonlyArray<string> = [
  'A2-150',   // Cyrus
  'A1-225',   // Sabrina
  'A3a-064',  // Repel
  'B1-225',   // Copycat
  'A1-223',   // Giovanni
  'A2-146',   // Pokémon Communication
  'A1-219',   // Erika
  'A2-155',   // Mars
  'B2a-089',  // Iono
  'A2-152',   // Cynthia
  'A2b-070',  // Pokémon Center Lady
  'A3-155',   // Lillie
  'A3-144',   // Rare Candy (owner request, #137)
];
