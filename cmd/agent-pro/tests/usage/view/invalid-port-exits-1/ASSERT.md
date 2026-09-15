## Expected

- Exit 1 with `agent-pro: invalid --port -1: must be between 0 and 65535` on stderr:
  the message quotes the value the user typed, so a typo is obvious.
- Empty stdout: nothing was served, so there is no URL line to parse and no store
  line to read.
- No snapshot is written and no store directory is created.

## Side Effects

- None.

## Errors

- The invalid port itself: reported on stderr with exit 1.

```go
import (
"strings"
"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
if err != nil {
t.Fatal(err)
}
AssertExitCode(t, resp, 1)
AssertStderrContains(t, resp, "invalid --port -1")
AssertStderrContains(t, resp, "must be between 0 and 65535")
if strings.TrimSpace(resp.Stdout) != "" {
t.Fatalf("stdout = %q, want nothing on a rejected flag", resp.Stdout)
}
AssertNoSnapshots(t, resp)
}
```
