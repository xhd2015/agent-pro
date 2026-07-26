## Expected
- Every `*.in` with a sibling `.want` produces that exact formatted message on stdout (plus trailing newline).
- Every `*.in` with a sibling `.want_err` hard-fails (resp.Err != nil).
- No fixture is skipped: corpus size must match the on-disk anti_patterns set.

## Side Effects
- Does not require `--commit` (pure message path). Named leaves cover git subject / HEAD.

```go
import (
	"strconv"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run harness error: %v", err)
	}
	// Discard placeholder Run; re-execute every fixture against the staged repo.
	names := ListAntiPatternNames(t)
	var failures []string
	for _, name := range names {
		req.Commit = false
		r := RunAntiPatternFixture(t, req, name)
		if HasAntiPatternWantErr(name) {
			if r.Err == nil {
				failures = append(failures, name+": expected reject, got success stdout="+strconv.Quote(r.Stdout))
			}
			continue
		}
		want := ReadAntiPatternWant(t, name)
		if r.Err != nil {
			failures = append(failures, name+": unexpected err: "+r.Err.Error())
			continue
		}
		wantOut := want + "\n"
		if r.Stdout != wantOut {
			failures = append(failures, name+": stdout mismatch got="+strconv.Quote(r.Stdout)+" want="+strconv.Quote(wantOut))
		}
	}
	if len(failures) > 0 {
		t.Fatalf("anti_patterns corpus failures (%d/%d):\n%s", len(failures), len(names), strings.Join(failures, "\n"))
	}
}
```
