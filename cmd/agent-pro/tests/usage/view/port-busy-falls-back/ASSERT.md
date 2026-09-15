## Expected

- The dashboard serves from a port other than the one requested, and the printed URL
  is the URL the harness reached: falling back is real, not cosmetic.
- stderr carries exactly one warning naming the busy port, so a user who pinned a
  port learns why they got another one.
- The warning is non-fatal: the server is up and `/api/healthz` answers `ok`, so a
  busy port never costs a working dashboard.

## Side Effects

- None: the port is bound, not written to.

## Errors

- None: a busy explicit port is a warning, not an error.

```go
import (
"fmt"
"strings"
"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
if err != nil {
t.Fatal(err)
}
AssertHTTPStatus(t, resp, 200)
if strings.TrimSpace(resp.HTTPBody) != "ok" {
t.Fatalf("healthz body = %q, want ok", resp.HTTPBody)
}

want := fmt.Sprintf("warning: port %d is busy, using ", req.ViewPort)
if !strings.Contains(resp.Stderr, want) {
t.Fatalf("stderr = %q, want %q", resp.Stderr, want)
}
if got := strings.Count(resp.Stderr, "warning: port"); got != 1 {
t.Fatalf("stderr has %d port warnings, want 1:\n%s", got, resp.Stderr)
}
if strings.Contains(req.ViewBaseURL, fmt.Sprintf(":%d", req.ViewPort)) {
t.Fatalf("serving URL %s is the busy port %d, want a fallback", req.ViewBaseURL, req.ViewPort)
}
if !strings.HasPrefix(req.ViewBaseURL, "http://127.0.0.1:") {
t.Fatalf("serving URL = %q, want the local listener", req.ViewBaseURL)
}
}
```
