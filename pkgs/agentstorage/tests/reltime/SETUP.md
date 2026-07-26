# Scenario

**Feature**: pure relative age formatting with fixed clock for session UPDATED cells

```
fixed now + target t
  -> agentstorage.FormatRelativeAge(now, t)
  -> "-", "just now", or "<units> ago"
```

## Preconditions

- Implementer exports `FormatRelativeAge(now, t time.Time) string` from
  `github.com/xhd2015/agent-pro/pkgs/agentstorage` (e.g. `reltime.go`).
- Leaves use a shared fixed `now` (UTC) so assertions are exact strings.
- No filesystem, no CLI, no store.

## Steps

1. Root `Setup` sets default fixed `req.Now` when unset.
2. Leaf `Setup` appends `req.Cases` with target times and expected strings.
3. `Run` calls `FormatRelativeAge` for each case.
4. `Assert` compares `resp.Got[i]` to `req.Cases[i].Want`.

## Context

- Default `now`: `2026-07-12T12:00:00Z`.
- Duration helpers build targets as `now.Add(-d)` so examples match the design table.
- Future targets are still expressed as `now.Add(+d)`; formatter clamps negative age to 0.

```go
import (
	"testing"
	"time"
)

// fixedNow is the shared clock for all reltime leaves (overridable per leaf).
var fixedNow = time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.Now.IsZero() {
		req.Now = fixedNow
	}
	return nil
}

func ageTarget(now time.Time, d time.Duration) time.Time {
	return now.Add(-d)
}

func assertCases(t *testing.T, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
	if len(resp.Got) != len(req.Cases) {
		t.Fatalf("got %d results, want %d cases", len(resp.Got), len(req.Cases))
	}
	for i, c := range req.Cases {
		if resp.Got[i] != c.Want {
			t.Errorf("case %q: FormatRelativeAge = %q, want %q", c.Name, resp.Got[i], c.Want)
		}
	}
}
```
