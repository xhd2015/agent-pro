# Scenario

**Feature**: grok/codex last coalesced message

```
last kind + live -> activity
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Source = "chat"
	return nil
}
```
