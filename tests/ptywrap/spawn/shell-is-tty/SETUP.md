# Scenario

**Feature**: default shell child has TTY stdout

```
# probe via WS
WS attach -> python isatty(1) -> output tty=1
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "spawn-shell"
	req.Cwd = absTempDir(t)
	req.Name = "tty-probe"
	return nil
}
```