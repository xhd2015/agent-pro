# Scenario

**Feature**: config load or validation fails before send

```
slack-send -> findConfigPath -> loadConfig / botToken check -> stderr error -> exit 1
```

## Preconditions

- Isolated `WorkDir` with minimal `go.mod` stops config search at temp dir.

## Steps

1. Grouping enables `WriteGoMod` and clears `UseRepoConfig`.
2. Leaf chooses missing file vs empty-token fixture.

## Context

- No `Sending to...` line on these paths.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.WriteGoMod = true
	req.UseRepoConfig = false
	req.SlackAPIURL = ""
	return nil
}
```