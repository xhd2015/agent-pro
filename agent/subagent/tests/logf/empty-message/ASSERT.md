## Expected
- Output matches `[YYYY-MM-DDTHH:MM:SS] \n` (timestamp prefix, space, exactly one newline).
- The output contains the timestamp even though the message is empty.

```go
import (
    "regexp"
    "strings"
    "testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("run failed: %v", err)
    }

    tsPattern := regexp.MustCompile(`^\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\] \n$`)
    if !tsPattern.MatchString(resp.Stdout) {
        t.Fatalf("expected timestamp with empty message and one newline, got:\n%q", resp.Stdout)
    }

    if strings.Count(resp.Stdout, "\n") != 1 {
        t.Fatalf("expected exactly one newline, got %d in:\n%q", strings.Count(resp.Stdout, "\n"), resp.Stdout)
    }
}
```
