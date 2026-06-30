# ptywrap Server Library Doctests

Tests for `github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap` — PTY session
manager, WebSocket attach protocol, scrollback, resize, REST session API, and
exited-session retention.

# DSN (Domain Specific Notion)

**Participants**

- **Session manager** — in-memory registry of PTY sessions; assigns `session-N`
  ids; tracks `running` / `exited` status and `connected` WS state.
- **PTY spawner** — starts shell or arbitrary command in a pseudo-terminal with
  optional cwd, name, and RC-patch env.
- **Scrollback buffer** — ring buffer of recent PTY output; replayed on attach;
  query sequences stripped for safe reconnect.
- **HTTP handler** — REST endpoints under `/api/terminal/sessions` and WS
  upgrade at `/api/terminal`.
- **WebSocket bridge** — binary client→PTY input; text JSON resize + session_id
  messages; binary PTY→client output.
- **Test harness** — ephemeral `httptest` or TCP listener wrapping ptywrap
  handlers; WS/HTTP clients without a real browser.

**Behaviors**

- Default shell spawn: child stdout is a TTY (`isatty`).
- Arbitrary command spawn: PTY output (e.g. `echo hello`) reaches WS client.
- WS keystrokes written to PTY stdin appear in output stream.
- `{"type":"resize","cols":N,"rows":N}` updates PTY window size.
- WS disconnect keeps session alive; re-attach replays scrollback.
- `POST /api/terminal/sessions` returns new session metadata with id.
- `PATCH /api/terminal/sessions/{id}` renames; list reflects new name.
- Exited sessions remain listed with `status: "exited"` until deleted.
- WS connect with `name` + `cwd` query (no `session_id`) creates shell session
  and sends `session_id` JSON (ai-critic legacy compat).

## Version

0.0.2

## Decision Tree

```
[ptywrap server]
 |
 +-- spawn/
 |    |
 |    +-- shell-is-tty/              (LEAF)  default shell child has TTY stdout
 |    +-- arbitrary-command/          (LEAF)  echo hello captured via WS
 |
 +-- websocket/
 |    |
 |    +-- input-round-trip/         (LEAF)  keystrokes reach PTY output
 |    +-- resize-updates-pty/        (LEAF)  resize JSON changes PTY dimensions
 |    +-- reconnect-scrollback/     (LEAF)  detach then re-attach replays output
 |    +-- create-on-connect/        (LEAF)  name+cwd query creates shell session
 |
 +-- rest-api/
 |    |
 |    +-- create-session/           (LEAF)  POST returns session id
 |    +-- rename-session/            (LEAF)  PATCH updates name in list
 |
 +-- session-lifecycle/
      |
      +-- exited-stays-listed/      (LEAF)  exited session remains in GET list
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `spawn/shell-is-tty` | Shell child reports `isatty(stdout)==true` via PTY |
| 2 | `spawn/arbitrary-command` | `echo hello` output captured through WS |
| 3 | `websocket/input-round-trip` | Binary WS input appears in PTY output |
| 4 | `websocket/resize-updates-pty` | Resize JSON updates PTY cols/rows |
| 5 | `websocket/reconnect-scrollback` | Scrollback replayed after WS reconnect |
| 6 | `websocket/create-on-connect` | Legacy `name`+`cwd` WS create sends `session_id` |
| 7 | `rest-api/create-session` | REST POST creates session and returns id |
| 8 | `rest-api/rename-session` | REST PATCH renames; GET list shows new name |
| 9 | `session-lifecycle/exited-stays-listed` | Exited session listed with `status: exited` |

## How to Run

```sh
doctest vet ./tests/ptywrap
doctest test ./tests/ptywrap/...
```

```go
import (
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap/ptytest"
)

type Request = ptytest.Request
type Response = ptytest.Response

func Run(t *testing.T, req *Request) (*Response, error) {
	return ptytest.Run(t, req)
}

func startTestServer(t *testing.T) (base string, cleanup func()) {
	return ptytest.StartTestServer(t)
}

func absTempDir(t *testing.T) string {
	return ptytest.AbsTempDir(t)
}
```