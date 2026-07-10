# Scenario

**Feature**: live send with explicit --config and required message

```
slack-send --config repo/slack-config.json "Hello from doctest" -> real API -> OK
```

## Steps

1. Inherit integration grouping setup (repo config path).
2. Explicit message positional; no auto-discovery.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--config", req.ConfigPath, "Hello from doctest"}
	return nil
}
```