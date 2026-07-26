# Scenario

**Feature**: Attach rejects non-interactive stdin/stdout

```
# piped stdin
Attach(pipe, pipe) -> error mentions interactive terminal
```

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "attach-requires-tty"
	req.UsePipeStdin = true
	req.UsePipeStdout = true
	return nil
}
```