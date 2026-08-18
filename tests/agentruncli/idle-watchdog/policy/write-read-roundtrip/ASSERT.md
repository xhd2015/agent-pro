## Expected

- `IdlePolicyPath` is `home/sessions/<id>/idle-policy.json` (not `meta.json`).
- Write then Read: found=true, `ExitOnIdle=true`, timeout `10m`.
- File JSON field `idle_timeout` is compact `"10m"` (not `"10m0s"`).
- No API error.

## Side Effects

- Creates `idle-policy.json` under the per-leaf home (not `meta.json`).

## Errors

- None.

## Exit Code

N/A

```go
import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/agentrunapi"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	assertNoError(t, err)
	assertNoAPIError(t, resp)
	want := wantPolicyPath(req.Home, req.SessionID)
	if resp.Path != want {
		t.Fatalf("IdlePolicyPath: got %q, want %q", resp.Path, want)
	}
	if strings.Contains(resp.Path, "meta.json") {
		t.Fatalf("policy must not live in meta.json; path %q", resp.Path)
	}
	if !resp.Found || !resp.FileExists {
		t.Fatalf("expected policy file found; found=%v exists=%v", resp.Found, resp.FileExists)
	}
	if !resp.PolicyOn {
		t.Fatal("ExitOnIdle: got false, want true")
	}
	if resp.PolicyTimeout != 10*time.Minute || resp.PolicyTimeout != agentrunapi.DefaultIdleTimeout {
		t.Fatalf("IdleTimeout: got %s, want 10m", resp.PolicyTimeout)
	}
	var wire struct {
		ExitOnIdle  bool   `json:"exit_on_idle"`
		IdleTimeout string `json:"idle_timeout"`
	}
	if err := json.Unmarshal([]byte(resp.FileBody), &wire); err != nil {
		t.Fatalf("policy JSON: %v\n%s", err, resp.FileBody)
	}
	if !wire.ExitOnIdle {
		t.Fatalf("json exit_on_idle: got false; body %s", resp.FileBody)
	}
	if wire.IdleTimeout != "10m" {
		t.Fatalf("json idle_timeout: got %q, want 10m (compact); body %s", wire.IdleTimeout, resp.FileBody)
	}
}
```
