# Battle spec answers — 2026-10-18

The owner's answers to the battle spec v1 (`docs/battle-spec.html`), verbatim.
Each entry quotes the spec row, then the owner's answer. These override the spec where they differ.

## S2

> Spec: You choose up to 3 energy types for a deck; the Energy Zone rolls among them.

Yes, I'd say energy choice need to be locked in by end of drafting (final build of deck), but energy should default to energy requirements of pokemon in deck with opportunity to edit -- most decks should be biased towards mono-energy unless they build around certain multi-energy decks like dragons (dragonite, mega rayquaza, etc.)

## S4

> Spec: The coin flip happens and is shown first; then both players set up.

Yes the coin flip happens as you say. 
In the game, a designated player's coin is flipped (randomized) - the coin is mostly an opportunity for the game to have player show off their particular coin cosmetic but functionally it comes into play for 2 categories: 1 for deciding who goes first (the player who flipped at start gets Heads > they go first. If tails, the other player goes first) and for Heads/Tails determination in gameplay (some status recovery, some pokemon attacks and abilities, some Trainer/support cards, etc. ) Each time coin is used, it should show. Pocket typically has user do just first flip manually and if there are more than 1 flip to do, it automates the rest (displays it in a grid 3 coins wide and as many rows it needs to finish other rolls, I believe there may be a maximum but I don't know that number. You could probably research and find that number)

## S5

> Spec: Both players place their Active and up to 3 Bench at the same time, face down, and they are revealed together.

Largely true. The setup happens at same time, no taking turns for setting up bench before start of game. and when both players confirm they are done setting up (each player would press a button/confirmation in-game, I do it without thinking so I forget what it even looks like)

## T7

> Spec: The Bench holds 3. Basics can be played to it any number of times during your turn.

I'd be careful about calling it 'can be played to it any number of times' during your turn. Or at least clarify/reword 
You can't play other basics on top of pre-existing basics on each bench slot. Only one pokemon will occupy any given slot. 
It also matters which bench slot player decides to place for some game mechanics. 
Bench slots are physically placed behind the active spot. 
1 card active slot facing opponent
3 cards bench slots behind. 
> 1 card directly behind active slot, and a card slot each to left and right of this.

## T8

> Spec: No evolving on either player's own first turn, or on the turn the Pokémon was played or evolved.

Mostly true but exceptions can come from pokemon ability - only one I know of at the moment that can get around this is one of the eevee card's abilities that allows it to evolve the turn it is played and in the active spot. Not all eevees do this. It is a specific eevee cared with this specific ability, I just forget what it's called but I can pinpoint it if you need me to

## T9

> Spec: Retreat once per turn, discarding energy equal to the retreat cost.

yes for multiple choices for energy to discard from, player gets to choose. ordering of how energy was attached should not matter for what player decides to discard when retreating.

## T11

> Spec: Weakness adds 20 damage. There is no Resistance.

some pokemon lack weakness modifier by default
there are cards to increase damage, increase weakness bonus (I think?), and there's also a stadium card that can turn weakness from +20 to x2. (I believe bounded field is the name?) And there are possibly other exceptions I don't know about.

## A3

> Spec: A played Item or Supporter goes to the discard pile. A Tool stays on its Pokémon and is discarded with it.

tools can be removed by certain cards (field blower and guzma comes to mind, mainly). If removed, the player of that pokemon can re-attach a different tool

## A4

> Spec: Abilities (451 cards): an ability you activate can be used once per turn per Pokémon unless its text says otherwise; a passive one is always on. Is that right?

That is probably mostly right but this is where pokemon can get complex and run into all kind of niche and edge cases that even devs probably haven't thought very well through. But let's work with this principle in general. I think card wording often makes it try to clear it's once per turn and ones that are not labeled that way are assumed to be able to be utilized multiple times or triggered passively.

## A5

> Spec: Fossil Items are played to the Bench like a Basic.

