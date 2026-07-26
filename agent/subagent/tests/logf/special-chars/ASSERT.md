## Expected
- Output starts with `[YYYY-MM-DDTHH:MM:SS] ` followed by the multiline message.
- Special characters (tabs, newlines) are preserved.
- Exactly one trailing newline at the very end.

```go
import (
    "regexp"
    "strings"
    "testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("run failed: %v", err)
    }

    tsPattern := regexp.MustCompile(`^\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\] `)
    if !tsPattern.MatchString(resp.Stdout) {
        t.Fatalf("expected timestamp prefix, got:\n%q", resp.Stdout)
    }

    if !strings.Contains(resp.Stdout, "line1") {
        t.Fatalf("expected 'line1' in output, got:\n%q", resp.Stdout)
    }
    if !strings.Contains(resp.Stdout, "line2\tindented") {
        t.Fatalf("expected 'line2\\tindented' in output, got:\n%q", resp.Stdout)
    }
    if !strings.Contains(resp.Stdout, "line3") {
        t.Fatalf("expected 'line3' in output, got:\n%q", resp.Stdout)
    }

    if !strings.HasSuffix(resp.Stdout, "\n") {
        t.Fatalf("expected trailing newline, got:\n%q", resp.Stdout)
    }
}
```
