## Expected

```go
import "strings"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertOK(t, resp)
	want := "Opened new window; forking grok session " + fixtureTabSessionID + "\n"
	if resp.Stdout != want {
		t.Fatalf("stdout=%q want %q", resp.Stdout, want)
	}
	if len(resp.Opened) != 1 {
		t.Fatalf("Opened=%v, want 1", resp.Opened)
	}
	if !strings.Contains(resp.Opened[0], fixtureTabSessionID) {
		t.Fatalf("follow-up missing session id: %q", resp.Opened[0])
	}
	if !strings.Contains(resp.Opened[0], "--fork-session") {
		t.Fatalf("follow-up missing --fork-session: %q", resp.Opened[0])
	}
	if resp.RunForegroundN != 0 {
		t.Fatal("-n must not run foreground")
	}
}
```
