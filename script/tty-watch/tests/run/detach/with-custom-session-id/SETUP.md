# Scenario

**Feature**: detach accepts `--session-id` before command argv

```
# custom id printed on stdout and written to registry
harness -> tty-watch run --detach --session-id my-job sleep 120 -> exit 0
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "run-detach-with-custom-session-id"
	req.CustomSessionID = "my-job"
	req.RunCommand = []string{"sleep", "120"}
	return nil
}
```