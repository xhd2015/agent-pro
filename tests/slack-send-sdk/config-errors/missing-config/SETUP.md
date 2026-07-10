# Scenario

**Feature**: missing slack-config.json next to isolated go.mod

```
loadConfig -> open failure -> stderr failed to load config -> exit 1
```

## Steps

1. Write go.mod only; do not write slack-config.json.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ConfigFixture = ""
	req.ConfigInline = ""
	req.Args = nil
	return nil
}
```