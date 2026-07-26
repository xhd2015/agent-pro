# Scenario

**Feature**: public package surface for agentrunapi

```
import agentrunapi
  -> ModeRun | ModeSend | ModeResume
  -> Classify(...)
  -> AutoSendOrResume(...)
  -> Opts | ProbeReport types
  -> LifecycleProbe | EmptyProbe
```

## Preconditions

- Package path `github.com/xhd2015/agent-pro/pkgs/agentrunapi` must compile.
- Symbols listed in root DOCTEST planned API must exist.

## Steps

1. Set harness mode `api_surface`.
2. `Run` touches constants/types, probes, Classify + AutoSendOrResume.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "api_surface"
	return nil
}
```
