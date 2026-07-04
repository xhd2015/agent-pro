# Scenario

**Feature**: two attach clients both write; input is queued, not dropped

```
# two attach_mode=attach clients; distinct markers; both reach PTY
harness -> detached cat -> attach A writes MARKER_A -> attach B writes MARKER_B -> both present
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "attach-second-writes"
	req.AttachInput = "ATTACH_WRITE_MARKER_A"
	req.AttachInputB = "ATTACH_WRITE_MARKER_B"
	return nil
}
```