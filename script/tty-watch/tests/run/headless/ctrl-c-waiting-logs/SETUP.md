# Scenario

**Feature**: headless Ctrl-C grace window logs once on stderr then force-kills

```
# sleep ignores INT; SIGINT headless parent -> stderr status after 1s -> kill at 10s exit 1
harness SIGINT -> tty-watch run --headless sleep 300 -> waiting for program to exit...
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "run-headless-ctrl-c-waiting-logs"
	req.RunCommand = []string{"sleep", "300"}
	return nil
}
```