---
label: real-grok, e2e
explanation: Direct llm-mock-run-grok + real grok must crash on invalid UTF-8 argv within 3s (user-facing failure).
---

## Expected

- Process **exits within 3s** (no hang).
- Stderr/stdout contains real grok panic:
  - `panicked at library/std/src/env.rs`
  - and/or `called \`Result::unwrap()\``
- Exit code non-zero (panic / child failure).

## Errors

- Still running after 3s (timeout / deadline) → **FAIL** (did not crash fast).
- Clean exit 0 without panic → **FAIL**.

## Exit Code

non-zero

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	// Still running after 3s → fail (user crash is instant).
	if isTimeoutErr(err) {
		t.Fatalf("direct llm-mock-run-grok still running after %v; expected real grok env.rs crash within budget\nstdout:\n%s\nstderr:\n%s",
			invalidUTF8Budget, resp.Stdout, resp.Stderr)
	}

	combined := ""
	if resp != nil {
		combined = resp.Stdout + "\n" + resp.Stderr
	}
	if err != nil && !containsGrokEnvPanic(combined) {
		// exec error without panic body — still show it
		combined = combined + "\n" + err.Error()
	}

	if !containsGrokEnvPanic(combined) {
		t.Fatalf("expected real grok env.rs panic from llm-mock-run-grok argv within %v\nexit=%d err=%v\nstdout:\n%s\nstderr:\n%s",
			invalidUTF8Budget, resp.ExitCode, err, resp.Stdout, resp.Stderr)
	}
	if resp != nil && resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit after env.rs panic, got 0\n%s", combined)
	}
}
```
