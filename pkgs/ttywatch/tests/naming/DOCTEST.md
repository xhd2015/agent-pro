# ttywatch naming tests

Doc-style tests for `github.com/xhd2015/agent-pro/pkgs/ttywatch` command-line
slug helpers used when re-executing the tty-watch binary as an embedded PTY
server.

# DSN (Domain Specific Notion)

**SlugifyCommandLine** takes a session command argv (the user command passed to
`tty-watch run`, not the tty-watch CLI argv) and produces a stable slug token.
**ServeSubcommand** wraps that slug as the internal reexec argv0 token
`__serve_{slug}__`.

```
run argv (codex, flags, paths, metachars)
  -> SlugifyCommandLine(argv) -> slug (sanitized, collapsed, maybe truncated+hash)
  -> ServeSubcommand(argv) -> __serve_{slug}__
reexec dispatch: argv[0] hasPrefix __serve_ && hasSuffix __
```

Slug rules (from refactor spec):

- Join argv with `_`
- `/` and `\` become `_`; shell metacharacters become `_`
- Strip leading `--` on flag tokens before joining
- Allow alnum, `-`, `.` in output (everything else sanitized to `_`)
- Collapse repeated `_`, trim leading/trailing `_`
- When slug exceeds ~120 characters, truncate and append a hash suffix
- Final reexec token: `__serve_{slug}__`

Tests call `SlugifyCommandLine` and `ServeSubcommand` directly (no subprocess).

## Version

0.0.2

## Decision Tree

```
pkgs/ttywatch/tests/naming/
├── DOCTEST.md
├── SETUP.md
├── full-argv-codex/          codex argv with flags: join + -- stripping
├── path-in-argv/             absolute / and \ paths sanitized to underscores
├── shell-metachars/          ; | & $ and similar become underscores
└── truncate-long/            very long argv: slug capped ~120 + hash suffix
```

Parameter ranking (most → least significant):

1. **Sanitization trigger** — path separators vs shell metachars vs flag `--` stripping
2. **Argv shape** — short typical argv vs intentionally oversized argv
3. **Output surface** — bare slug vs wrapped `ServeSubcommand` token

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `full-argv-codex` | `codex run --model gpt-5.5 medium` → `codex_run_model_gpt-5.5_medium` and matching `__serve_*__` (RED) |
| 2 | `path-in-argv` | Unix + Windows path argv: no raw `/` or `\` in slug; collapsed underscores (RED) |
| 3 | `shell-metachars` | `sh -c` script fragment with `;|&$`: metachars replaced, slug stable (RED) |
| 4 | `truncate-long` | 40+ flag pairs: slug length bounded; hash suffix; `ServeSubcommand` still dispatches prefix/suffix (RED) |

## How to Run

```sh
doctest vet ./pkgs/ttywatch/tests/naming
doctest test ./pkgs/ttywatch/tests/naming/...
doctest test -v ./pkgs/ttywatch/tests/naming/full-argv-codex
```

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/ttywatch"
)

type Request struct {
	Argv []string
}

type Response struct {
	Slug            string
	ServeSubcommand string
}

func Run(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if len(req.Argv) == 0 {
		t.Fatal("req.Argv must be set by leaf Setup")
	}
	slug := ttywatch.SlugifyCommandLine(req.Argv)
	serve := ttywatch.ServeSubcommand(req.Argv)
	return &Response{
		Slug:            slug,
		ServeSubcommand: serve,
	}, nil
}
```