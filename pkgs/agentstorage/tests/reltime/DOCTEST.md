# agentstorage FormatRelativeAge

Doc-style tests for pure relative-age formatting used by human `agent-run sessions`
UPDATED column. Fixed `now` — no wall-clock flakiness.

# DSN (Domain Specific Notion)

**FormatRelativeAge** turns a pair of instants into a short human age string for
list UIs. Callers pass an explicit clock (`now`) and a target time `t` (session
`updated_at`, else `created_at`). Missing/zero `t` is a sentinel dash; ages under
one second read as “just now”; longer ages use at most two short units and a
trailing ` ago`.

```
now, t
  -> FormatRelativeAge(now, t) -> display string

participants:
  clock (now)     fixed in tests; production uses wall clock
  target (t)      session updated_at / created_at parsed as time.Time
  formatter       pure function in pkgs/agentstorage

behaviors:
  zero/missing t  -> "-"
  d = now.Sub(t); if d < 0 treat as 0
  d < 1s          -> "just now" (no " ago")
  else            -> up to 2 units from d/h/m/s, short labels, " ago"
  zero unit       stops the chain (omit that unit and everything smaller)
```

**Expected symbol (implementer):**

```text
package agentstorage

// FormatRelativeAge formats the age of t relative to now for human session lists.
// Zero t → "-"; age < 1s (after clamping future to 0) → "just now";
// otherwise short units (s/m/h/d), max 2 non-zero units, zero stops chain, " ago".
func FormatRelativeAge(now, t time.Time) string
```

CLI human list must call this (or an equivalent wrapper) when rendering UPDATED.
JSON list keeps absolute RFC3339 timestamps (covered under `cmd/agent-run/tests/sessions/list`).

## Version

0.0.2

## Decision Tree

```
pkgs/agentstorage/tests/reltime/
├── DOCTEST.md
├── SETUP.md
├── missing-or-zero/     zero time.Time → "-"
├── just-now/            age < 1s (incl. exact 0 and future clamped) → "just now"
├── single-unit/         1s, 2s, 1h, 90d → one unit + " ago"
├── two-units/           65s→1m5s, 1h2m → two non-zero units
├── zero-stops-chain/    1h0m5s→1h, 4d0h2m→4d (zero unit ends chain)
└── max-two-units/       4d5h12m→4d5h (third unit dropped)
```

Parameter ranking (most → least significant):

1. **Target presence** — zero `t` vs real instant
2. **Sub-second vs aged** — `just now` branch vs unit formatting
3. **Unit-selection rules** — single unit, two units, zero-stop, max-two

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `missing-or-zero` | Zero `time.Time` → `-` |
| 2 | `just-now` | 0.5s age, exact equal, future target → `just now` |
| 3 | `single-unit` | `1s ago`, `2s ago`, `1h ago`, `90d ago` |
| 4 | `two-units` | `1m5s ago` (65s), `1h2m ago` |
| 5 | `zero-stops-chain` | `1h0m5s` → `1h ago`; `4d0h2m` → `4d ago` |
| 6 | `max-two-units` | `4d5h12m` → `4d5h ago` (minutes omitted) |

## How to Run

```sh
doctest vet ./pkgs/agentstorage/tests/reltime
doctest test -v ./pkgs/agentstorage/tests/reltime
doctest test -v ./pkgs/agentstorage/tests/reltime/zero-stops-chain
```

```go
import (

	"fmt"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/doctest/session"
)

// FormatCase is one (target, want) pair under a shared fixed now.
type FormatCase struct {
	Name   string
	Target time.Time
	Want   string
}

type Request struct {
	Now   time.Time
	Cases []FormatCase
}

type Response struct {
	Got []string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	if len(req.Cases) == 0 {
		return nil, fmt.Errorf("req.Cases must be set by leaf Setup")
	}
	got := make([]string, 0, len(req.Cases))
	for _, c := range req.Cases {
		got = append(got, agentstorage.FormatRelativeAge(req.Now, c.Target))
	}
	return &Response{Got: got}, nil
}
```
