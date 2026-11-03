# M4 unified look: mockups (#131, round 2)

Owner feedback (PR #129, 2026-11-03): the blended look feels like a jigsaw. Holo sheen reads as "special" on
cards but is fine on main buttons. Asked what one look would run from landing page to battle to replay.

**Idea:** the battle board already has a look. Derive everything else from it: a navy corner tab per
screen (the board's Opponent / You labels), pill buttons (the End Turn pill), thin borders, one soft
shadow, the lavender-over-blue split. Nothing app-specific: no ink outlines, no hard shadows, no italic caps.
Sheen only on the main button, never on cards.

| Direction | Mood |
|---|---|
| **board** (`board-all.png`) | The board's own lavender/blue, frosted panels, cyan main button with sheen. Light. |
| **cream** (`cream-all.png`) | Calmest. Pale page, white panels, flat cyan main button, no sheen anywhere. Board unchanged. |
| **night** (`night-all.png`) | The board in indigo for dark theme; same shapes. Pairs with either light direction as the dark theme. |

Each row: Home, Draft, Run, Battle, Replay at 390×844. Cards are stand-ins (art can't load in the cloud).
Battle and Replay show the board with the same tab, pill and border language; the real board layout is not changed.

Regenerate: `node docs/chief/look/m4-unify/render.cjs`. Source: `mockup.html?dir=board|cream|night&screen=...`.
