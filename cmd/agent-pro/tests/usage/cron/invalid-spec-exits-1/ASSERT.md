## Expected

- Exit 1 and no collection: the spec is parsed before the first cycle, so a
  mistyped schedule cannot leave a snapshot behind (a scheduler would retry
  forever against a loop that only ever ran once).
- The message names the offending spec and the accepted forms:
  `invalid --cron spec "every 5 min"`, quoting the raw value so the typo is visible.
- stdout stays empty.

## Side Effects

- None: no snapshot file is created.

## Errors

- Reported as `agent-pro: invalid --cron spec "every 5 min": ...` on stderr.

```go
import (
"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
if err != nil {
t.Fatal(err)
}
AssertExitCode(t, resp, 1)

AssertStderrContains(t, resp, `agent-pro: invalid --cron spec "every 5 min"`)
AssertStderrContains(t, resp, "expected 5 fields")
if resp.Stdout != "" {
t.Fatalf("stdout = %q, want empty", resp.Stdout)
}
AssertNoSnapshots(t, resp)
}
```
