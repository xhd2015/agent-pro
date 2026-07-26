# Scenario

**Feature**: pre-branch validation for `--auto-send-or-resume` (session-id required, mutex, help)

```
run --auto-send-or-resume [flags…]
  -> missing --session-id → exit 1
  -> + --session-id-from-prompt → exit 1 mutex
run -h -> documents --auto-send-or-resume
```

## Steps

1. Tag validation leaves with short exec timeout (flag parse only).

```go
import (
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Validation leaves hit flag gates only; keep timeouts tight.
	if req.ExecTimeout <= 0 {
		req.ExecTimeout = 15 * time.Second
	}
	return nil
}
```
