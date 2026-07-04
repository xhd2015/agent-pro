# Scenario

**Feature**: tty-watch surfaces CLI dispatch errors

```
# invalid subcommand
tty-watch <unknown> -> non-zero exit + error message
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if req.Bin == "" {
		t.Fatalf("errors setup: tty-watch binary not built")
	}
	return nil
}
```