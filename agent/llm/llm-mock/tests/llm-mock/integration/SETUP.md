## Preconditions
- This branch tests integration scenarios where agents (opencode, pi) use the mock server as a drop-in LLM backend.
- Each leaf verifies both client-side output and server-side request recording via `GET /admin/requests`.

## Steps
1. Ensure the endpoint defaults to `/v1/chat/completions` with `POST` method.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Endpoint = "/v1/chat/completions"
    req.Method = "POST"
    return nil
}
```
