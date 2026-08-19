# Scenario

**Feature**: active transcript discovery ignores stale same-cwd transcripts

```
old rollout JSONL has session_meta.cwd = workspace and stale text
  -> current run creates a newer same-cwd rollout JSONL
  -> newest current transcript wins
```

## Preconditions

- The stale and current transcripts both have matching `session_meta.cwd`.
- The stale transcript exists before `agent-run` starts and has an older mtime.

## Steps

1. Write an older matching-cwd transcript with stale assistant text.
2. Schedule the current run transcript creation and assistant message.
3. Assert stdout includes current text and excludes stale text.

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
	"time"
	"github.com/xhd2015/doctest/session"
)

const codexStaleCWDText = "JSONL_STALE_CWD_SHOULD_NOT_STREAM"
const codexCurrentCWDText = "JSONL_CURRENT_CWD_SHOULD_STREAM"

func writeStaleCodexTranscript(t *testing.T, req *Request) {
	t.Helper()
	staleSessionID := "019f20fd-8569-7910-ab0b-000000000001"
	stalePath := filepath.Join(req.CodexHome, "sessions", "2026", "07", "01", "rollout-2026-07-01T00-00-00-"+staleSessionID+".jsonl")
	if err := os.MkdirAll(filepath.Dir(stalePath), 0755); err != nil {
		t.Fatalf("mkdir stale codex transcript dir: %v", err)
	}
	lines := []string{
		codexSessionMetaLine(staleSessionID, req.TempDir),
		codexAgentMessageLine(codexStaleCWDText),
	}
	if err := os.WriteFile(stalePath, []byte(strings.Join(lines, "\n")+"\n"), 0644); err != nil {
		t.Fatalf("write stale codex transcript: %v", err)
	}
	old := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(stalePath, old, old); err != nil {
		t.Fatalf("set stale codex transcript mtime: %v", err)
	}
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.CodexTTYCommand = fakeTUINoResumeUntilLate(5)
	writeStaleCodexTranscript(t, req)
	req.CodexTranscriptSchedules = []CodexTranscriptSchedule{
		{
			Delay: 200 * time.Millisecond,
			Lines: []string{
				codexSessionMetaLine(req.CodexTranscriptSessionID, req.TempDir),
			},
		},
		{
			Delay: 800 * time.Millisecond,
			Lines: []string{
				codexAgentMessageLine(codexCurrentCWDText),
			},
		},
	}
	req.StreamProbeSubstring = codexCurrentCWDText
	req.StreamProbeTimeout = 12 * time.Second
	req.ExecTimeout = 35 * time.Second
	return nil
}
```
