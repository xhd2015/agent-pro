## Expected

- `Config()` succeeds with no error.
- All config fields are empty/zero: `DefaultAgentRunner`, `DefaultModel`, `LastSession`.

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	want := agentstorage.Config{}
	assertEqual(t, "DefaultAgentRunner", resp.Config.DefaultAgentRunner, want.DefaultAgentRunner)
	assertEqual(t, "DefaultModel", resp.Config.DefaultModel, want.DefaultModel)
	assertEqual(t, "LastSession", resp.Config.LastSession, want.LastSession)
}
```