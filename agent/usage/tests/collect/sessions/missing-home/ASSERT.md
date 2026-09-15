## Expected

- Every provider reports `total: 0` with no `oldest`, no `newest` and no error, so
  a never-used provider looks the same in every snapshot: an empty session trend.
- Usage is still reported (the credentials are valid), which shows the session
  block is independent of the usage block.

## Side Effects

- None.

## Errors

- None: a missing session tree is not an error.

```go
import (
"testing"

"github.com/xhd2015/agent-pro/agent/usage"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
if err != nil {
t.Fatal(err)
}

for _, rec := range resp.Records {
AssertUsageOK(t, rec)
AssertSessions(t, rec, 0, "", "")
}
}
```
