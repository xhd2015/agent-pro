# Scenario

**Feature**: ResolveTarget unique name match

```go
import (
	"testing"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "resolve-name"
	req.Target = "my-grok"
	req.Sessions = []ptywrap.SessionInfo{
		{ID: "session-1", Name: "other", Status: "running", CreatedAt: time.Now()},
		{ID: "session-2", Name: "my-grok", Status: "running", CreatedAt: time.Now()},
	}
	return nil
}
```