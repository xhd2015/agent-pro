# Scenario

**Feature**: flat layout supports per-file path overrides

```
# EventsPath/MetaPath/PIDPath can point inside or outside Dir
test -> subagent.Run(custom paths) -> artifacts at override locations
```

## Preconditions

- Flat `SessionLayout.Dir` exists.

## Steps

1. Descendant leaves set explicit path overrides on `SessionLayout`.

## Context

- Grouping for custom `EventsPath`, `MetaPath`, and related overrides.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if req.AgentRunner == "" {
		req.AgentRunner = "fake-codex"
	}
	return nil
}```
