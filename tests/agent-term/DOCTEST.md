# agent-term CLI Doctests

End-to-end tests for the `agent-term` daemon CLI: `serve`, `list`, `run`,
`attach`, `rename`, and `web`.

# DSN (Domain Specific Notion)

**Participants**

- **agent-term subprocess** — built from `./cmd/agent-term` (or implementer path);
  subcommands `serve`, `list`, `run`, `attach`, `rename`, `web`.
- **ptywrap daemon** — TCP listener (default `127.0.0.1:7681`, overridable via
  `--listen`); in-memory session store.
- **Ephemeral web server** — `agent-term web` serves embedded xterm.js on
  `--port` (or OS-assigned ephemeral port).
- **Test harness** — starts/stops daemon, uses free TCP ports, isolated temp dirs.

**Behaviors**

- `serve` binds TCP and accepts HTTP connections; logs listen address and HTTP/WS
  requests to **stderr** (stdout stays clean for machine consumers).
- `list` without running daemon fails with message mentioning `agent-term serve`.
- `run <cmd>` creates session, **attaches** (TTY bridge via `ptyclient.Attach`), blocks;
  prints **only** session id on stdout after remote exit.
- `run` with non-TTY stdin/stdout errors like `attach` (interactive terminal required).
- Ctrl-C during interactive `run` detaches the client; remote session keeps running.
- `run sleep N` (and long-lived commands like `grok`) must not panic when WS read
  deadlines expire before the command exits.
- PTY `run sleep N` / `run grok` must not error with `read tcp ... i/o timeout`
  when the remote process is idle (websocket read deadline cleared after handshake).
- `serve` stderr logs one event per line; listen address must not mash into request lines.
- `attach <name>` resolves renamed session and connects (WS handshake succeeds).
- `web <id-or-name>` serves HTTP 200 HTML containing xterm.js markup.

## Version

0.0.2

## Decision Tree

```
[agent-term CLI]
 |
 +-- daemon/
 |    |
 |    +-- serve-accepts-tcp/          (LEAF)  TCP port accepts HTTP
 |    +-- serve-logs-startup/         (LEAF)  stderr logs listen address (RED)
 |    +-- serve-logs-on-create/       (LEAF)  stderr logs POST on session create (RED)
 |    +-- serve-logs-clean-lines/      (LEAF)  logs not mashed (no 7681POST) (RED)
 |
 +-- list/
 |    |
 |    +-- no-daemon-error/            (LEAF)  error mentions agent-term serve
 |
 +-- run/
 |    |
 |    +-- prints-id-on-exit/          (LEAF)  stdout is only session id after exit
 |    +-- prints-id-after-attached-exit/ (LEAF)  PTY run true prints only session id (RED)
 |    +-- attaches-pty-output/        (LEAF)  PTY run captures remote echo before id (RED)
 |    +-- interactive-session-survives-detach/ (LEAF)  SIGINT detach leaves session running (RED)
 |    +-- requires-tty/               (LEAF)  piped stdin errors like attach (RED)
 |    +-- waits-sleep-command/       (LEAF)  sleep exits cleanly, no WS read panic (RED)
 |    +-- pty-sleep-no-read-timeout/ (LEAF)  PTY run sleep 3 no i/o timeout (RED)
 |    +-- grok-no-read-timeout/       (LEAF)  PTY run grok stays attached 4s (RED)
 |
 +-- wait-session/
 |    |
 |    +-- sleep-exit/                 (LEAF)  WaitSession returns after sleep (RED)
 |
 +-- attach/
 |    |
 |    +-- by-name/                    (LEAF)  attach resolves renamed session
 |
 +-- web/
      |
      +-- serves-xterm-page/          (LEAF)  HTTP 200 with xterm HTML
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `daemon/serve-accepts-tcp` | `serve` listens on TCP; HTTP GET succeeds |
| 2 | `daemon/serve-logs-startup` | `serve` stderr logs listen address (RED) |
| 3 | `daemon/serve-logs-on-create` | `serve` stderr logs POST on session create (RED) |
| 4 | `daemon/serve-logs-clean-lines` | `serve` logs one line per event, no addr+POST mash (RED) |
| 5 | `list/no-daemon-error` | `list` without daemon mentions `agent-term serve` |
| 6 | `run/prints-id-on-exit` | `run true` prints only session id after exit |
| 7 | `run/prints-id-after-attached-exit` | PTY `run true` prints only session id (RED) |
| 8 | `run/attaches-pty-output` | PTY `run` captures `RUN_OK` before session id (RED) |
| 9 | `run/interactive-session-survives-detach` | SIGINT during `run sleep` leaves session running (RED) |
| 10 | `run/requires-tty` | `run bash` with pipes errors on interactive terminal (RED) |
| 11 | `run/waits-sleep-command` | `run sleep 2` exits without WS read panic (RED) |
| 12 | `run/pty-sleep-no-read-timeout` | PTY `run sleep 3` completes without `i/o timeout` (RED) |
| 13 | `run/grok-no-read-timeout` | PTY `run grok` no `i/o timeout` after 4s idle (RED) |
| 14 | `wait-session/sleep-exit` | `WaitSession` returns after `sleep 2` (RED) |
| 15 | `attach/by-name` | `attach` connects to session resolved by name |
| 16 | `web/serves-xterm-page` | `web` returns xterm.js HTML page |

## How to Run

```sh
doctest vet ./tests/agent-term
doctest test ./tests/agent-term/...
```

```go
import (

	"testing"

	"github.com/xhd2015/agent-pro/tests/agent-term/termtest"
	"github.com/xhd2015/doctest/session"
)

type Request = termtest.Request
type Response = termtest.Response

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	_ = d
	return termtest.Run(t, req)
}

func buildAgentTerm(t *testing.T, d *session.Doctest) string {
	return termtest.BuildAgentTerm(t, d.DOCTEST_ROOT)
}

func pickFreePort(base int) (int, error) {
	return termtest.PickFreePort(base)
}
```