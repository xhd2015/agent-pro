# Scenario

**Feature**: list sorts by created-desc and applies limit rules

```
# seed N sessions with distinct dirname timestamps
explain list [--limit K] -> newest first, at most effective_limit cards
# effective_limit = 10 if omitted or N<=0; min(N, 100) if N>0
```

## Preconditions

- Session dirname prefixes encode creation time `YYYY-MM-DD-HH-mm-ss`.
- No reliance on filesystem mtime.

## Steps

1. Leaves seed multiple `SessionSeed` entries with ordered timestamps.
2. Leaves set `--limit` when testing non-default limits.
3. Assert order (newest first), shown count, and title limit value.

## Context

- Title must include both shown count and limit (exact title wording may vary
  slightly but fixtures lock the requirement example shape).

```go
import (
	"fmt"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.Bin == "" {
		t.Fatalf("order-and-limit setup: explain binary not built")
	}
	return nil
}

// seedNSessions creates n simple one-turn sessions with increasing timestamps
// so higher i is newer. Dirnames: 2026-07-01-HH-mm-ss-s{i}-aaaaaaaa
func seedNSessions(n int, hourBase int) []SessionSeed {
	out := make([]SessionSeed, 0, n)
	for i := 0; i < n; i++ {
		// Spread across hours/minutes so lexicographic and parse order match.
		h := hourBase + i/60
		m := i % 60
		dir := fmt.Sprintf("2026-07-01-%02d-%02d-00-s%02d-aaaaaaaa", h, m, i)
		q := fmt.Sprintf("question-%02d", i)
		a := fmt.Sprintf("answer-%02d", i)
		out = append(out, simpleSession(dir, "opencode", "deepseek-chat", q, a))
	}
	return out
}
```
