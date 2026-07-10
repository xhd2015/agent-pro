# Scenario

**Feature**: missing app token for auth status --app

```
Caller -> slack-msg auth status --app -> app token required
```

## Steps

1. Args `["auth", "status", "--app"]` only.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"auth", "status", "--app"}
	return nil
}
```