Yes but important distinction for fossils - you can always choose to discard this without losing a point on your turn. Fossils also do not count as a basic card on opening hand - (you can't just start with fossil card as your leading basic or as pokemon to place on bench in the initial pokemon placement before game starts. fossil is an item until it's placed on your own turn)

## C1

> Spec: Conditions apply only to the Active and resolve at the Pokémon Checkup at the end of each player's turn.

I don't know of an exception to this. I think one of the best written up rules for pokemon pocket conditions are listed here in link below:
https://game8.co/games/Pokemon-TCG-Pocket/archives/483132

## C2

> Spec: Poison deals 10, Burn deals 20 at each Checkup.

yes at base level assuming nothing is modified or impeded by other card effects, this is true.

## C3

> Spec: Poison and Burn stack with each other and with one other condition. At most 3 at once.

I didn't know about this before but apparently that seems to be the case. I can't recall seeing more than 3 either including poison and burn.

## C9

> Spec: Cured by evolving, by moving to the Bench, by Lum Berry, and by Koga.

yes anything that leaves active spot, evolving, lum berry cures conditions. 
You can also prevent conditions, notably found in Arceus, Ogerpon Ex, and Steel Apron (for steel types only, tool)

## E4

> Spec: Simultaneous resolution: apply all damage from one event, then remove every knocked-out Pokémon, then award all points, then check for a win or tie.

the wording sounds good but there are some rules I'm not aware of and how sequencing works. 
For example, if a pokemon that damages itself upon attack like charizard Y Mega Ex were to deal damage, knockout opponent pokemon for last point needed but itself also dies from its own attack effect + something like opponent's rocky helmet. This would result in a tie. 
However, if the charizard would have died from poison or burn after the attack, the win would be awarded to the charizard player. 
I'm sure there's more niche cases I could not possibly fathom but this is one of the few I'm aware of.

## E5

> Spec: After a knockout, its owner picks the new Active from the Bench before play continues. If both players must pick, who picks first, and does either see the other's pick?

whoever's turn is next chooses first if there is a double knockout scenario that has not ended the game yet. opponent sees the pick once the pokemon has been moved from bench to active

## M1

> Spec: 90 seconds per turn. Letting it run out without acting halves your next turn to 45 seconds.

yes when it goes to 0, it just move on to next player's turn. 
I'm not sure about when exactly player is being 'removed' after timeouts.
I've seen in go multiple rounds but lately I think I see it around 2-3 turns for the same player, tops. but maybe that's because the game recognized the other player has disconnected

## M3

> Spec: Turn cap: turn 1 is the first player's first turn, +1 per turn. Cap is 30 against a player, 50 against the bot.

just to clarify, each player's turn add +1 to the turn count. So effectively, each player gets 15 turns in a competitive match against other players. I believe when max turns are reached (so when player ends Turn 30) I believe it ends in a tie regardless of score

## L1

> Spec: Copy Pocket's layout: opponent at the top, you at the bottom, Actives facing each other in the middle, Benches behind them, Energy Zone beside your Active, hand along the bottom edge.

the mockup you have on the right shows End Turn between your board and hand - that should be something that should be towards middle of screen. 
I also don't see a visual representation of your deck and the discard pile below it - for both yourself and the opponent. In the actual game it is off to the right a little bit away from bench and active spot. it's mirrored for opponent.
For Energy and next turn energy indicators, those are placed to the right of hand and a little below the bench. Current turn's energy has a bigger visual indicator and when it has been attached, it leaves an empty grey circular area, which is filled with the next turn's energy by start of your next turn.

## L3

> Spec: Always visible: both players' points, and this turn's and next turn's energy.

points should be visualized via pips or maybe even a little pokeball icon to fill up to 3

## L5

> Spec: Whose turn it is, the turn number, and the turn timer sit between the two halves.

