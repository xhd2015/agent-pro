# Scenario

**Feature**: a codex session's start comes from its rollout file name, its activity from the file

```
sessions/2026/09/10/rollout-2026-09-10T09-00-00-<uuid>.jsonl   mtime now-48h
sessions/2026/09/12/rollout-2026-09-12T15-30-00-<uuid>.jsonl   mtime now-1h
sessions/2026/09/12/notes.txt                                  not a session
doctest <- total 2, oldest from the oldest name, newest now-1h
```

## Preconditions

- Inherits the root fixtures; nothing else seeds the codex home.
- A codex rollout file name carries the session start in local time, and the file
  modification time stands in for the last activity.

## Steps

1. Write two rollouts, the earlier one modified first.
2. Write a non-rollout file next to them.

```go
import (
"os"
"path/filepath"
"testing"
"time"

"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
WriteCodexRollout(t, req.CodexHome, "2026/09/10", "2026-09-10T09-00-00",
"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", req.Now.Add(-48*time.Hour))
WriteCodexRollout(t, req.CodexHome, "2026/09/12", "2026-09-12T15-30-00",
"bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb", req.Now.Add(-1*time.Hour))

dir := filepath.Join(req.CodexHome, "sessions", "2026", "09", "12")
return os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("not a rollout\n"), 0o644)
}
```
