## Preconditions
- The plugin is installed.

## Steps
1. Prepare the plugin, then run `agent-hub integration disable opencode`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    _ = t
    return nil
}
```
