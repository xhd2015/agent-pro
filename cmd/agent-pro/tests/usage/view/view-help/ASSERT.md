## Expected

- Exit 0 with a usage line, every flag this sub-command accepts, and the store layout
  it reads.
- The help states that the store is only read, which is the promise the dashboard
  rests on.
- The exit status section names the three fatal cases and says that an unparseable
  snapshot file is a warning instead, matching the leaves that exercise both.
- No `--provider`-style flag appears: view always shows every provider in the store.

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
for _, want := range []string{
"Usage: agent-pro usage view [options]",
"--port <n>",
"--open",
"--no-open",
"--since <dur>",
"--static-dir <d>",
"--json",
"primary_percent",
"sessions_total",
"read-only",
"$AGENT_PRO_HOME/usages/<provider>/<YYYY-MM-DD>/<HH-MM-SS>-snapshot.jsonl",
"0 on a clean shutdown",
"warning, not a failure",
} {
if !strings.Contains(resp.Stdout, want) {
t.Fatalf("help is missing %q:\n%s", want, resp.Stdout)
}
}
if strings.Contains(resp.Stdout, "--provider") {
t.Fatalf("help mentions --provider:\n%s", resp.Stdout)
}
if strings.TrimSpace(resp.Stderr) != "" {
t.Fatalf("stderr = %q, want nothing", resp.Stderr)
}
}
```
