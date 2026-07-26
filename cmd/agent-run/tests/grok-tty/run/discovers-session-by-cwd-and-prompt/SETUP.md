# Scenario

**Feature**: grok session discovery matches encoded cwd + first user_message_chunk prompt

```
two grok session dirs under same encoded cwd
  -> only session with user_message_chunk "run ls" is tailed
  -> stdout shows DISCOVERY_MATCH_B, not DISCOVERY_MATCH_A
```

## Steps

1. Point `GROK_HOME` at temp dir; do **not** set `AGENT_RUN_GROK_TTY_GROK_SESSION_ID`.
2. After banner, create two session dirs (wrong prompt vs `run ls`).
3. Append assistant markers to each session; only the matching session should stream.
4. `Mode=stream-probe` waits for `DISCOVERY_MATCH_B`.

```go
import (
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
	"path/filepath"
)

const (
	discoveryWrongUUID = "11111111-1111-1111-1111-111111111111"
	discoveryRightUUID = "22222222-2222-2222-2222-222222222222"
	discoveryMarkerA   = "DISCOVERY_MATCH_A"
	discoveryMarkerB   = "DISCOVERY_MATCH_B"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.GrokHome = filepath.Join(req.TempDir, "grok-home")
	appendGrokHomeEnv(req)

	var wrongPath, rightPath string
	req.GrokUpdatesSchedules = []GrokUpdatesSchedule{
		{
			Delay: 400 * time.Millisecond,
			OnFire: func() {
				wrongPath = writeFakeGrokSessionDir(t, req.GrokHome, req.TempDir, discoveryWrongUUID, "say wrong")
				rightPath = writeFakeGrokSessionDir(t, req.GrokHome, req.TempDir, discoveryRightUUID, "run ls")
			},
		},
		{
			Delay: 900 * time.Millisecond,
			OnFire: func() {
				if wrongPath != "" {
					_ = appendUpdatesJSONL(wrongPath, acpAgentMessageChunk(discoveryMarkerA))
				}
				if rightPath != "" {
					_ = appendUpdatesJSONL(rightPath, acpAgentMessageChunk(discoveryMarkerB))
				}
			},
		},
	}

	req.GrokTTYCommand = fakeTUIHoldSeconds(5)
	appendGrokTTYEnv(req)
	req.Args = append(req.Args, "run ls")

	req.Mode = "stream-probe"
	req.StreamProbeSubstring = discoveryMarkerB
	req.StreamProbeTimeout = 12 * time.Second
	return nil
}
```