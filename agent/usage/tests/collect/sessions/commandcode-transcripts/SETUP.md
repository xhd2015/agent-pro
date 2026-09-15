# Scenario

**Feature**: Command Code transcripts count and their checkpoint sidecars do not

```
projects/<project>/<uuid>.jsonl              {"type":"session","timestamp":...}
projects/<project>/<uuid>.checkpoints.jsonl  sidecar, not a session
projects/<project>/<uuid>.jsonl              {"id":...,"createdAt":...} fallback
projects/<project>/notes.md                  not a session
doctest <- total 2, oldest from createdAt, newest from the newest mtime
```

## Preconditions

- Inherits the root fixtures; nothing else seeds the Command Code home.
- A transcript's first line is the session record; its start time is either
  `timestamp` (the session record shape) or `createdAt` (the id-bearing shape).

## Steps

1. Write one transcript with a `type: session` first record and one whose first
   record only carries `createdAt`.
2. Write a `.checkpoints.jsonl` sidecar and a non-transcript file.

```go
import (
"os"
"path/filepath"
"testing"
"time"

"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
WriteCommandCodeTranscript(t, req.CommandCodeHome, "-Users-tester-alpha",
"44444444-4444-4444-8444-444444444444.jsonl",
`{"type":"session","timestamp":"2026-09-12T09:00:00Z"}`,
req.Now.Add(-2*time.Hour))
WriteCommandCodeTranscript(t, req.CommandCodeHome, "-Users-tester-beta",
"55555555-5555-4555-8555-555555555555.jsonl",
`{"id":"55555555-5555-4555-8555-555555555555","createdAt":"2026-09-11T09:00:00Z"}`,
req.Now.Add(-3*time.Hour))
WriteCommandCodeTranscript(t, req.CommandCodeHome, "-Users-tester-alpha",
"44444444-4444-4444-8444-444444444444.checkpoints.jsonl",
`{"id":"checkpoint"}`, req.Now)

dir := filepath.Join(req.CommandCodeHome, "projects", "-Users-tester-alpha")
return os.WriteFile(filepath.Join(dir, "notes.md"), []byte("not a transcript\n"), 0o644)
}
```
