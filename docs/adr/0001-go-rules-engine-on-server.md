# Rules engine in Go, running server-side

We will write the gameplay rules engine (the simulator core) in **Go**, running on a server. The browser is a thin client that sends player *intents* and renders the state the server sends back — it never runs the rules itself. This holds even for local single-machine play, which talks to a Go server on `localhost`.

## Considered Options

- **TypeScript engine in the browser.** Same language as the existing frontend; the first local-play milestone could ship as a pure static page (no backend) and run the same code on a Node server later. Rejected despite being the lower-friction path.
- **Go engine on a server (chosen).** The browser cannot run Go, so even local play requires a running Go backend.

## Why

Two reasons, one practical and one personal:
1. Online lobby matches are a real goal. Because Pokémon TCG Pocket is a hidden-information game (concealed hands, shuffled decks), a networked match needs a neutral server to referee so neither client can cheat by reading the opponent's state. A server-side engine is the correct end-state regardless of language.
2. The maintainer wants hands-on exposure to backend development in Go. This is an explicit learning goal, not a performance requirement — a turn-based card game does not need Go's speed or concurrency.

## Consequences

- The first simulator milestone is **not** a static page like the existing draft tool — it requires a Go backend running (localhost in development, a deployed server in production). Backend complexity is pulled forward to day one rather than deferred until online play.
- No engine will ever be ported between languages: it is Go from the first commit and stays server-side. Updating the rules for new game mechanics is ordinary maintenance, unrelated to this decision.
- The frontend and the Go server communicate over a single intent/state protocol that is identical for local and online play — going online adds matchmaking and lobbies in front of the engine, but does not change the engine.
