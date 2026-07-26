---
label: e2e
---

## Expected

- Exit code 0.
- `OPENCODE_CONFIG_DIR` used equals explicit `LLM_MOCK_OPENCODE_CONFIG_DIR`.
- Orchestrator output or child env references `llm-mock` provider, `mock-model`, and mock `127.0.0.1` baseURL.

## Side Effects

- Explicit config dir and home directories used for opencode child isolation.

## Exit Code

0

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if resp.OpencodeConfigDirUsed != req.OpencodeConfigDir {
		t.Fatalf("OpencodeConfigDirUsed = %q, want explicit %q", resp.OpencodeConfigDirUsed, req.OpencodeConfigDir)
	}
	combined := resp.Stdout + resp.Stderr
	if !strings.Contains(combined, "127.0.0.1") {
		t.Fatalf("output missing mock localhost URL:\n%s", combined)
	}
	if !strings.Contains(combined, "llm-mock") && !strings.Contains(combined, "mock-model") {
		t.Fatalf("output missing llm-mock provider or mock-model:\n%s", combined)
	}
}
```