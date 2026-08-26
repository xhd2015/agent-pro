# Scenario

**Feature**: unknown `--agent-runner` is rejected

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"--agent-runner", "nope", "hello"}
	return nil
}
```
