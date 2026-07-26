# Scenario

**Feature**: full server-client cycle of subscribe, send prompt, and read SSE events

## Preconditions
- Server is running.
- A workspace is created.
- SSE subscription is established.
- A prompt is sent.

## Steps
1. Set `ServerOperation` to `"send-and-receive"`.
2. Root `Run` will:
   a. Ensure server is running.
   b. Create a workspace.
   c. Subscribe to SSE events.
   d. Send the prompt "hello".
   e. Read SSE events until `run_complete`.
   f. Return the collected events as JSON.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ServerOperation = "send-and-receive"
	return nil
}
```
