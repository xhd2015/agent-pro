# Scenario

**Feature**: ResolveTarget errors when name matches multiple sessions

```go
import (
	"testing"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "resolve-ambiguous"
	req.Target = "dup-name"
	req.Sessions = []ptywrap.SessionInfo{
		{ID: "session-3", Name: "dup-name", Status: "running", CreatedAt: time.Now()},
		{ID: "session-4", Name: "dup-name", Status: "exited", CreatedAt: time.Now()},
	}
	return nil
}
```