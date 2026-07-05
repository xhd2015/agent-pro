## Expected

- Non-zero exit before session starts.
- Stderr mentions both `--headless` and `--detach`.
- No registry entry created.

## Errors

- `run: --headless and --detach cannot be used together`

## Exit Code

1

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertNonZeroExit(t, resp)
	lower := strings.ToLower(resp.Combined)
	if !strings.Contains(lower, "--headless") {
		t.Fatalf("expected --headless in error, got combined %q", resp.Combined)
	}
	if !strings.Contains(lower, "--detach") {
		t.Fatalf("expected --detach in error, got combined %q", resp.Combined)
	}
	if !strings.Contains(lower, "cannot be used together") {
		t.Fatalf("expected mutual exclusion error, got combined %q", resp.Combined)
	}
	if len(resp.RegistryIDs) > 0 {
		t.Fatalf("expected no registry entry, got ids %v", resp.RegistryIDs)
	}
	if strings.HasPrefix(strings.TrimSpace(resp.Stdout), "session-id: ") {
		t.Fatalf("should not print session-id on mutual exclusion error, got stdout %q", resp.Stdout)
	}
}
```