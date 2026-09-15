## Expected

- The server starts and answers 200 on an empty and even missing store root: an
  empty store is an empty dashboard, not a failure.
- The store line says `0 providers, 0 snapshots` and names the root it looked at, so
  a user can see whether `AGENT_PRO_HOME` points where they expect.
- The payload carries empty arrays rather than nulls for providers, recent, metrics
  and errors, so the page can render its empty state without a null check.
- No warnings and nothing written to the store.

## Side Effects

- None: no directory is created, not even the store root.

## Errors

- None.

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
AssertServingURL(t, resp, req)
AssertStoreSummaryLine(t, resp, "0 providers, 0 snapshots")
AssertNoWarnings(t, resp)

summary := Summary(t, resp)
if summary.SnapshotCount != 0 || len(summary.Providers) != 0 || len(summary.Recent) != 0 {
t.Fatalf("empty store summary = %+v, want no providers and no records", summary)
}
for _, want := range []string{`"providers":[]`, `"recent":[]`, `"metrics":[]`, `"errors":[]`} {
if !strings.Contains(resp.HTTPBody, want) {
t.Fatalf("body is missing %s: %s", want, resp.HTTPBody)
}
}
if len(req.ViewFiles) != 0 {
t.Fatalf("no snapshot was planted, but req.ViewFiles = %v", req.ViewFiles)
}
}
```
