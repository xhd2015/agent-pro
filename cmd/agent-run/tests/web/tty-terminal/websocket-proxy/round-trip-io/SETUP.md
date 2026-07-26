# Scenario

**Feature**: terminal websocket proxies PTY output and browser input

```
browser WS attach -> fake ptywrap sends terminal-ready
browser sends "hello from browser\n" -> fake ptywrap echoes bytes
```

## Preconditions

- Parent setup created live tty session and registry.

## Steps

1. Attach to terminal websocket.
2. Send binary input including Enter.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.WSInput = "hello from browser\n"
	return nil
}
```
