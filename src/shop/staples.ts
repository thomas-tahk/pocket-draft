// Fixed staple-shelf: 16 generic Trainers (no Erika/Cynthia: they target a subset of Pokémon). Always available at shop
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
  'A2-155',   // Mars
  'B2a-089',  // Iono
  'A2b-111',  // Poké Ball (owner, M4 PR #129)
  'P-A-007',  // Professor's Research
  'A2-148',   // Rocky Helmet
  'B3-147',   // Field Blower
  'A3-151',   // Guzma
  'A2b-070',  // Pokémon Center Lady
  'A3-155',   // Lillie
  'A3-144',   // Rare Candy (owner request, #137)
];
