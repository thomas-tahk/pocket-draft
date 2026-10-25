# Draft rework: 20 picks, no free Trainers, offers that fit, a real Auto-draft

**Done-gate:** "On my phone I tap Auto-draft and get a 20-card deck with at most 3 attack types and 2 copies of its key cards; a manual 20-pick draft mostly offers cards that fit what I already picked."

**Status:** passes on the production build (the WebAssembly engine the live site uses, run locally). You haven't tried it yet.

- **Auto-draft:** run 8 times at phone size (390×844, touch) and desktop size (1440×900). Every deck had 20 cards and 1–2 attack types. Each had 7–9 cards at 2 copies and 3–4 Trainers it drafted. Every deck started a battle.
- **Manual draft:** 4 scripted 20-pick drafts. In packs 16–20, 64–96% of offers fit what was already picked, against 0–40% in packs 2–5. Pack 1 always had an ex. Deckbuild never showed free Trainers.
- **Tests on the branch:** engine and server Go tests, lint and build all pass, plus 65 browser tests. Server tests take 44 s, the same as before.

## How to try it

Open the PR's preview link (Vercel posts it on the PR) on your phone.

1. Tap **Auto-draft (test)**. Look at the deck: 20 cards, few types, pairs, some Trainers.
2. Go back and do a manual draft: **Start draft**, 20 picks, then shop, then deckbuild.

Watch for:
- **Draft length:** the header now counts "Pack N / 20".
- **Pack 1:** always includes an ex.
- **Later packs:** lean toward your types. Trainers that name your types or your Pokémon show up more.
- **Deckbuild:** no "Free universals" section. Poké Ball and Professor's Research now appear as normal picks.
- **Saved draft:** if your phone had a draft in progress from before, it opens on the start screen. Your History is kept.

## What was built (5 issues, branch `draft-rework`)

| Issue | Built |
|---|---|
| [#106](https://github.com/thomas-tahk/pocket-draft/issues/106) 20 picks | Draft is 20 rounds, for players and the bot |
| [#107](https://github.com/thomas-tahk/pocket-draft/issues/107) No free Trainers | Trainers only come from picks and the shop. Poké Ball and Professor's Research became ordinary draftable cards |
| [#108](https://github.com/thomas-tahk/pocket-draft/issues/108) Offers lean to your types | Type weighting back on, at its old strength. Pack 1 always has an ex |
| [#109](https://github.com/thomas-tahk/pocket-draft/issues/109) Offers that fit your picks | Cards are 3× as likely when a Trainer names one of your Pokémon, or a card's text names another type you're drafting (Misty → Water) |
| [#110](https://github.com/thomas-tahk/pocket-draft/issues/110) Real Auto-draft | Bot and Auto-draft stay on 1–2 types, take pairs and draft Trainers. The bot now benches in every game |

**Fixed along the way:**
- **Stuck saved draft:** a saved 16-pick draft would have opened on a stuck, empty screen. It now resets once and keeps History ([1123657](https://github.com/thomas-tahk/pocket-draft/commit/1123657)).
- **Slow name matching:** matching your Pokémon's names in Trainer text nearly tripled the server test time. It's back to normal ([d40063c](https://github.com/thomas-tahk/pocket-draft/commit/d40063c)).
- **Silent no-result:** a battle action with no log lines sent "no result" and broke the move list. It now sends an empty list ([bbb9746](https://github.com/thomas-tahk/pocket-draft/commit/bbb9746)).

**Known, not fixed:**
- **Run end not enforced:** 3 losses / 7 wins is only written in the docs. The app doesn't end a run; you type wins and losses into History. That's part of the app-flow work (the run screen).
- **Early packs fit loosely:** pack 1 has nothing to fit, and an off-type first pick starts the draft mixed. Late packs fit well.
- **Hover zoom:** hovering a card while drafting opens a full zoom over the offers (your earlier card-zoom complaint). Not in any issue yet.

**Stuck:** none.

## Decisions that need you (most important first)

1. **Copies per pick.** Today any card you pick once can go into your deck at 2 copies. That's how deckbuild already worked, and the bot does the same. One line in `docs/rules.md` reads as if one pick = one copy. Keep 2 (current), or change to 1?
2. **Rule for offers that fit.** Built: Trainers naming your Pokémon, and cards whose text names another type you're drafting, are 3× as likely. A card naming its own type doesn't count. An "evolution line" rule was dropped, because the draft only offers final evolutions. Keep it, tune it, or change what counts?
3. **Type lean strength.** After 8 Fire picks, Fire shows up about 6× as often as with no lean. That's the strength from before it was switched off in May. Keep it or soften it after you play?
4. **Poké Ball and Professor's Research** are now normal draftable cards. Without that, nobody could ever get them. Keep, or pull them from the game?
5. **Merge:** the [PR](#) from `draft-rework` into `main` deploys it. Your call.

Every other interpretation is in [`docs/questions/draft-rework.md`](../questions/draft-rework.md).
