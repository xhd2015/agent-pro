# Scenario

**Feature**: host `idle-probe-10s-verify-v10` chrome classifies idle

```
› reply with exactly: pong
› Run /review …gpt-5.6-terra medium · ~
  -> DetectScreenStatus
  -> idle
```

Crime scene had `screen=banner` on this exact tail.

## Steps

1. Load the host v10 snapshot fixture.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Fixture = fixtureHostV10
	return nil
}
```
