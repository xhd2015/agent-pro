# Scenario

**Feature**: list should return quickly when only a small limit is requested

## Preconditions

- Many rollout JSONL files exist under the synthetic Codex home (more than the
  default list limit).
- Each rollout file contains hundreds of displayable events so `List` cannot
  cheaply ignore file bodies.

## Steps

1. Write 200 rollout sessions, each with 400 `agent_message` lines (~80k parsed
   events total across the tree).
2. Call `sessions.List(codexHome, 20)` via the harness `Run`.

## Context

Listing should read only enough metadata to pick the 20 newest sessions.
Current behavior fully reads and parses every rollout file to compute TITLE and
MSGS before sorting and truncating — this matches the user's ~10s
`agent-pro codex sessions` delay with hundreds of on-disk rollouts.

```go
import (
	"fmt"
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Limit = 20
	base, err := time.Parse(time.RFC3339, "2026-07-03T00:00:00.000Z")
	if err != nil {
		t.Fatalf("parse base time: %v", err)
	}
	for i := 0; i < 200; i++ {
		id := fmt.Sprintf("01900001-0000-7000-8000-%012d", i)
		ts := base.Add(time.Duration(i) * time.Minute).UTC().Format("2006-01-02T15:04:05.000Z")
		var extra []string
		for j := 0; j < 400; j++ {
			extra = append(extra, agentMessageLine(fmt.Sprintf("bulk message %d-%d", i, j)))
		}
		if i == 199 {
			extra = append(extra, userMessageLine("newest session title"))
		}
		writeRolloutSession(t, req.CodexHome, id, ts, "/tmp/bulk-project", extra...)
	}
	return nil
}
```