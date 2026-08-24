## Expected

- One hosted row.
- TITLE and WORKSPACE come from summary.json on disk.

```go
import "path/filepath"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	assertContains(t, resp.Stdout, "1 sessions", "footer")
	assertContains(t, resp.Stdout, fixtureListLiveSID, "sid")
	assertContains(t, resp.Stdout, fixtureListLiveDiskTitle, "title from disk")
	wantCwd, errAbs := filepath.Abs(fixtureListLiveDiskCWD)
	if errAbs != nil {
		t.Fatalf("abs fixture cwd: %v", errAbs)
	}
	assertContains(t, resp.Stdout, wantCwd, "workspace from disk")
}
```
