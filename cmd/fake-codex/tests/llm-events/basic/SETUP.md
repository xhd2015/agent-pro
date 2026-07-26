## Preconditions
- The mock config contains a `think` event in `llm_events`.

## Steps
1. Run fake codex with the mock config.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-codex","llm_events":[{"type":"think","text":"thinking in codex"}]}`)
    return nil
}
```
