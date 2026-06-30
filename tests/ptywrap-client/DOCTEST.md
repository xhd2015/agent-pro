# ptywrap/client Library Doctests

Tests for `github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap/client` — HTTP
session client, `ResolveTarget`, and local TTY↔WS `Attach` bridge.

# DSN (Domain Specific Notion)

**Participants**

- **HTTP client** — `List`, `Create`, `Delete`, `Rename` against daemon REST API.
- **ResolveTarget** — maps id-or-name string to exactly one `SessionInfo`.
- **Attach bridge** — dials WS, raw TTY mode, forwards stdin/stdout; requires
  interactive terminal.
- **Mock daemon** — test HTTP+WS server backed by ptywrap handlers or lightweight
  stub returning canned session lists and `session_id` messages.
- **PTY simulator** — for non-TTY tests, uses pipe fds instead of `/dev/tty`.

**Behaviors**

- `Attach` returns error when stdin or stdout is not a terminal.
- `ResolveTarget` exact id match returns that session.
- `ResolveTarget` unique name match returns the session.
- `ResolveTarget` with multiple same-name sessions returns ambiguity error.
- `Attach` parses initial `session_id` JSON from server and returns it in
  `AttachResult`.

## Version

0.0.2

## Decision Tree

```
[ptywrap/client]
 |
 +-- attach/
 |    |
 |    +-- requires-tty/              (LEAF)  error when stdin not a terminal
 |    +-- captures-session-id/      (LEAF)  returns id from server message
 |
 +-- resolve-target/
      |
      +-- exact-id/                  (LEAF)  id match
      +-- single-name-match/         (LEAF)  unique name match
      +-- ambiguous-name/            (LEAF)  duplicate names → error
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `attach/requires-tty` | Attach errors when stdin is not a TTY |
| 2 | `attach/captures-session-id` | Attach returns session id from WS message |
| 3 | `resolve-target/exact-id` | ResolveTarget finds session by id |
| 4 | `resolve-target/single-name-match` | ResolveTarget finds session by unique name |
| 5 | `resolve-target/ambiguous-name` | ResolveTarget errors on ambiguous name |

## How to Run

```sh
doctest vet ./tests/ptywrap-client
doctest test ./tests/ptywrap-client/...
```

```go
import (
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap/client/clienttest"
)

type Request = clienttest.Request
type Response = clienttest.Response

func Run(t *testing.T, req *Request) (*Response, error) {
	return clienttest.Run(t, req)
}
```