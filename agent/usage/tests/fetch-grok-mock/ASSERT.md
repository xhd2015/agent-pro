## Expected

- `usage.Fetch` returns without error.
- `resp.Snapshot.Provider` is `grok`.
- `resp.Snapshot.UsagePercent` is `1%`.
- `resp.Snapshot.Reset` is `July 9, 16:55 PT`.
- `resp.Snapshot.CreditsUsed` and `CreditsTotal` are empty (Grok has no credit fields).

## Side Effects

- Ephemeral PTY session only (no persistent registry).

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
	assertSnapshotField(t, "Provider", string(resp.Snapshot.Provider), string(usage.Grok))
	assertSnapshotField(t, "UsagePercent", resp.Snapshot.UsagePercent, "1%")
	assertSnapshotField(t, "Reset", resp.Snapshot.Reset, "July 9, 16:55 PT")
	assertSnapshotField(t, "CreditsUsed", resp.Snapshot.CreditsUsed, "")
	assertSnapshotField(t, "CreditsTotal", resp.Snapshot.CreditsTotal, "")
}
```