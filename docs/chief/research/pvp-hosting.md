# PvP hosting: where the matchmaking queue and authoritative game live

Researched 2026-10-24. All provider claims fetched 2026-10-24 from the URLs cited. "UNVERIFIED" = page did not give the number.

## Requirements (from the brief)
- Queue pairs two players with drafted decks; server-authoritative, event-sourced (ADR-0005).
- Reconnect within 90 s else forfeit; 2 min no match -> bot.
- Near-zero players at first; part-time owner; no host that sleeps/pauses idle (kills sockets or DB).
- Go modules stay stdlib-only (CLAUDE.md: new dependency = owner decision).

## Repo facts
- `server/serve.go` is stdlib `net/http`: `/api/new`, `/api/state`, `/api/move`, `/api/bot`; no Dockerfile exists; `vercel.json` exists (SPA only).
- Game state is in-process memory on one server. One process = no cross-instance sync problem. Keep it that way.

## Transport without a dependency
Stdlib has no WebSocket server. `golang.org/x/net/websocket` (or gorilla/coder) = a dependency decision.
Stdlib alternative, fits turn-based play and ADR-0005:
- Server -> client: SSE (`text/event-stream`, `http.Flusher`) streaming appended log events. Browser `EventSource` auto-reconnects and sends `Last-Event-ID`: that IS the "re-fold from log tail" reconnect path.
- Client -> server: `POST /api/move` (already exists).
- Queue: `POST /api/queue`, then SSE `matched` event; or long-poll `GET /api/queue/wait`.
- Disconnect detection for the 90 s forfeit: SSE request context cancel + a per-player "last seen" timer; timeout appends a `Forfeit`/`Timeout` event (ADR-0005 allows wall-clock to trigger an event).
- Needs no proxy buffering and a long response timeout (host-dependent, below). Heartbeat comment line every ~15 s.
Verdict: SSE + POST is sufficient for a turn-based card game; WebSocket buys little. No dependency needed.

## Options

| | Fly.io | Railway | Render (paid) | Hetzner VPS | Vercel Functions | Cloudflare DO |
|---|---|---|---|---|---|---|
| $/mo at ~0 players | ~$3.69 (512 MB always-on) | $5 plan (usage inside credit) | UNVERIFIED | EUR 5.99 | $0 (Hobby) | $0 (free) |
| $/mo at ~50 players | ~$3.69 + <$1 egress | ~$5-10 | UNVERIFIED | EUR 5.99 | not viable (see) | ~$0-5 |
| Long-lived conn | Yes (est.; see note) | not stated on page | Yes, no max duration | Yes (you run it) | WS beta; 300 s cap on Hobby | Yes (WS hibernation) |
| Sleep | Only if autostop on; can disable | not stated | Paid: no spin-down stated; Free sleeps | Never | n/a | n/a |
| Deploy effort | Dockerfile + fly.toml | Dockerfile/Nixpacks | Dockerfile | Server admin, TLS, patching | n/a for Go | Rewrite engine in JS/TS |
| Match state | In-process memory | In-process | In-process | In-process | External store needed | DO storage |

### 1. Fly.io (recommended)
- Price: shared-cpu-1x 256 MB $2.19/mo, 512 MB $3.69/mo, billed per second running; egress $0.02/GB NA/EU; inbound free; shared IPv4/IPv6 free, dedicated IPv4 $2/mo; no monthly platform fee, no free tier (trial: 2 h runtime or 7 days). Source: https://docs.fly.io/about/pricing (2026-10-24).
- Sleep: opt-in. `auto_stop_machines = "off"` + `auto_start_machines = false` => Machines never stopped. Source: https://docs.fly.io/launch/autostop-autostart (2026-10-24). Must set this explicitly; autostop would kill sockets.
- Websockets/SSE: that docs page and the fly-proxy reference page do NOT mention WebSockets or long-lived connections; support is UNVERIFIED from fetched docs. Fly Proxy is an ordinary TCP/HTTP proxy and long HTTP streams are widely used there, but confirm with a 10-minute SSE smoke test before committing.
- Size: Go engine + card data in memory; 256 MB may be tight, budget 512 MB. Measure RSS first (UNVERIFIED).
- Deploy: add Dockerfile (multi-stage Go build, both modules, `replace ../engine`), `fly launch`, one Machine, one region. New files, no code change besides SSE endpoints.
- Restart risk: deploys restart the process and drop in-memory games. Mitigation: persist the event log (ADR-0005 makes recovery = re-fold) to a Fly volume (volume pricing not fetched: UNVERIFIED) or accept "deploy kills live games" at this scale.

### 2. Railway
- Hobby: $5/mo, includes $5 usage credit; CPU $0.00000772/vCPU-s (~$20/vCPU-mo), RAM $0.00000386/GB-s (~$10/GB-mo), egress $0.05/GB. Source: https://railway.com/pricing (2026-10-24). Sleep behaviour: page silent (UNVERIFIED). 0.5 GB RAM + light CPU likely ~$5-8/mo total. WebSocket/long connection docs not fetched: UNVERIFIED.
- Comparable to Fly, slightly dearer, simpler UX. Not chosen: less verified on sleep/sockets.

