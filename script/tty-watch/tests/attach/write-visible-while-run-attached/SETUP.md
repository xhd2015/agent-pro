# Scenario

**Feature**: attach writes while `run` holds screen writer; output visible to run and watch

```
# run sleep attached (screen writer); separate attach writes; run stdout + watch see marker
harness -> run sleep (writer) -> watch + attach -> attach writes -> run + watch output
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "attach-visible-while-run"
	req.RunCommand = []string{"sleep", "300"}
	req.AttachInput = "ATTACH_WRITE_MARKER_A"
	return nil
}
```