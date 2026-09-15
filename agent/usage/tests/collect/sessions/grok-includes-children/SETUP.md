# Scenario

**Feature**: every grok session on disk counts, including subagent and fork children

```
sessions/<main>/summary.json                        created 09-01, updated 09-01
sessions/<main>/subagents/<child>/summary.json      created 09-02, updated 09-02
sessions/<fork>/summary.json                        created 09-03, updated 09-03
sessions/<no-timestamps>/summary.json               counts, moves no bound
sessions/<malformed>/summary.json                   ignored
doctest <- total 4, oldest 09-01, newest 09-03
```

## Preconditions

- Inherits the root fixtures; nothing seeds sessions except this leaf.
- One summary carries no timestamps and one is not JSON at all.

## Steps

1. Write three normal summaries: a main session, a subagent child under the main
   session's `subagents/` directory, and a fork.
2. Write a summary with no `created_at`/`updated_at`, and one that is not JSON.

```go
import (
"os"
"path/filepath"
"testing"

"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
WriteGrokSession(t, req.GrokHome, "2026-09-01T08-00-00-main",
"2026-09-01T08:00:00Z", "2026-09-01T09:00:00Z")
WriteGrokSession(t, req.GrokHome, "2026-09-01T08-00-00-main/subagents/child",
"2026-09-02T08:00:00Z", "2026-09-02T09:00:00Z")
WriteGrokSession(t, req.GrokHome, "2026-09-03T08-00-00-fork",
"2026-09-03T08:00:00Z", "2026-09-03T09:00:00Z")
WriteGrokSession(t, req.GrokHome, "2026-09-04T08-00-00-no-timestamps", "", "")

broken := filepath.Join(req.GrokHome, "sessions", "2026-09-05T08-00-00-broken")
if err := os.MkdirAll(broken, 0o755); err != nil {
return err
}
return os.WriteFile(filepath.Join(broken, "summary.json"), []byte("not json"), 0o644)
}
```
