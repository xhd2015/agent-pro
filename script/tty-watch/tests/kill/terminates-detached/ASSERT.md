## Expected

- Exit code 0.
- Registry file removed for killed session.
- Session no longer TCP-reachable.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("kill exit %d combined %q", resp.ExitCode, resp.Combined)
	}
	if resp.RegistryExists {
		t.Fatalf("registry should be removed after kill for %s", resp.SessionID)
	}
	if resp.SessionRunning {
		t.Fatalf("session %s should not be reachable after kill", resp.SessionID)
	}
}
```