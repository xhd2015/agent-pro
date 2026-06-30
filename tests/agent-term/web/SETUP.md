# Scenario

**Feature**: web serves ephemeral xterm.js page

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.StartDaemon = true
	return nil
}
```