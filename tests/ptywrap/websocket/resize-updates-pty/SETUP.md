# Scenario

**Feature**: resize JSON updates PTY window dimensions

```
# stty probe
WS resize JSON -> PTY -> stty size reports new cols/rows
```

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "ws-resize"
	req.ResizeCols = 100
	req.ResizeRows = 40
	return nil
}
```