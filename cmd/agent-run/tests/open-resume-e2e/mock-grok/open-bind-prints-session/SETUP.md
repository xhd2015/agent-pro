# Scenario

**Feature**: open prints grok-tty session id and grok session uuid on stderr

```
open with mock → stderr contains "grok-tty:" and "grok session"
  + Paris in snapshot
```

## Steps

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Scenario = "open-bind-prints-session"
	req.SessionID = "e2e-open-bind-print"
	req.GrokSessionUUID = "b2222222-2222-4222-8222-222222222206"
	req.OpenPrompt = defaultOpenPrompt
	req.WantParis = defaultWantParis
	configureMockGrokEnv(t, req)
	return nil
}
```
