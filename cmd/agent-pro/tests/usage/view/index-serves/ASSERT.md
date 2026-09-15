## Expected

- The first stdout line is `serving http://127.0.0.1:<port>` and the harness probed
  exactly that URL, so the printed URL is the URL that works.
- The store line reports the store read-only, with one provider and one snapshot:
  the dashboard's own log says what it is reading.
- `GET /` answers 200 with `text/html` and the dashboard markup: the chart
  container and the two API paths the page polls.
- No warning on stderr: a healthy store warns about nothing.

## Side Effects

- None: serving reads the store, the page is the embedded copy.

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
AssertStoreSummaryLine(t, resp, "read-only, 1 provider, 1 snapshot")
AssertNoWarnings(t, resp)

if !strings.Contains(resp.HTTPBody, "<title>agent-pro usage</title>") {
t.Fatalf("page is missing the dashboard title:\n%s", firstLines(resp.HTTPBody, 6))
}
for _, marker := range []string{`id="chart"`, `id="cards"`, "/api/summary", "/api/series"} {
if !strings.Contains(resp.HTTPBody, marker) {
t.Fatalf("page is missing %q", marker)
}
}
AssertStoreUnchanged(t, req)
}

func firstLines(body string, count int) string {
lines := strings.Split(body, "\n")
if len(lines) > count {
lines = lines[:count]
}
return strings.Join(lines, "\n")
}
```
