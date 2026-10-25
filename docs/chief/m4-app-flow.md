# M4 · App flow and UX — shape notes

Status: Shape, round 1 answered 2026-10-24. Vision of record stays docs/VISION.md.

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

## Parking lot (ideas, not scope)
- Roguelike mode using Pocket cards (deck changes between battles).

## Assumptions (for Research)
- Assumption: the 3L/7W run end needs real code — today W/L are only manual History entries (#106 finding). (verified)
- Assumption: a queue against other people's drafted decks is built on the same "two players in one game" server piece VISION.md calls rooms; matchmaking is a policy on top. (open)
