## Expected

- `usage.Fetch` returns without error.
- `resp.Snapshot.Provider` is `codex`.
- `resp.Snapshot.UsagePercent` is `58%` (parsed from `42% left`).
- `resp.Snapshot.CreditsUsed` is `6519`.
- `resp.Snapshot.CreditsTotal` is `11250`.
- `resp.Snapshot.Reset` is `08:00 on 1 Aug`.

## Side Effects

- Ephemeral ttywatch session under `req.TTYWatchHome` is torn down after fetch.

## Errors

- None from `Run`.

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/agent/usage"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.Snapshot == nil {
		t.Fatal("expected non-nil Snapshot")
	}
	assertSnapshotField(t, "Provider", string(resp.Snapshot.Provider), string(usage.Codex))
	assertSnapshotField(t, "UsagePercent", resp.Snapshot.UsagePercent, "58%")
	assertSnapshotField(t, "CreditsUsed", resp.Snapshot.CreditsUsed, "6519")
	assertSnapshotField(t, "CreditsTotal", resp.Snapshot.CreditsTotal, "11250")
	assertSnapshotField(t, "Reset", resp.Snapshot.Reset, "08:00 on 1 Aug")
}
```