## Expected

- Paris then exited true.
- `resume.ready: yes` (or JSON ready true) after exit.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if !resp.HasParis {
		t.Fatalf("want Paris")
	}
	if !resp.ExitedTrue {
		t.Fatalf("want exited true:\n%s\njson:%s", resp.StatusAfterExit.Stdout, resp.StatusJSONAfterExit)
	}
	if !resp.ResumeReady {
		t.Fatalf("want resume.ready yes after exit:\n%s\njson:%s",
			resp.StatusAfterExit.Stdout, resp.StatusJSONAfterExit)
	}
}
```
