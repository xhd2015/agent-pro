## Expected

- Exit code 1.
- Stderr contains `failed to load config` and `slack-config.json`.
- Stdout empty.

## Exit Code

1

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 1)
	if resp.Stdout != "" {
		t.Fatalf("expected empty stdout, got:\n%s", resp.Stdout)
	}
	assertStderrContains(t, resp, "failed to load config")
	assertStderrContains(t, resp, "slack-config.json")
}
```