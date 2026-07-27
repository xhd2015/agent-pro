# Scenario

**Subcommand group**: `send` against live codex-tty with real UI + mock LLM

```
open idle → send → submitted turn (or draft with --no-submit)
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
