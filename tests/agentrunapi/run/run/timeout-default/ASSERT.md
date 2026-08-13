## Expected

- Launch sees Timeout == `DefaultRunTimeout` (30m).

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/agentrunapi"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertNoAPIError(t, resp)
	assertEqual(t, "LaunchTimeout", resp.LaunchTimeout, agentrunapi.DefaultRunTimeout)
}
```
