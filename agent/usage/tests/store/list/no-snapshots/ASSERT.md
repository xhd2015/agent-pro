## Expected

- `List(0)` returns no records and no error: a file that is not a snapshot and a
  top-level file that is not a provider directory are both ignored.
- Nothing was written: the leaf appends nothing, so `Paths` stays empty.

## Side Effects

- None.

## Errors

- None: an unrecognized tree is empty, not broken.

```go
import (
"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
if err != nil {
t.Fatal(err)
}

if len(resp.Paths) != 0 {
t.Fatalf("wrote %v, want no appends", resp.Paths)
}
if got := resp.Stored[0]; len(got) != 0 {
t.Fatalf("List(0) returned %d records, want none", len(got))
}
}
```
