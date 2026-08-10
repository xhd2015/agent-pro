# Scenario

**Feature**: publish failure writes `warning:` and does not fail the caller

```
NotifyTTYStarted + Publisher returns error (non-2xx / transport)
  -> WarnWriter has warning: …
  -> Run err is nil (open still succeeds; best-effort)
```

## Steps

1. Set URL.
2. Inject publisher that fails.
3. Assert warning prefix; no fatal error.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.EventBusURL = "http://127.0.0.1:23891"
	req.EventBusToken = ""
	req.UseInjectPublisher = true
	req.InjectPublishFail = true
	return nil
}
```
