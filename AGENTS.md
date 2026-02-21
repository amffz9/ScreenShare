# AGENTS.md — Agent Guidelines for WebRTC Screen Sharing App

## Role Definitions

### Code Review Agent
- Focus on security (XSS, injection), concurrency (mutex ordering, goroutine leaks), and WebRTC correctness.
- Check that innerHTML usage escapes user-controlled data via `escapeHtml()` from `common.js`.
- Verify WebSocket message handling covers all `type` cases and validates `msg.To` targets.

### Build Agent
- Use `make build-all` for release builds.
- Use `make run` for local development/testing.
- Ensure `go-winres` is installed before running `make icon`.
- The `.syso` file in the project root is a Windows resource — do not delete it.

### Frontend Agent
- All pages share `public/common.js` — add shared utilities there.
- Use `escapeHtml()` for any user-controlled data rendered via innerHTML.
- RTC configuration is fetched from `/config` at page load. Do not hardcode STUN/TURN servers.
- Use CSS variables from `style.css` `:root` for colors — do not hardcode color values.

### Backend Agent
- Single-file Go server (`main.go`). Keep it that way unless complexity demands splitting.
- Global `mu` (RWMutex) protects `clients` and `broadcasters` maps. Per-client `client.mu` protects WebSocket writes.
- Lock ordering: always acquire global `mu` before `client.mu`. Never hold `client.mu` while acquiring `mu`.
- Use `sendJSON()` for all WebSocket writes — it handles the per-client mutex.
- Config priority: hardcoded defaults -> `config.json` -> CLI flags.

## Task Coordination

- Frontend and backend are tightly coupled through the signaling protocol. Changes to message types must be updated in both `main.go` (switch cases) and the relevant HTML files.
- The signaling message types are: `register-broadcaster`, `register-viewer`, `get-broadcasters`, `offer`, `answer`, `ice-candidate`, `registered`, `broadcaster-list`, `broadcaster-joined`, `broadcaster-left`.
- Adding a new message type requires: (1) add to `IncomingMessage` struct if needed, (2) add switch case in `handleWebSocket`, (3) handle in the relevant JS `handleSignalingMessage`.

## Testing

- No automated tests currently exist. Test manually:
  1. `make run` to start the server.
  2. Open `/broadcaster` in one tab, `/viewer` in another.
  3. Start broadcasting, then connect from the viewer.
  4. Verify video stream appears in viewer.
  5. Test disconnect/reconnect behavior.
- For cross-machine testing, use the local network IP shown in the server logs.
