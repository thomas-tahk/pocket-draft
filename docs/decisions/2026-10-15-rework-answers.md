# Rework answers — 2026-10-15

The owner's answers to a question round on reworking the battle and the draft.
Rules marked here are the owner's statement of how Pokémon TCG Pocket works; they
override anything the engine or older docs say. The battle spec is written from
this file.

## Battle: approach

- **Keep the Go rules engine; rebuild the battle screen from scratch.** Engine rule
  bugs get fixed one at a time. The current screen (a button per action) is
  discarded.
- **Source of truth:** a short battle spec, corrected by the owner, plus screenshots
  of key moments in Pocket.
- **Devices:** phone and desktop equally from day one. It should feel native to play
  on either, even if that costs extra work.
- **Input:** tap a card, legal spots light up, tap a spot — as in Pocket. Some cards
  have no spot to pick (most Trainers, some Items, abilities); the spec must cover
  how those are played.
- **Layout:** copy Pocket's layout (opponent top, you bottom, energy zone beside your
  Active, hand along the bottom).
- **Motion:** short functional motion only, under 150ms.
- **Opponent's turn:** their actions play out one at a time (~0.5s each), plus a
  small action log like Hearthstone's. Skipping only makes sense against the bot,
  never in PvP.
- **Bot / debug buttons:** removed from the UI. The bot takes its own turns.
- **Undo:** no. Stepping through turns belongs to replays.
- **Confirm prompts:** ask before ending a turn while skipping an energy attach, and
  while skipping an attack you could make (Pocket does both, at least against bots).
- **Always visible:** energy on each Pokémon as type icons, remaining HP, status
  icons, both players' points, this turn's and next turn's energy. Pocket does not
  show "Supporter/energy/retreat used this turn" — low priority here.
- **Card text:** long-press (hover on desktop) shows the full card; release closes it.

### Observed problems in the current battle

"All of it pretty much." The owner's side worked somewhat better; the bot often took
no action on its turn. Biggest single problem: **ending a game** — by normal rules,
by conceding, or otherwise.

## Battle: rules (owner-confirmed)

Correct as stated:

- Opening hand is 5 cards and always contains at least one Basic Pokémon.
- A coin flip decides who goes first.
- Every turn starts by drawing 1 card, including the first player's first turn.
- The player going first gets no energy on their first turn.
- The Energy Zone shows this turn's energy and the next turn's.
- You choose up to 3 energy types for a deck, and the zone rolls among them.
- One energy attach per turn, onto your Active or a benched Pokémon.
- The bench holds up to 3 Pokémon.
- No evolving on either player's own first turn, or on the turn a Pokémon was
  played or evolved.
- One Supporter per turn. Items are unlimited. One Tool per Pokémon.
- Retreat once per turn, discarding energy equal to the retreat cost.
- Attacking ends your turn.
- Weakness adds 20 damage. There is no Resistance.
- If your deck runs out you simply stop drawing. It is not a loss.

Corrected by the owner:

- **Supporters on turn 1:** one Supporter per turn applies from the very first turn,
  including the player going first.
- **Points and ties:** a KO scores 1, an ex 2, a Mega ex 3; first to 3 wins; you also
  win if the opponent has no Pokémon left in play. **If both players reach 3 or more
  at the same time** (one attack, an effect, or any other simultaneous resolution),
  **the game is a tie.** The order in which the engine resolves simultaneous events
  matters here; the real game may itself have bugs in this area.
- **Special conditions** apply only to the Active Pokémon and resolve during the
  Pokémon Checkup at the end of each player's turn. There are five:
  - Poison and Burn deal damage. They stack with each other and with one other
    condition.
  - Sleep, Paralysis and Confusion block actions and are mutually exclusive: a new one
    from this group replaces the existing one.
  - At most three at once (e.g. Poison, Burn and Sleep).
  - Cured by: evolving; retreating to the Bench (Sleep and Paralysis prevent normal
    retreat); the Lum Berry Tool; Koga (forced retreat for Weezing or Muk).
  - Sleep and Burn: coin flip at Checkup, heads cures.
  - Paralysis clears automatically after the owner's next turn.
  - Confusion: coin flip before attacking; tails makes the attack fail but does not
    cure the condition.
- **Hand size:** maximum 10. Draws and other additions to the hand beyond 10 are
  prevented — nothing is discarded or burned from the deck.
- **Timers** (competitive and ranked games at least): 90 seconds per turn; letting it
  run out without acting halves it to 45 seconds (and usually removes the player).
  Each player's total action time cannot exceed 20 minutes; past that, points decide
  the winner, equal points is a tie. **Turn cap:** turn 1 is the first player's
  first turn, +1 per turn; the cap is 30 against a player and 50 against the bot.

## Features after the battle works

Most wanted first: **different games every time** (fix the fixed seed). Everything on
the list matters eventually: replays, playing on from a replay, a seen-cards tracker,
draw odds, a constructed deck builder, PvP rooms. Per-deck win/loss stats only make
sense for ranked/constructed; a draft deck needs its own run record, Arena-style.

## Draft

- **Card zoom:** stays up too long or is hard to dismiss; on phone it covers the
  screen badly or the page scrolls behind it.
- **"Too random":** picks spread across too many types, few cards work together,
  nothing strong to build around.
- **Fix with:** type weighting back on, synergy-aware offers, and the guaranteed ex in
  pack 1 restored. A hand-picked pool is not wanted now (the pool keeps growing past
  what anyone can curate) but may be the basis for "Seasons" later.
- **Numbers:**
  - Draft picks: 16 → **20**, so it feels like you pick the cards in your deck.
  - Free universal Trainers: 2 → **0**. A ruleset where a few cards are open to
    everyone may come later; explore the opposite direction first.
  - Run ends at: 3 losses / 9 wins → **3 losses / 7 wins**.
  - Shop tickets (4), cards per pick (5) and copy cap (2) unchanged.
- **Lower evolutions stay free:** drafting a final evolution still makes its lower
  stages available at deckbuild.

## Order of work

**The battle screen first.**

Done-gate: *On my phone, I play a full game against the bot by tapping cards onto
the board. The only buttons I use are attack (picking the attack on the card, not a
UI button) and End Turn, and every rule confirmed above is followed.*

## Housekeeping done with this file

The September vision rewrite was committed. The chief automation (label-triggered
cloud runs, `STATUS.md`, `DECISIONS.md`) was removed. PR #16 (Illumise effect) was
closed unmerged, because Trainers and the effect design come before more attack-only
effects. Stale issues #2, #5, #6 and #13 were closed.
