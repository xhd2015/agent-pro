# Scenario

**Feature**: terminal websocket forwards resize/control messages

```
browser sends {"type":"resize","cols":100,"rows":32} -> ptywrap receives resize message
```

## Preconditions

- Parent setup created live tty session and registry.

## Steps

1. Attach to terminal websocket.
2. Send JSON resize message before input.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.WSResizeJSON = `{"type":"resize","cols":100,"rows":32}`
	req.WSInput = ""
	return nil
}
```
