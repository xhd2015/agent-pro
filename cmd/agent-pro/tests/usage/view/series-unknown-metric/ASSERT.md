## Expected

- 200, not a 404 and not a 500: an unknown metric is a legitimate question with an
  empty answer, which the page renders as "no data in range".
- Exactly one block, named after the request, with no series at all.
- The block still declares a kind and step flag, so the page does not have to guess
  how to draw an empty answer.

## Side Effects

- None.

## Errors

- None: an unknown metric is not an error (an unparseable `since` is a 400, covered
  by the package tests).

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
AssertNoWarnings(t, resp)

payload := Series(t, resp)
if len(payload.Metrics) != 1 {
t.Fatalf("metrics = %d blocks, want one", len(payload.Metrics))
}
block := payload.Metrics[0]
if block.Name != "not_a_metric" {
t.Fatalf("metric name = %q, want the requested not_a_metric", block.Name)
}
if len(block.Series) != 0 {
t.Fatalf("series = %+v, want none", block.Series)
}
if block.Kind == "" {
t.Fatalf("block = %+v, want a declared kind", block)
}
if !strings.Contains(resp.HTTPBody, `"series":[]`) {
t.Fatalf("body = %s, want an empty series array rather than null", resp.HTTPBody)
}
AssertStoreUnchanged(t, req)
}
```
