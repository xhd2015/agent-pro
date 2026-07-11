# Scenario

**Feature**: `agent-run resume` — gate checks then run shortcut with provider `--resume`

```
seed meta -> agent-run resume [flags] <session-id> ["followup"]
  denied: not exited | unbound | missing | no prompt
  success: headless followup when exited (argv includes --resume <id>)
  --open accepted as known flag
```

## Steps

1. Leaf seeds meta / live fixtures and sets `req.Args` for resume.
2. `Run` executes CLI; assert gate errors or argv/session success.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	// Leaves finalize seed + Args for resume invocations.
	if len(req.Args) == 0 {
		req.Args = []string{"resume"}
	}
	return nil
}
```
