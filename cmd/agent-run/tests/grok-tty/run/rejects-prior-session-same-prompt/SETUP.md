# Scenario

**Feature**: strict discovery rejects a prior grok session with the same prompt

```
prior session dir (created_at 2h ago, first message "run ls")
  + new session dir appears after run start (same prompt)
  -> tail new session marker PRIOR_REJECT_NEW, not PRIOR_REJECT_OLD
```

## Steps

1. Seed old session with `created_at` two hours ago and prompt `run ls`.
2. After run starts, create new session with same prompt and stream new marker.
3. `Mode=stream-probe` waits for `PRIOR_REJECT_NEW`.

```go
import (
	"path/filepath"
	"testing"
	"time"
)

const (
	priorOldUUID   = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	priorNewUUID   = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	priorMarkerOld = "PRIOR_REJECT_OLD"
	priorMarkerNew = "PRIOR_REJECT_NEW"
)

func Setup(t *testing.T, req *Request) error {
	req.GrokHome = filepath.Join(req.TempDir, "grok-home")
	appendGrokHomeEnv(req)

	oldCreated := time.Now().Add(-2 * time.Hour)
	_ = writeFakeGrokSessionDirAt(t, req.GrokHome, req.TempDir, priorOldUUID, "run ls", oldCreated,
		acpAgentMessageChunk(priorMarkerOld))

	var newPath string
	req.GrokUpdatesSchedules = []GrokUpdatesSchedule{
		{
			Delay: 500 * time.Millisecond,
			OnFire: func() {
				newPath = writeFakeGrokSessionDirAt(t, req.GrokHome, req.TempDir, priorNewUUID, "run ls", time.Now())
			},
		},
		{
			Delay: 900 * time.Millisecond,
			OnFire: func() {
				if newPath != "" {
					_ = appendUpdatesJSONL(newPath, acpAgentMessageChunk(priorMarkerNew))
				}
			},
		},
	}

	req.GrokTTYCommand = fakeTUIHoldSeconds(5)
	appendGrokTTYEnv(req)
	req.Args = append(req.Args, "run ls")

	req.Mode = "stream-probe"
	req.StreamProbeSubstring = priorMarkerNew
	req.StreamProbeTimeout = 12 * time.Second
	return nil
}
```
