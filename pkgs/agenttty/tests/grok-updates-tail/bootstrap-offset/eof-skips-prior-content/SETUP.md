# Scenario

**Feature**: `startOffset=EOF` skips stale on-disk content (no replay)

```
updates.jsonl contains STALE_EOF_SKIP_MARKER before tail starts
  -> TailUpdatesFromOffset(ctx, path, fileSize, emit)
  -> stale marker NOT in EventTexts
```

## Steps

1. Pre-seed file with stale marker line.
2. Set `StartOffsetAtEOF=true` (offset = current file size).
3. Do not append new lines; cancel tail and assert stale content absent.

```go
import (
	"testing"
	"time"
)

const staleEOFSkipMarker = "STALE_EOF_SKIP_MARKER"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.StartOffsetAtEOF = true
	req.InitialLines = []string{
		acpUserMessageChunk(staleEOFSkipMarker),
		acpAgentMessageChunk("stale assistant should not replay"),
	}
	req.TailStartDelay = 100 * time.Millisecond
	req.HoldAfterSchedule = 200 * time.Millisecond
	return nil
}
```