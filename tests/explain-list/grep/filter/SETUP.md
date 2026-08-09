# Scenario

**Feature**: --grep filters which sessions are listed (combine modes + limit)

```
# multi-session store; explain list --grep ...
-> only matching sessions; title of N = match count; or dedicated empty msgs
```

## Preconditions

- Leaves seed two or more sessions with distinctive body markers when testing
  keep/drop; empty-store leaf seeds none.
- No `--color` in filter leaves (plain output; highlight covered under
  `grep/highlight/`).

## Steps

1. Seed fixtures with unique Q/A substrings.
2. Run `list` with one or more `--grep` and optional `--or`/`--and`/`--limit`.
3. Assert keep-set, order (newest first among matches), title counts, or
   empty/no-match messages.

## Context

- Non-matching sessions must not appear in stdout.
- Empty store message is distinct from zero-match-with-store message.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.Bin == "" {
		t.Fatalf("grep/filter setup: explain binary not built")
	}
	return nil
}

// grepTripleSessions: three one-turn sessions newest→oldest markers.
// s2 (newest): kubernetes/pods
// s1: docker/containers
// s0 (oldest): redis/cache
func grepTripleSessions() []SessionSeed {
	return []SessionSeed{
		simpleSession(
			"2026-07-10-10-00-00-redis-aaaaaaaa",
			"opencode", "deepseek-chat",
			"redis marker-redis", "cache marker-redis-a",
		),
		simpleSession(
			"2026-07-11-10-00-00-docker-bbbbbbbb",
			"opencode", "deepseek-chat",
			"docker marker-docker", "containers marker-docker-a",
		),
		simpleSession(
			"2026-07-12-10-00-00-k8s-cccccccc",
			"opencode", "deepseek-chat",
			"kubernetes marker-k8s", "pods marker-k8s-a",
		),
	}
}
```
