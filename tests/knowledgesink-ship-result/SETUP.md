# Scenario

**Feature**: validate `result.json` `git_commit_files` object for create-mr ship

```
result.json + hub worktree
  -> knowledgesink.ReadValidateShipResult
  -> ShipResult | Error
```

## Preconditions

- Package `github.com/xhd2015/agent-pro/pkgs/knowledgesink` exports
  `ReadValidateShipResult`, `ShipResult`, `ShipCommitFiles`.
- Leaves set `ResultJSON`, hub seed, and expect OK vs error substring.
- Parallel-safe: temp dirs only; no process cwd/env mutation.

## Steps

1. Root Setup is a no-op (leaves own hub mode + JSON).
2. Run writes hub + result.json and calls `ReadValidateShipResult`.
3. Leaf Assert checks ship fields or error substring.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/knowledgesink"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = t
	_ = d
	_ = req
	return nil
}

func assertNoRunError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected Run error: %v", err)
	}
}

func assertOK(t *testing.T, req *Request, resp *Response, err error) *knowledgesink.ShipResult {
	t.Helper()
	assertNoRunError(t, err)
	if resp == nil {
		t.Fatal("nil response")
	}
	if resp.Err != "" {
		t.Fatalf("unexpected validate error: %s", resp.Err)
	}
	if resp.Ship == nil {
		t.Fatal("nil ship")
	}
	if !req.ExpectOK {
		t.Fatal("leaf marked ExpectOK=false but got success")
	}
	return resp.Ship
}

func assertErrContains(t *testing.T, req *Request, resp *Response, err error) {
	t.Helper()
	assertNoRunError(t, err)
	if resp == nil {
		t.Fatal("nil response")
	}
	if resp.Ship != nil {
		t.Fatalf("expected error, got ship %+v", resp.Ship)
	}
	if req.ExpectErrSubstr == "" {
		t.Fatal("ExpectErrSubstr empty")
	}
	if !strings.Contains(resp.Err, req.ExpectErrSubstr) {
		t.Fatalf("error %q does not contain %q", resp.Err, req.ExpectErrSubstr)
	}
}

func baseMsgBranch() (msg, branch string) {
	return "docs(kb): test", "tester/2026-03-24-ship-test"
}
```
