## Preconditions
- The plugin is installed.

## Steps
1. Prepare the plugin file, then run `agent-hub integration uninstall opencode`.

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
