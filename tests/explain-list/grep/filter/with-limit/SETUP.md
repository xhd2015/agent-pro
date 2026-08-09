# Scenario

**Feature**: --limit applies after --grep filter; title of N is match count

```
# 6 sessions; 4 match "hit"; --grep hit --limit 2
-> 2 newest matches shown; title "2 shown of 4, limit 2"
```

## Preconditions

- Six sessions: four with `hit-N` markers (N=0..3 increasing time), two miss.

## Steps

1. Seed six sessions.
2. Args: `list --grep hit --limit 2`.
3. Assert only two newest hits; title match totals; misses absent.

## Context

- Pipeline: sort newest-first → filter → total=match count → apply limit.

```go
import (
	"fmt"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"list", "--grep", "hit", "--limit", "2"}
	// i=0..3 hit (increasing time); i=4,5 miss (newest miss sessions still filtered out).
	var seeds []SessionSeed
	for i := 0; i < 4; i++ {
		dir := fmt.Sprintf("2026-07-20-%02d-00-00-hit%02d-hhhhhhhh", 10+i, i)
		q := fmt.Sprintf("hit-%02d marker-hit-%02d", i, i)
		seeds = append(seeds, simpleSession(dir, "opencode", "deepseek-chat", q, "a"))
	}
	seeds = append(seeds,
		simpleSession(
			"2026-07-20-20-00-00-miss0-mmmmmmmm",
			"opencode", "deepseek-chat",
			"miss-00 marker-miss-00", "a",
		),
		simpleSession(
			"2026-07-20-21-00-00-miss1-mmmmmmmm",
			"opencode", "deepseek-chat",
			"miss-01 marker-miss-01", "a",
		),
	)
	req.Sessions = seeds
	return nil
}
```
