# M4 review answers (owner, 2026-10-28)

Answers to the decisions in `docs/milestones/M4.md` and the three bold-look
directions in `docs/chief/look/m4-bold/` (#131). Built on `m4-draft-runs`
(PR #129) unless marked "next milestone".

## Look (#131): get the best of all 3
- **B (Holo Night):** keep its shine / foil lighting treatment on cards and highlights.
- **Dark background:** liked, but it must be an option, not forced. Light stays
  available. Follow the phone setting, plus a manual toggle.
- **A (Pack Burst):** its comic font and presentation read better than C and are fine
  for the project, but it is not the definitive look.
- **C (Type Colour):** its run-info screen presentation, in general.
- **Build in M4:** one blended look across every non-battle screen (home, draft, shop,
  deckbuild, run screen, end of run, past runs):
  - C's structure and run-screen layout;
  - A's readability: heavier display type, strong outlines and contrast, not full comic;
  - B's foil / holo shine on cards and highlights, off under reduced motion;
  - light and dark themes with a toggle.
- Token-driven (`src/tokens.css`). Saturated, high contrast, WCAG AA.
- Screenshots at 390×844, light and dark, in `docs/chief/look/m4-blend/` and in a
  PR #129 comment.

## Decisions to build in M4
1. **Type lean:** keep the strength, add a cap so off-type cards (e.g. Dragons) never
   become near-impossible. The chief picks the cap and reports it with numbers.
2. **Type-support Trainers:** the card text decides. The hand list is only a fallback
   for cards whose text does not make it obvious.
3. **Ancient / Future:** a card counts only if the card itself says Ancient / Future.
   Remove the by-name rule for Koraidon / Miraidon.
4. **Eevee ex and Eeveelutions** raise each other's odds, both directions.
5. **Eevee ex rule question:** badly framed (all Eeveelutions are Stage 1; Rare Candy is
   Basic to Stage 2 only). Dropped; follow the card text. If Alolan Muk's ability text
   switches off Eevee ex's evolve ability, follow the card text and note it in one line.
6. **Energy Zone picker:** far less wordy. Label like "Choose up to 3 energies to use",
   default selection from the deck's contents (attack costs). Cut the extra lines.
7. **Battle end-screen button size:** no opinion; leave as built.

## Draft design: next milestone, not M4
Filed as GitHub issues in the next milestone, with the owner's words and evidence:
- (a) 20 picks give too many uninteresting picks: the owner feels 35–60% of picks are
  not even contenders for the deck. Maybe each pick should be 1 copy of the card again.
- (b) A build-around "god pack" pick at the start that ignores type lean.
- (c) Mega ex never appears: implement Mega ex rules and texts so they enter the pool.
- (d) Shop Trainers / Items: Stage 2 strategies are crippled without Rare Candy and
  search / tutor cards.
- (e) Trainer / support lean may be a bit low; do not change without evidence.
