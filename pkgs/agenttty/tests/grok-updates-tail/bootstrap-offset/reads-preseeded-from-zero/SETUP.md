# Scenario

**Feature**: `startOffset=0` replays pre-seeded lines present before watch starts

```
updates.jsonl pre-seeded with user + assistant
  -> TailUpdatesFromOffset(ctx, path, 0, emit)
  -> both texts in EventTexts from bootstrap read
```

## Steps

1. Write user + assistant lines before starting tail.
2. Set `StartOffset=0`; no scheduled appends.
3. Assert both texts appear without waiting for new appends.

```go
import (
	"testing"
	"time"
)

const (
	bootstrapUserText      = "bootstrap user from zero"
	bootstrapAssistantText = "BOOTSTRAP_ASSISTANT_FROM_ZERO"
)

func Setup(t *testing.T, req *Request) error {
	req.StartOffset = 0
	req.InitialLines = []string{
		acpUserMessageChunk(bootstrapUserText),
		acpAgentMessageChunk(bootstrapAssistantText),
	}
	req.TailStartDelay = 100 * time.Millisecond
	req.HoldAfterSchedule = 200 * time.Millisecond
	return nil
}
```