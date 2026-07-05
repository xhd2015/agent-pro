## Expected

- Exit code 0.
- Argv probe file contains `ARGV_RECORD=`.
- Argv record includes grok-tty defaults `--always-approve` and
  `--permission-mode=bypassPermissions`.

## Exit Code

0

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)
	probe, err := os.ReadFile(req.ArgvProbePath)
	if err != nil {
		t.Fatalf("read argv probe %s: %v", req.ArgvProbePath, err)
	}
	record := strings.TrimSpace(string(probe))
	assertContains(t, record, "ARGV_RECORD=")
	assertContains(t, record, "--always-approve")
	assertContains(t, record, "--permission-mode=bypassPermissions")
}

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("missing %q in:\n%s", want, got)
	}
}
```