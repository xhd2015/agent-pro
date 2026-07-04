# Scenario

**Branch**: POST /v1/chat/completions

```
HTTP client -> POST /v1/chat/completions -> exchange matcher
```

## Preconditions
- This branch tests the `POST /v1/chat/completions` endpoint.
- The server is configured with exchange rules that define expected request→response mappings.

## Steps
1. Set the endpoint to `/v1/chat/completions` and method to `POST`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Endpoint = "/v1/chat/completions"
    req.Method = "POST"
    return nil
}
```
