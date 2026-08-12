# Scenario

**Regression**: production open attach (OpenCloseExits → `attach_mode=open`)
must preserve child mouse-tracking CSI on the host TTY stream.

**Stack**: `llm-mock-run-grok` + real `grok` (crime scene parity).

```
keep-tty + llm-mock-run-grok + real grok enables mouse
  -> AttachWriter(mode=open) under host PTY
  -> DESIRED: host bytes contain mouse CSI
```

## Steps

1. Root builds llm-mock-run-grok + agent-run; skips if no `grok` on PATH.
2. HostAttachMode=open (production OpenCloseExits path).
3. Assert desired host fidelity.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.HostAttachMode = "open"
	req.SessionID = "llm-mock-open-mouse-1"
	return nil
}
```
