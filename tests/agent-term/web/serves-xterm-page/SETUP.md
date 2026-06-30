# Scenario

**Feature**: web command serves xterm HTML on ephemeral port

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "web-xterm"
	req.WebSessionName = "web-target"
	req.StartDaemon = true
	return nil
}
```