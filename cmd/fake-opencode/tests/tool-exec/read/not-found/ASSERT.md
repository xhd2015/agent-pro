## Expected
- The event has a non-empty `error` field or a non-zero exit_code, indicating the file was not found.

```go
import (
    "testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    events := parseJSONLines(t, resp.Stdout)
    if len(events) == 0 {
        t.Fatal("no events in stdout")
    }
    event := events[0]
    part, _ := event["part"].(map[string]any)
    state, _ := part["state"].(map[string]any)

    errorStr, hasError := state["error"].(string)
    exitCode, hasExit := state["exit_code"]

    if hasError && errorStr != "" {
        return // error field present — correct
    }
    if hasExit {
        if code, ok := exitCode.(float64); ok && int(code) != 0 {
            return // non-zero exit — correct
        }
    }
    t.Fatalf("expected error or non-zero exit for missing file, state: %v", state)
}
```
