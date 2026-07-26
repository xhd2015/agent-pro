# Scenario

**Feature**: after /exit, status shows exited true and resume.ready yes

```
open Paris → /exit → status
  -> exited:true; resume.ready:yes (no resume step)
```

## Steps

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Scenario = "status-ready-after-exit"
	req.SessionID = "e2e-status-ready"
	req.GrokSessionUUID = "b2222222-2222-4222-8222-222222222208"
	req.OpenPrompt = defaultOpenPrompt
	req.WantParis = defaultWantParis
	req.SkipResume = true
	configureMockGrokEnv(t, req)
	return nil
}
```
