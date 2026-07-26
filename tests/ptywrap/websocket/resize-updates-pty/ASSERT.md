## Expected

- After resize message, `stty size` output reflects requested dimensions.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.PTYCols != req.ResizeCols || resp.PTYRows != req.ResizeRows {
		t.Fatalf("expected PTY size %dx%d, parsed %dx%d from %q",
			req.ResizeCols, req.ResizeRows, resp.PTYCols, resp.PTYRows, resp.WSOutput)
	}
}
```