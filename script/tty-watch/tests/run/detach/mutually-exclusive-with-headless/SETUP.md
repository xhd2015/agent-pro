# Scenario

**Feature**: `--detach` and `--headless` cannot be used together

```
# reject before session start; no registry entry
harness -> tty-watch run --detach --headless true -> exit 1
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "run-detach-mutually-exclusive-with-headless"
	req.RunCommand = []string{"true"}
	return nil
}
```