whose turn it is should be obvious by some visual indicators, not a midde of game board persistent indicator. When it's your turn, there's a flash of graphics (or a banner) that says YOUR TURN and same for enemy's turn Opponent's TURN with color differences. In addition, when it's your turn, there's a pulsating aura for cards that can be used. We don't need the same exact effect but similar visual indicators would be good to have

## F1

> Spec: Tap a card: it lifts, the spots it can legally go glow, and the rest dims. Tap a glowing spot to play it. Tap anywhere else, or the card again, to cancel.

you hold cards to make them do anything and drop them somewhere on board (sometime you have to target them properly)  for the cards to play.  can't queue cards the way some card games do. just holding press on card in hand will show a blown up version to allow you to look at card art and read card text (takes up most of the screen, center aligned vertically and horizontally). and when you start dragging it towards board, it will minify to normal size. 
Maybe on computer, tapping as you say can work reasonably well

## F3

> Spec: Setup: tap a Basic and the Active spot glows; after that the Bench spots glow. Finish with a Ready button. Is Ready allowed, given the done-gate names only attack and End Turn?

Yes there's something like a Done or Ready button at the initial active/bench setup before game begins in earnest.
The order is - randomized player designated for coin flip > coin flip result > announce whether you go first or second > active/bench setup simultaneously (this view would also start showing your and opponent's first energy) > both players Ready/Done > Game starts

## F6

> Spec: Attach energy: tap this turn's energy in the zone, all your Pokémon glow, tap one. (Pocket uses a drag; tap-tap keeps one input model.)

although pocket game itself makes player drag and drop, that may be more efficient way to do it for both mobile and PC. Though this may take some 'unlearning' for mobile only players. Then this system should make use of a UI change where everything dims a little, there's a message at the bottom saying something like 'choose pokemon to attach energy to'

## F8

> Spec: Trainer with no target (Professor's Research, Poké Ball, Red Card, Giovanni): tap it and the middle of the board glows as a play area; tap that. Or would you rather tap once to enlarge the card and press Play?

middle of the board being where exactly?  it wouldn't be bad if the entire board (excluding hands) were to glow in general. 
The alternative you've brought up wouldn't be that bad either though it'd feel quite a bit different from Pocket

## F9

> Spec: A Trainer that needs a choice after it is played (pick from your deck; Sabrina's new Active) opens a picker over the board for whoever owns the choice.

I'm not entirely sure what you mean but pocket handles this sort of thing by dimming the screen except for the cards directly involved in the choice and prompting the appropriate user to make a decision. If there's only one choice possible, it's just automatically carried out.

## F16

> Spec: A small action log, Hearthstone style.

Not sure if I want text unless it's on desktop view. 
It should be a little pokeball notebook icon or something on mobile to tap on and examine which cards were used recently, and maybe can be scrolled
Desktop view should also leave this clickable but probably have it easier to see more information all at once. 
(but one view shouldn't provide an actual difference in the information provided over the other, just the presentation)

## F17

> Spec: Motion is functional and under 150 ms (opponent pacing aside). No undo.

what is this about motion? I don't know what you are referring to

## F18

> Spec: An illegal move is never sent: a card with nowhere legal to go shows nothing glowing and a one-line reason.

reason being something that flashes by briefly on the screen when the user tries to make motions/actions that seem to want to use said card? seems good in that case.

## Follow-up answers (2026-10-18, in chat)

- **Card text coverage:** "then it's fucking time to build them in." Every card's text gets implemented; no permanent cut-down pool.
- **Simultaneous 3+ points:** tie, even though Game8/Dexerto say the player with a Benched Pokémon wins. "seems stupid if 3+ points are scored on both sides by the same action/attack. If it's under 3 points for both, then the bench rule makes sense." (Spec E3, E5.)
- **Unused energy:** "It displays into enemy's turn but is discarded by your next turn." It counts as discarded energy. (Spec T6.)
- **Orange up-arrows:** "simple visual indicators for damage modifiers in place, from the stadium card as well as lucario on bench in this case." Tapping one "will show where the effects are coming from and the numeric values". (Spec L6.)
