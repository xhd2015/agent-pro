# tty/detection — resting change + occupy space compare

L2 coverage for runner-agnostic exit-on-idle primitives under
`pkgs/tty/detection/{occupied,changed}`. Idle scheduling lives in
`tests/agentruncli/idle-watchdog`.

**Out of scope:** live TTY inject e2e (`run-exit-on-idle-*-tty`), serve wiring,
policy file I/O.

# DSN (Domain Specific Notion)

Both packages compare before/after snapshots.

- **changed** — this resting snap vs last; **raw** byte inequality ⇒ activity
  (newlines count; no strip)
- **occupied** — after space inject vs before; newline-stripped then exactly
  one more ASCII draft space ⇒ occupied

## Version

0.1.0

## Decision Tree

```
tests/tty/detection/
├── DOCTEST.md
├── SETUP.md
├── occupied/
│   ├── exactly-one-space/
│   ├── not-exactly/
│   └── empty-before/
└── changed/
    ├── differs/
    └── newlines-count/
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `occupied/exactly-one-space` | draft + trailing space ⇒ true |
| 2 | `occupied/not-exactly` | placeholder collapse ⇒ false |
| 3 | `occupied/empty-before` | `""` → `" "` ⇒ false |
| 4 | `changed/differs` | content change ⇒ true |
| 5 | `changed/newlines-count` | newline-only ⇒ true (raw equality) |

## How to Run

```sh
doctest vet ./tests/tty/detection
doctest test ./tests/tty/detection
```

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/tty/detection/changed"
	"github.com/xhd2015/agent-pro/pkgs/tty/detection/occupied"
	"github.com/xhd2015/doctest/session"
)

const (
	opOccupied = "occupied"
	opChanged  = "changed"
)

type Request struct {
	Op            string
	Before, After string
}

type Response struct {
	ExactlyOneMoreSpace bool
	Changed             bool
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d
	switch req.Op {
	case opOccupied:
		return &Response{
			ExactlyOneMoreSpace: occupied.ExactlyOneMoreSpace([]byte(req.Before), []byte(req.After)),
		}, nil
	case opChanged:
		return &Response{
			Changed: changed.Changed([]byte(req.Before), []byte(req.After)),
		}, nil
	default:
		t.Fatalf("unknown Op %q", req.Op)
		return nil, nil
	}
}
```
