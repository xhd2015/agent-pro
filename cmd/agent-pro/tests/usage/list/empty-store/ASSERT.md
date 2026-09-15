## Expected

- Exit 0 with `no usage snapshots under <path>` on stdout: an empty store is empty,
  not an error, and the message names the path that was read so the user can tell
  whether `AGENT_PRO_HOME` points where they expect.
- No table and no JSON: there is nothing to print.
- No snapshot file is created by listing.

## Side Effects

- None.

## Errors

- None.

```go
import (
"strings"
"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
if err != nil {
t.Fatal(err)
}
AssertExitCode(t, resp, 0)

if !strings.Contains(resp.Stdout, "no usage snapshots under") {
t.Fatalf("stdout = %q, want the empty-store message", resp.Stdout)
}
if !strings.Contains(resp.Stdout, "usages") {
t.Fatalf("stdout = %q, want the store path", resp.Stdout)
}
if lines := resp.StdoutLines(); len(lines) != 1 {
t.Fatalf("stdout has %d lines, want 1:\n%s", len(lines), resp.Stdout)
}
if resp.Stderr != "" {
t.Fatalf("stderr = %q, want empty", resp.Stderr)
}
AssertNoSnapshots(t, resp)
}
```
