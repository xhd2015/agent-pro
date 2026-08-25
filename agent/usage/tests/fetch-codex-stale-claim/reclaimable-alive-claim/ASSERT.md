## Expected

- `usage.Fetch` returns without error despite a pre-planted alive
  `registry/.codex-status-usage.claim`.
- Snapshot matches the default Codex fake TUI fixture (`58%`, credits, reset).

## Side Effects

- Ephemeral ttywatch session under `req.TTYWatchHome` is torn down after fetch.
- Stale claim is cleared by reclaim (or by successful reserve after reclaim).

## Errors

- `run: session id "codex-status-usage" already in use` reproduces the Marcus
  menu-bar bug: `FetchStatus` reserved without reclaiming a reclaimable hold
  (headless `Run` already reclaim-once-and-retries).

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/agent/usage"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "already in use") {
			t.Fatalf("FetchStatus blocked by reclaimable leftover claim (Marcus menu bug): %v", err)
		}
		t.Fatal(err)
	}
	if resp == nil || resp.Snapshot == nil {
		t.Fatal("expected non-nil Snapshot")
	}
	assertSnapshotField(t, "Provider", string(resp.Snapshot.Provider), string(usage.Codex))
	assertSnapshotField(t, "UsagePercent", resp.Snapshot.UsagePercent, "58%")
	assertSnapshotField(t, "CreditsUsed", resp.Snapshot.CreditsUsed, "6519")
	assertSnapshotField(t, "CreditsTotal", resp.Snapshot.CreditsTotal, "11250")
	assertSnapshotField(t, "Reset", resp.Snapshot.Reset, "08:00 on 1 Aug")
}
```
