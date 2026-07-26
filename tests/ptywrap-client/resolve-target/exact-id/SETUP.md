# Scenario

**Feature**: ResolveTarget exact id match

```go
import (
	"testing"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "resolve-id"
	req.Target = "session-7"
	req.Sessions = []ptywrap.SessionInfo{
		{ID: "session-7", Name: "alpha", Status: "running", CreatedAt: time.Now()},
		{ID: "session-8", Name: "beta", Status: "running", CreatedAt: time.Now()},
	}
	return nil
}
```