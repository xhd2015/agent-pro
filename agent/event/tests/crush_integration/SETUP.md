# Scenario

**Feature**: A `crush` binary is available on `PATH` (or `CrushPath` is configured)

## Preconditions
- A `crush` binary is available on `PATH` (or `CrushPath` is configured).
- The test will start `crush server` on a free port and manage the lifecycle.

## Steps
1. Set `Target` to `"crush_server"` so the root `Run` function enters the server-mode dispatch.
2. Apply the default model name when none is provided.

## Context
- The root `Run` function (defined in `../SETUP.md`) is responsible for the full
  server lifecycle: start, health check, workspace/session creation, SSE
  subscription, prompt delivery, event collection, `FromCrush` mapping, and
  teardown.
- `Request.CrushPath`, `Request.HostPort`, `Request.Prompt`, and
  `Request.ModelName` control the server configuration.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Target = "crush_server"
	if req.ModelName == "" {
		req.ModelName = "deepseek-v4-pro"
	}
	return nil
}
```
