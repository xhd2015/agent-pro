# Scenario

**Feature**: missing config file returns empty defaults

```
fresh AGENT_RUN_HOME (no config.json) -> Config() -> zero-value Config
```

## Preconditions

- Home directory is empty; no prior `SaveConfig` call.
- `config.json` does not exist on disk.

## Steps

1. Set `req.Action = "load_missing"`.
2. Call `Config()` on a fresh store.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Action = "load_missing"
	return nil
}
```