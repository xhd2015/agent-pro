# Scenario

**Feature**: public package surface for agentrunapi

```
import agentrunapi
  -> ModeRun | ModeSend | ModeResume
  -> Classify(...)
  -> AutoSendOrResume(...)
  -> Opts | ProbeReport types
```

## Preconditions

- Package path `github.com/xhd2015/agent-pro/pkgs/agentrunapi` must compile.
- Symbols listed in root DOCTEST planned API must exist.

## Steps

1. Set harness mode `api_surface`.
2. `Run` touches constants/types and calls Classify + AutoSendOrResume.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "api_surface"
	return nil
}
```
