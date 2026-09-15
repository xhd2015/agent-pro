## Expected

- 200 with both providers, so the server really read the store.
- The store tree still holds exactly the two planted files with exactly the lines
  they were written with: same paths, same bytes, no new file, no appended line and
  no extra directory.
- A later `usage list` therefore sees only what `collect` wrote, and the snapshot
  files stay a record of collection cycles rather than of viewing sessions.
- No warnings on stderr.

## Side Effects

- None by design: this leaf exists to prove that.

## Errors

- None.

```go
import (
"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
if err != nil {
t.Fatal(err)
}
AssertHTTPStatus(t, resp, 200)
AssertNoWarnings(t, resp)

summary := Summary(t, resp)
if len(summary.Providers) != 2 || summary.SnapshotCount != 2 {
t.Fatalf("summary = %d providers over %d snapshots, want the 2 planted ones",
len(summary.Providers), summary.SnapshotCount)
}
if len(req.ViewFiles) != 2 {
t.Fatalf("the leaf planted %d snapshots, want 2", len(req.ViewFiles))
}
AssertStoreUnchanged(t, req)
}
```
