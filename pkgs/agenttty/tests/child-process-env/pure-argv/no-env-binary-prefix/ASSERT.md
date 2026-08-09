## Expected

- `err` is nil; `resp.Argv` non-empty.
- `resp.Argv[0]` is **not** `"env"`.
- `resp.Argv` starts with the pure agent command `codex`, `exec` (exact prefix).
- No `env` token at index 0 even when color + configHome would have required policy.

## Errors

- Returning `env -u NO_COLOR … codex exec` (legacy ApplyChildProcessEnv).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	if err != nil {
		t.Fatalf("ApplyChildProcessEnv harness error: %v", err)
	}
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if len(resp.Argv) == 0 {
		t.Fatal("Argv empty")
	}
	if resp.Argv[0] == "env" {
		t.Fatalf("argv must not start with env binary; got %#v", resp.Argv)
	}
	wantHead := req.Argv
	if len(wantHead) == 0 {
		wantHead = []string{"codex", "exec"}
	}
	if len(resp.Argv) < len(wantHead) {
		t.Fatalf("argv shorter than pure command: got %#v want head %#v", resp.Argv, wantHead)
	}
	for i := range wantHead {
		if resp.Argv[i] != wantHead[i] {
			t.Fatalf("argv pure prefix mismatch at [%d]: got %#v want head %#v", i, resp.Argv, wantHead)
		}
	}
}
```
