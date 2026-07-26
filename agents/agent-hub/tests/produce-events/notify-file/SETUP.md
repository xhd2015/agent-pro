## Preconditions
- A valid JSON event is written to a temp file.

## Steps
1. Write event JSON to file.
2. Run `agent-hub notify --file <path>`.

```go
import (
    "testing"
    "path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    path := filepath.Join(req.TempDir, "event.json")
    writeFile(t, path, `{"event_type":"agent.session.started","runner":"fake-opencode","runner_session_id":"s-file"}`)
    req.Args = []string{"notify", "--file", path}
    return nil
}
```
