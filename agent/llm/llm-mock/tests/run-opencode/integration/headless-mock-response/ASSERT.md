---
label: e2e, real-opencode, slow
explanation: Requires real opencode CLI on PATH; headless LLM round-trip up to 60s; exit 0 not required.
---

## Expected

- Combined stdout/stderr contains `"Paris"` (mocked LLM response).
- HTTP log at `LogHTTPPath` has at least 1 exchange with `request.path` containing `/v1/chat/completions`.
- At least one logged request has model `llm-mock/mock-model` or `mock-model`.
- Exit 0 is **not** required (opencode agent loop may continue after exchanges exhaust).

## Side Effects

- Opencode session artifacts may be written under isolated config dir (not asserted in MVP).

## Exit Code

not required (any)

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil && resp.ExitCode == 0 {
		t.Fatal(err)
	}

	combined := resp.Stdout + resp.Stderr
	assertContains(t, combined, "Paris")

	if len(resp.LogHTTPLines) < 1 {
		t.Fatalf("log-http file: want >=1 JSONL line, got %d\ncontent:\n%s",
			len(resp.LogHTTPLines), resp.LogHTTPContent)
	}
	records, parseErr := parseHTTPExchangeMaps(resp.LogHTTPLines)
	if parseErr != nil {
		t.Fatal(parseErr)
	}
	if !httpLogHasChatCompletionsModel(records, "llm-mock/mock-model", "mock-model") {
		t.Fatalf("no recorded /v1/chat/completions request with model llm-mock/mock-model or mock-model in:\n%s",
			resp.LogHTTPContent)
	}
}
```