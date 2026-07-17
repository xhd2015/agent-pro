## Expected

- Handle returns nil (`ErrString` empty).
- Stdout contains usage and lists `web`, `run`, `sessions`, `status`, and `--agent-runner`.

## Expected Output

Key tokens (substring contract; full help layout remains implementer-owned as
long as zero intentional behavior change vs current agent-run help):

```text
Usage:
web
run
sessions
status
--agent-runner
```

Stdout ends with a trailing newline when help is printed as today.

## Side Effects

- None beyond captured stdout/stderr.

## Errors

- No harness error; no Handle error.

## Exit Code

N/A (package call; success is nil error)

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if resp == nil {
		t.Fatal("nil response")
	}
	assertNoHandleError(t, resp)
	if resp.Stdout == "" {
		t.Fatal("expected non-empty help on stdout")
	}
	// Substring contract aligned with cmd/agent-run/tests/help/top-level.
	// Full bounded assert.Output is left to existing CLI help trees.
	for _, token := range []string{
		"Usage:",
		"web",
		"run",
		"sessions",
		"status",
		"--agent-runner",
	} {
		assertContains(t, resp.Stdout, token)
	}
}
```
