# Scenario

**Feature**: one `usage.Collect` cycle over all three providers

```
Collect(opts) -> [grok, codex, commandcode]   # each: fixtures + home session tree
doctest <- UsageBlock (values, display, endpoint, ok) + SessionsBlock (total, oldest, newest)
```

## Preconditions

- Inherits the root fixed clock, homes, credential files and fixtures.
- Every provider starts with one session on disk, so a leaf only adds the shape
  it is about: grok `2026-09-10T09-00-00-main`, codex
  `rollout-2026-09-10T09-00-00-...`, one Command Code transcript.

## Steps

1. Seed one session per provider home.
2. Leaf changes one variable: a missing fixture, or credentials.
3. Run collects; Assert checks the provider blocks the leaf is about.

## Context

- Session timestamps are absolute (`oldest` from the tree, `newest` from the file
  modification time), so leaves assert them without depending on the host timezone.

```go
import (
"testing"
"time"

"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
WriteGrokSession(t, req.GrokHome, "2026-09-10T09-00-00-main",
"2026-09-10T09:00:00Z", "2026-09-10T10:00:00Z")
WriteCodexRollout(t, req.CodexHome, "2026/09/10", "2026-09-10T09-00-00",
"11111111-1111-4111-8111-111111111111", req.Now.Add(-24*time.Hour))
WriteCommandCodeTranscript(t, req.CommandCodeHome, "-Users-tester-project",
"22222222-2222-4222-8222-222222222222.jsonl",
`{"type":"session","timestamp":"2026-09-10T09:00:00Z"}`,
req.Now.Add(-24*time.Hour))
return nil
}
```