### 3. Render (paid web service)
- Free tier spins down after 15 min no inbound traffic (WS messages count as traffic) and takes ~1 min to wake: REJECTED by constraint. Source: https://render.com/docs/free (2026-10-24).
- Paid: WebSockets supported, "does not enforce a maximum duration", no idle timeout documented, no limit on open connections; connections close on deploy (SIGTERM, 30 s default grace, up to 300 s). Source: https://render.com/docs/websocket.
- Starter spec = 0.5 CPU / 512 MB (https://render.com/docs/compute-plans). Dollar price: NOT on any page fetched (pricing page renders no prices). UNVERIFIED: do not quote.
- Good fit technically; price must be read from the Render dashboard before approving.

### 4. Hetzner Cloud VPS
- CX23: 2 vCPU, 4 GB RAM, 40 GB NVMe, 20 TB traffic, EUR 5.99/mo excl. VAT; CAX11 (Arm) EUR 6.49. Locations: Nuremberg, Helsinki (eu-central). The page flagged those listings "currently unavailable" at fetch time. Source: https://www.hetzner.com/cloud/cost-optimized/ (2026-10-24). EU-only latency for US players (UNVERIFIED impact; ~100+ ms is fine for turn-based).
- Never sleeps; no connection limits beyond what you configure. Most headroom per euro. Cost: owner is the sysadmin (Caddy for TLS, systemd, OS patches, backups). Disk survives restarts so the event log can be a plain file.
- Best if the owner is willing to run a box; worst on part-time maintenance.

### 5. Vercel itself
- Functions WebSockets: public beta, all plans, announced Node.js 2026-06-22 / Python 2026-07-23; the page lists Node, Bun, Nitro, Python, Next.js, and does NOT mention Go. Source: https://vercel.com/kb/guide/do-vercel-serverless-functions-support-websocket-connections (2026-10-24).
- Hobby max connection/function duration 300 s (Pro 800 s, 1800 s beta). Source: https://vercel.com/docs/functions/limitations (2026-10-24). Every connection ends; clients must reconnect.
- Connections pin to the accepting instance; a reconnect can land on another instance; in-memory state is not shared; needs an external store (Redis via Marketplace) with pub/sub. Memory is billed for instance lifetime while a socket is open.
- The Go engine would need to run as a Go function (not shown supported) or be re-hosted in Node/wasm. Doesn't fit "authoritative single-process referee". Static SPA stays on Vercel regardless.
- Verdict: no for matchmaking/game state. Pairing two players across stateless instances = Redis + polling, more parts than a $4 box.

### 6. Cloudflare Durable Objects (non-Go)
- Free plan: 100,000 req/day, 13,000 GB-s/day, SQLite-backed only. Paid: 1 M req/mo and 400,000 GB-s/mo included, then $0.15/M req and $12.50/M GB-s; WS Hibernation does not bill duration while idle; incoming WS messages billed 20:1. Paid plan $5/mo minimum noted in the examples (the Workers plan fee itself was not on this page: confirm). Source: https://developers.cloudflare.com/durable-objects/platform/pricing/ (2026-10-24).
- Technically elegant (one DO per match, hibernation, alarms for the 90 s timer), but the authoritative engine is Go. Would mean porting the rules engine (~1500 cards of effects) to JS/TS, or running wasm in a Worker: violates ADR-0001 "engine on server in Go" spirit and doubles maintenance. Rejected unless the owner drops Go-on-server.

## Not-sleeping cross-check
Any option needs the process to be a single always-on instance. Fly: set autostop off. Render: paid. Hetzner: inherent. This also means a deploy or crash drops in-memory games; ADR-0005 log persistence is the fix (append to disk/volume, re-fold on boot). Out of scope for the first slice.

## Recommendation
**Fly.io, one always-on 512 MB shared-cpu-1x Machine, stdlib Go server, SSE + POST, state in process memory.**
- ~$3.69/mo at 0 players; at ~50 concurrent players CPU/RAM unchanged (single-digit KB/s each), egress well under 10 GB => cents. Estimate, not measured.
- Zero new Go dependencies. Reuses existing `/api/move`; adds `/api/queue`, an SSE events stream, 90 s disconnect timer, 120 s bot fallback timer (all stdlib `time`).
- Fallback if Fly's SSE smoke test fails or the owner prefers a flat fee: Hetzner CX23 (EUR 5.99). If he wants a managed PaaS with documented WS: Render paid (get price first).
- Frontend stays on Vercel; browser calls the Fly origin cross-origin (CORS headers in the Go server) or via a Vercel rewrite (rewrites may buffer streams: UNVERIFIED, prefer direct origin).

## Owner must approve
1. Spend: ~$3.69/mo (Fly 512 MB) plus a card on file; ~$2/mo more only if a dedicated IPv4 is wanted (not needed).
2. Deploying anything (CLAUDE.md): first deploy of `server/` to Fly.
3. Dependency: none required. Only if he rejects SSE for WebSocket would `golang.org/x/net/websocket` (or similar) become a dependency decision.
4. Decision: accept that a redeploy/crash drops live games for now, or fund slice for event-log persistence (volume cost UNVERIFIED).
5. New files outside this research: Dockerfile, fly.toml, CORS handling.

## Open questions / unverified
- Fly SSE/long-connection behaviour and any idle timeout (docs fetched silent).
- Render and Railway sleep/price details noted above; Fly volume price; actual Go RSS.
