# Scenario

**Feature**: terminal status available while running and after finish (keep-tty)

```
POST codex-tty -> terminal available running -> wait finished -> terminal still available
```

## Steps

1. Create web codex-tty session.
2. Probe terminal status before and after finish.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	createWebCodexTTYSession(t, req, "keep tty availability")
	req.Mode = "http"
	return nil
}
```
