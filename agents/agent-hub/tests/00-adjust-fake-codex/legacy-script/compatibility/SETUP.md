## Preconditions
- The legacy script contains one message event.

## Steps
1. Run fake Codex with `--script`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    writeLegacyScript(t, req, `{"events":[{"type":"item.completed","item":{"id":"m1","type":"message","text":"legacy still works","status":"completed"}}]}`)
    req.Args = []string{"exec", "--json", "--script", req.LegacyScriptPath, "hello"}
    return nil
}
```

