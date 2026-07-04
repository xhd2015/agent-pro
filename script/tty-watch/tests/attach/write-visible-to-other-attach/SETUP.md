# Scenario

**Feature**: attach A write is visible to attach B (output broadcast)

```
# two attachers; A writes; B observes A's marker without writing it
harness -> detached cat -> attach A + attach B -> A writes -> B output has marker
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "attach-visible-to-other"
	req.AttachInput = "ATTACH_WRITE_MARKER_A"
	return nil
}
```