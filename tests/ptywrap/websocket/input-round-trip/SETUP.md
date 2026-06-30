# Scenario

**Feature**: WS keystrokes reach PTY stdin and appear in output

```
# cat echo
WS binary input -> cat -> same bytes in output stream
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "ws-input"
	req.WSInput = "roundtrip-marker\n"
	return nil
}
```