# Scenario

**Feature**: optional token becomes Bearer authorization on publish

```
NotifyTTYStarted(URL set, Token=secret)
  -> Publisher used with token (product may set Bearer on real HTTP client)
  -> inject path: opts.Token non-empty is accepted; publish still once
```

## Steps

1. Set URL and Token.
2. Use inject publisher; assert one successful publish and opts.Token wired
   (product EventBusOpts.Token must be passed through; real HTTP path sets
   `Authorization: Bearer <token>` via eventbus.WithToken).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.EventBusURL = "http://127.0.0.1:23891"
	req.EventBusToken = "test-token-n3"
	req.SessionID = "sess-notify-n3"
	req.UseInjectPublisher = true
	return nil
}
```
