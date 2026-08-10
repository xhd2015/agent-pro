## Expected

- Result preserves BaseArgs as a prefix (or contains the same tokens).
- Contains `--event-bus-url` followed by the URL value (adjacent argv pair or
  `--event-bus-url=URL` form).
- Contains `--event-bus-token` followed by the token value (or `=` form).
- BaseArgs content is not dropped.

## Errors

- None.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	args := resp.ResultArgs
	if len(args) < len(req.BaseArgs) {
		t.Fatalf("result shorter than base: %#v", args)
	}
	// All base tokens must appear in order as a prefix (append-only contract).
	for i, b := range req.BaseArgs {
		if args[i] != b {
			t.Fatalf("base prefix mismatch at %d: got %q want %q; full %#v", i, args[i], b, args)
		}
	}
	joined := strings.Join(args, "\x00")
	if !hasFlagValue(args, "--event-bus-url", req.EventBusURL) {
		t.Fatalf("missing --event-bus-url %q in %#v (joined scan %q)", req.EventBusURL, args, joined)
	}
	if !hasFlagValue(args, "--event-bus-token", req.EventBusToken) {
		t.Fatalf("missing --event-bus-token %q in %#v", req.EventBusToken, args)
	}
}

// hasFlagValue accepts either ["--flag", "value"] or ["--flag=value"].
func hasFlagValue(args []string, flag, value string) bool {
	eq := flag + "=" + value
	for i, a := range args {
		if a == eq {
			return true
		}
		if a == flag && i+1 < len(args) && args[i+1] == value {
			return true
		}
	}
	return false
}
```
