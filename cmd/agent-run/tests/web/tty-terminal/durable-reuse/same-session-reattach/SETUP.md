# Scenario

**Feature**: same session keeps same backend tty after navigation

```
GET /terminal -> terminal id/session id
GET /sessions/<runner>/<id> -> user navigates away/back
GET /terminal again -> same registry-backed terminal
```

## Preconditions

- Parent setup wrote one session and one live registry.

## Steps

1. Run the first terminal status request.
2. Assertion simulates navigation by fetching session detail and terminal status again.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.HTTPPath = terminalStatusPath(req.Runner, req.SessionID)
	return nil
}
```
