# Scenario

**Feature**: GET /v1/models

```
GET /v1/models -> model list JSON
```

## Steps
1. No exchanges needed — `/v1/models` ignores the exchange config.
2. Send a GET request to `/v1/models`.
3. The server returns a static model list.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.ConfigJSON = `{
  "port": 8080,
  "exchanges": []
}`
    req.Endpoint = "/v1/models"
    req.Method = "GET"
    req.Requests = []string{""}
    return nil
}
```
