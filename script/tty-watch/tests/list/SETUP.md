# Scenario

**Feature**: `tty-watch list` scans registry and prints live sessions

```
# registry scan + TCP probe
tty-watch list -> registry dir -> probe listen_addr -> print id command uptime
```

## Preconditions

- `list-fields` leaf starts a detached session first via harness helper.
- `list-empty` uses fresh isolated home with no registry entries.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if req.Bin == "" {
		t.Fatalf("list setup: tty-watch binary not built")
	}
	return nil
}
```