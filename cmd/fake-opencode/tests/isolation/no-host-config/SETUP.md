## Preconditions
- The command runs with a temporary HOME.

## Steps
1. Run fake opencode with an empty mock config.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","llm_events":[]}`)
    return nil
}
```

