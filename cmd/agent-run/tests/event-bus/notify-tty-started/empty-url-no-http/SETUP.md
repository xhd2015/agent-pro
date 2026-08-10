# Scenario

**Feature**: empty event-bus URL skips publish (no HTTP)

```
NotifyTTYStarted(URL="") -> no HTTP; no warning
```

## Steps

1. Leave `EventBusURL` empty.
2. Provide inject publisher that would record if called (must stay idle).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.EventBusURL = ""
	req.EventBusToken = ""
	// Publisher still injected; empty URL must not call Publish.
	req.UseInjectPublisher = true
	return nil
}
```
