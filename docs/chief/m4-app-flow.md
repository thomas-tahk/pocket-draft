# M4 · App flow and UX — shape notes

Status: Shape, rounds 1–2 answered 2026-10-24; round 3 open. Vision of record stays docs/VISION.md.

## Pain today (owner, 2026-10-24)
- No home base: after a battle there's no clear "what now", no visible win/loss record.
- No visual unity, theme or transition between draft and battle — they feel like two apps.

## Decided
| # | Decision |
|---|---|
| 1 | Opening screen = pick a mode; show only modes that work (Draft, Practice vs bot). Ranked/Custom hidden until real. |
| 2 | A draft run has a home screen: record (W–L), deck, Next battle. |
| 3 | Deck locked after deckbuild for the whole run. Editing between battles = a different mode. |
| 4 | Bot is the stand-in run opponent for now, labelled. Target later: automated queue that matches you to a random opponent's drafted deck, repeat until 3 losses / 7 wins (Hearthstone Arena feel). Not lobbies. Owner open to alternatives. |
| 5 | Phone and desktop both good; accessible across the board. Phone is the native shape. |

| 6 | Look: borrow mostly from Pokémon TCG Pocket itself. This app = Pocket's battles plus modes and QoL Pocket lacks (review/replays, competitive resources). |
| 7 | Win/loss-count runs (3L/7W) are doubtful: they suit games where rewards hang on the numbers; this is just gameplay, no rewards, no profit. Replacement open (round 3). |
| 8 | A run continues after leaving. Leaving mid-game = loss after a reconnect grace period. |
| 9 | Bot plays a fixed deck; which one doesn't matter. |
| 10 | Replays: a link per battle on the run screen, never forced. |
| 11 | Accessibility (default taken, owner didn't object): keyboard for everything, visible focus, colour never the only signal, 44 px targets, reduced motion; screen reader for menus/screens now, board later. |

## Parking lot (ideas, not scope)
- Roguelike mode using Pocket cards (deck changes between battles).

## Assumptions (for Research)
- Assumption: the 3L/7W run end needs real code — today W/L are only manual History entries (#106 finding). (verified)
- Assumption: a queue against other people's drafted decks is built on the same "two players in one game" server piece VISION.md calls rooms; matchmaking is a policy on top. (open)
