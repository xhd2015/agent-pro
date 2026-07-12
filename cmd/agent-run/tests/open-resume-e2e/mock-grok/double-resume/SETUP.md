# Scenario

**Feature**: exit → resume → exit → resume again (double reclaim)

```
open Paris → /exit → resume --open hello → /exit → resume --open hello
  -> neither resume reports already in use; both succeed
```

## Steps

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Scenario = "double-resume"
	req.SessionID = "e2e-double-resume"
	req.GrokSessionUUID = "b2222222-2222-4222-8222-222222222204"
	req.OpenPrompt = defaultOpenPrompt
	req.WantParis = defaultWantParis
	req.FollowupPrompt = "hello"
	req.HelloMarker = defaultHelloMarker
	req.ResumeTwice = true
	req.ResumeOpen = true
	configureMockGrokEnv(t, req)
	return nil
}
```
