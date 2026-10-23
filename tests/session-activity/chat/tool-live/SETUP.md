# Scenario

**Feature**: last tool message while live is tool_calling

```
tool(bash) + live -> tool_calling
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.LastKind = "tool"
	req.Tool = "bash"
	req.Live = true
	return nil
}
```
