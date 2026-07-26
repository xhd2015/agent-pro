## Preconditions
- The mock config contains a nonzero exit code.

## Steps
1. Run fake Codex with the mock config.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-codex","stderr":"planned failure","exit_code":7,"llm_events":[{"type":"message","text":"before failure"}]}`)
    return nil
}
```

