## Expected Output

```
---
version: 3
---
--exit-on-idle\s+exit the keep-alive TTY after idle-timeout at a sendable prompt
\s+\(no-op unless --open, --detach, or --keep-tty actually keep a TTY\)
--idle-timeout DUR\s+idle window used with --exit-on-idle \(default: 10m\)
```

Help excerpt ends with a trailing newline. Full `RunHelpText()` body also ends
with `\n`.

## Expected

- `RunHelpText()` returns a non-empty help body (exit 0).
- Excerpt around the two idle flags matches the locked wording (flexible column
  spacing).
- Text contains `--exit-on-idle`, `--idle-timeout`, and default `10m`.
- Full help ends with trailing newline `\n`.

## Side Effects

- None.

## Errors

- None.

## Exit Code

0

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code: got %d, want 0", resp.ExitCode)
	}
	if strings.TrimSpace(resp.Stdout) == "" {
		t.Fatal("RunHelpText returned empty help")
	}
	for _, flag := range []string{"--exit-on-idle", "--idle-timeout"} {
		if !strings.Contains(resp.Stdout, flag) {
			t.Fatalf("run help must document %s; text:\n%s", flag, resp.Stdout)
		}
	}
	if !strings.Contains(resp.Stdout, "10m") {
		t.Fatalf("run help must document default 10m; text:\n%s", resp.Stdout)
	}
	if !strings.HasSuffix(resp.Stdout, "\n") {
		t.Fatal("run help text must end with trailing newline")
	}
	excerpt := idleHelpExcerpt(resp.Stdout)
	if excerpt == "" {
		t.Fatalf("run help has no --exit-on-idle / --idle-timeout block; text:\n%s", resp.Stdout)
	}
	assert.Output(t, excerpt, `---
version: 3
---
--exit-on-idle\s+exit the keep-alive TTY after idle-timeout at a sendable prompt
\s+\(no-op unless --open, --detach, or --keep-tty actually keep a TTY\)
--idle-timeout DUR\s+idle window used with --exit-on-idle \(default: 10m\)
`)
}
```
