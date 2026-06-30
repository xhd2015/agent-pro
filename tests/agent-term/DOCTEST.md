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

- `serve` binds TCP and accepts HTTP connections.
- `list` without running daemon fails with message mentioning `agent-term serve`.
- `run <cmd>` creates session, attaches, blocks; prints **only** session id on exit.
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
 |
 +-- list/
 |    |
 |    +-- no-daemon-error/            (LEAF)  error mentions agent-term serve
 |
 +-- run/
 |    |
 |    +-- prints-id-on-exit/          (LEAF)  stdout is only session id after exit
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
| 2 | `list/no-daemon-error` | `list` without daemon mentions `agent-term serve` |
| 3 | `run/prints-id-on-exit` | `run true` prints only session id after exit |
| 4 | `attach/by-name` | `attach` connects to session resolved by name |
| 5 | `web/serves-xterm-page` | `web` returns xterm.js HTML page |

## How to Run

```sh
doctest vet ./tests/agent-term
doctest test ./tests/agent-term/...
```

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/cmd/agent-term/termtest"
)

type Request = termtest.Request
type Response = termtest.Response

func Run(t *testing.T, req *Request) (*Response, error) {
	return termtest.Run(t, req)
}

func buildAgentTerm(t *testing.T) string {
	return termtest.BuildAgentTerm(t)
}

func pickFreePort(base int) (int, error) {
	return termtest.PickFreePort(base)
}
```