# Scenario

**Feature**: GET `/` returns SPA HTML shell with `#root`

```
agent-run web -> GET / -> 200 text/html containing id="root"
```

## Preconditions

- Web server running with isolated home (no sessions required).

## Steps

1. Start web on free port.
2. Probe `GET /`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Scenario = "root-html-has-root"
	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}
	req.HTTPPath = "/"
	req.HTTPAuth = "none"
	return nil
}
```
