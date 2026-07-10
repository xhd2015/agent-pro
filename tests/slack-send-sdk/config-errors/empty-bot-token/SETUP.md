# Scenario

**Feature**: config loads but botToken is empty

```
loadConfig ok -> botToken empty -> stderr botToken is empty -> exit 1
```

## Steps

1. Use `empty-token-config.json` fixture.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ConfigFixture = "empty-token-config.json"
	req.Args = nil
	return nil
}
```