---
label: e2e
---

## Expected
- The event indicates success (status "completed").
- The file exists on disk and its content matches `"written content for verification"`.

```go
import (
    "os"
    "testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    events := parseJSONLines(t, resp.Stdout)
    if len(events) == 0 {
        t.Fatal("no events in stdout")
    }
    event := events[0]
    part, _ := event["part"].(map[string]any)
    state, _ := part["state"].(map[string]any)
    if status, _ := state["status"].(string); status != "completed" {
        t.Fatalf("expected status 'completed', got: %q", status)
    }

    // Verify the file was actually created
    targetPath := req.TempDir + "/write-output.txt"
    data, readErr := os.ReadFile(targetPath)
    if readErr != nil {
        t.Fatalf("write should create file at %s: %v", targetPath, readErr)
    }
    content := string(data)
    if content != "written content for verification" {
        t.Fatalf("expected written content 'written content for verification', got: %q", content)
    }
}
```
