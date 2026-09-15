## Expected

- `List(0)` returns no records and no error, exactly like a root that exists but
  holds nothing: a missing store is empty, and listing never creates it.
- The root path is still absent afterwards, so a read-only command leaves no trace.

## Side Effects

- None.

## Errors

- None.

```go
import (
"os"
"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
if err != nil {
t.Fatal(err)
}

if got := resp.Stored[0]; len(got) != 0 {
t.Fatalf("List(0) returned %d records, want none", len(got))
}
if _, err := os.Stat(req.Root); !os.IsNotExist(err) {
t.Fatalf("listing root %s: stat err = %v, want the root to stay absent", req.Root, err)
}
}
```
