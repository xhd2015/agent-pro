## Expected

- Exit code 0.
- Child `PATH` starts with stored orig dir, then the resume-added dir (order: stored then new).
- Meta `prepend_paths` grows to `[orig, more]` (no dedup; both present in order).

## Exit Code

0

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)

	probe := readEnvProbe(t, req.EnvProbePath)
	assertProbePATHPrefixed(t, probe, req.PrependPathDir, req.PrependPathMore)

	meta := readMetaJSON(t, req.Home, req.SessionID)
	assertMetaStringSliceEquals(t, meta, "prepend_paths", []string{
		req.PrependPathDir,
		req.PrependPathMore,
	})
}
```
