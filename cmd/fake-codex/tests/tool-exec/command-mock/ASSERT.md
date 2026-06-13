## Expected
- The event output contains `"fake codex output"` (mock value).
- The output does **not** contain `"REAL CODEX OUTPUT"` in the `aggregated_output` field (the command field naturally contains it, but the output should be mock).

```go
import (
    "encoding/json"
    "strings"
    "testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    stdout := resp.Stdout
    if !strings.Contains(stdout, "fake codex output") {
        t.Fatalf("expected mock output 'fake codex output', got:\n%s", stdout)
    }
    // The aggregated_output should be mock, not real command output
    // Parse the JSON to verify aggregated_output specifically
    lines := strings.Split(strings.TrimSpace(stdout), "\n")
    for _, line := range lines {
        if strings.TrimSpace(line) == "" {
            continue
        }
        var event map[string]any
        if err := json.Unmarshal([]byte(line), &event); err != nil {
            continue
        }
        if item, ok := event["item"].(map[string]any); ok {
            if output, ok := item["aggregated_output"].(string); ok {
                if strings.Contains(output, "REAL CODEX OUTPUT") {
                    t.Fatalf("mock should prevent real execution in aggregated_output, got: %q", output)
                }
            }
        }
    }
}
```
