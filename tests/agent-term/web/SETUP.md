# Scenario

**Feature**: web serves ephemeral xterm.js page

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.StartDaemon = true
	return nil
}
```