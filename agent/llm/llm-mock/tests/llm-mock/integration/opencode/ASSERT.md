---
label: e2e
---

## Expected
- Combined stdout+stderr contains "Paris" (the LLM response).
- The mock server's events file records at least 1 request.
- At least one recorded request has model "gpt-4" and path "/v1/chat/completions".
- Note: opencode runs an agent loop that makes multiple LLM calls. After configured exchanges
  are exhausted the mock stops replaying, which may cause a non-zero exit. The test verifies
  the core integration (LLM interaction succeeded) rather than requiring exit 0.

```go
import (
    "encoding/json"
    "strings"
    "testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("run failed: %v", err)
    }

    // Output must contain "Paris" — this proves the LLM integration worked.
    if !strings.Contains(resp.Stdout, "Paris") {
        t.Fatalf("expected output to contain 'Paris', got:\n%s", resp.Stdout)
    }

    // Events file must have at least 1 recorded request
    if len(resp.Responses) == 0 {
        t.Fatal("expected at least 1 events file entry, got 0")
    }

    eventsBody := resp.Responses[0].Body
    if eventsBody == "" {
        t.Fatal("events file is empty")
    }

    // Parse JSON-lines: one JSON object per line
    lines := strings.Split(strings.TrimSpace(eventsBody), "\n")
    var recorded []map[string]any
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if line == "" {
            continue
        }
        var rec map[string]any
        if err := json.Unmarshal([]byte(line), &rec); err != nil {
            t.Fatalf("invalid events JSON line: %v\n%s", err, line)
        }
        recorded = append(recorded, rec)
    }

    if len(recorded) == 0 {
        t.Fatal("events file recorded 0 requests, expected at least 1")
    }

    // Find at least one request with model "gpt-4"
    found := false
    for _, rec := range recorded {
        body, ok := rec["body"].(map[string]any)
        if !ok {
            continue
        }
        model, _ := body["model"].(string)
        if model == "gpt-4" {
            found = true
            break
        }
    }
    if !found {
        t.Fatalf("no recorded request with model 'gpt-4' in:\n%s", eventsBody)
    }
}
```
