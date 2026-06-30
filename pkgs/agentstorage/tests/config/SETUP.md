# Scenario

**Feature**: persistent runner configuration at `config.json`

```
missing config.json -> Config() returns zero-value defaults
SaveConfig(cfg) -> config.json -> Config() roundtrip equality
```

## Preconditions

- Fresh home has no `config.json` until first `SaveConfig`.
- `Config` fields: `DefaultAgentRunner`, `DefaultModel`, `LastSession`.

## Steps

1. Set `req.Operation = "config"`.
2. Leaf Setup sets `req.Action` to `load_missing` or `save_reload` and optional `req.Config`.
3. `Run` calls `Config` and/or `SaveConfig` on the store.
4. Leaf `Assert` checks defaults or roundtrip values.

## Context

- `Response.Config` holds the loaded configuration.
- `Response.FilesWritten` includes `config.json` after save.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Operation = "config"
	return nil
}
```