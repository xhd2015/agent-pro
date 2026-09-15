## Expected

- Exit 0 with the three sub-commands on stdout, `view` among them with its one-line
  description: the dashboard is discoverable from `agent-pro usage --help`.
- stderr stays empty: help is a result, not a warning.
- The help does not run anything: no snapshot file appears.

## Side Effects

- None.

## Errors

- None: help exits 0.

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
for _, line := range []string{
"collect   collect each provider's usage into ~/.agent-pro/usages (snapshots)",
"list      list stored usage snapshots",
"view      serve a read-only dashboard of the stored snapshots",
} {
if !strings.Contains(resp.Stdout, line) {
t.Fatalf("help does not list %q:\n%s", line, resp.Stdout)
}
}
if strings.TrimSpace(resp.Stderr) != "" {
t.Fatalf("stderr = %q, want nothing", resp.Stderr)
}
AssertNoSnapshots(t, resp)
}
```
