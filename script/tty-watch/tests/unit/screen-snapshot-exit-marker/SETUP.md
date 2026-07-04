# Scenario

**Feature**: screen snapshot text must left-align `[Terminal exited]` when scrollback ends with LF+LF exit marker

```
scrollback: yes\n\n[Terminal exited]
  -> snapshot text lines must be ["yes", "[Terminal exited]"] at column 0
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "unit-screen-snapshot-exit-marker"
	return nil
}
```