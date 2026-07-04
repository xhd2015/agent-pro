# Scenario

**Feature**: port fallback when occupied

```
# ephemeral port N blocked before server start
block listener on :N -> server on :N+1
```

## Steps
1. Acquire a free ephemeral port N via `net.Listen("tcp", "127.0.0.1:0")`.
2. Re-bind `:%d` on all interfaces so the server's `:%d` bind is blocked (loopback-only block is insufficient).
3. Write config with `"port": N`.
4. Hold the listener open so N stays occupied during server startup.
5. The server should fall back to port N+1.
6. Verify the server is listening on port N+1.

```go
import (
    "fmt"
    "net"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    tmp, err := net.Listen("tcp", "127.0.0.1:0")
    if err != nil {
        return fmt.Errorf("acquire ephemeral port: %w", err)
    }
    port := tmp.Addr().(*net.TCPAddr).Port
    tmp.Close()

    ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
    if err != nil {
        return fmt.Errorf("block port %d: %w", port, err)
    }

    req.BlockListener = ln
    t.Cleanup(func() { ln.Close() })

    req.ConfigJSON = fmt.Sprintf(`{
  "port": %d,
  "exchanges": [
    {
      "request": {
        "role": "user",
        "content": "fallback test",
        "index": -1
      },
      "response": {
        "content": "port fallback works",
        "finish_reason": "stop"
      }
    }
  ]
}`, port)
    req.Port = port
    req.ExpectedFallbackPort = port + 1
    req.Requests = []string{
        `{"model":"gpt-4","messages":[{"role":"user","content":"fallback test"}]}`,
    }
    return nil
}
```