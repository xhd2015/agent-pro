# Scenario

**Subcommand group**: trust / open readiness with real Codex UI + mock LLM

```
run --open → trust cleared → idle › sendable
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = req
	return nil
}
```
