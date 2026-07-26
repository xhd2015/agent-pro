## Expected
- Output matches `[YYYY-MM-DDTHH:MM:SS] value=foo count=42\n`.
- Format verbs are resolved with the provided arguments.

```go
import (
    "regexp"
    "testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("run failed: %v", err)
    }

    tsPattern := regexp.MustCompile(`^\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\] value=foo count=42\n$`)
    if !tsPattern.MatchString(resp.Stdout) {
        t.Fatalf("expected timestamped 'value=foo count=42' with one newline, got:\n%q", resp.Stdout)
    }
}
```
