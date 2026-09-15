## Expected

- stderr carries one `warning: usage view: skipping <path>` line naming the broken
  file and the parse failure, so corruption is visible without being fatal.
- 200 with the readable provider still present: one bad file costs one provider, not
  the dashboard.
- The payload's `errors` array names the broken path with the same parse failure, so
  the page can list what it skipped; its path is absolute, like the record paths.
- The store line reports the unreadable count, and the doctor message in the page
  footer is not the only place a user can learn about it.

## Side Effects

- None: the broken file is neither fixed nor deleted.

## Errors

- The broken file is the error case under test; the run itself stays at exit 0 and
  serves.

```go
import (
"strings"
"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
if err != nil {
t.Fatal(err)
}
AssertHTTPStatus(t, resp, 200)

if !strings.Contains(resp.Stderr, "warning: usage view: skipping ") {
t.Fatalf("stderr = %q, want a skipping warning", resp.Stderr)
}
AssertStoreSummaryLine(t, resp, "1 unreadable")

summary := Summary(t, resp)
if len(summary.Providers) != 1 || summary.Providers[0].Provider != "grok" {
t.Fatalf("providers = %+v, want grok only", summary.Providers)
}
if len(summary.Errors) != 1 {
t.Fatalf("errors = %+v, want the one unreadable file", summary.Errors)
}
broken := summary.Errors[0]
if !strings.HasSuffix(broken.Path, "01-02-03-snapshot.jsonl") {
t.Fatalf("error path = %q, want the broken snapshot file", broken.Path)
}
if !strings.Contains(resp.Stderr, broken.Path) {
t.Fatalf("stderr does not name %q:\n%s", broken.Path, resp.Stderr)
}
if !strings.Contains(broken.Error, "parse") {
t.Fatalf("error = %q, want the parse failure", broken.Error)
}
if len(summary.Metrics) == 0 {
t.Fatal("metrics is empty, so the surviving provider would not be chartable")
}
}
```
