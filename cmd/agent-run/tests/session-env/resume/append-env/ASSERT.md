## Expected

- Exit code 0.
- Child env has `FOO=bar` and `NEW=1`.
- Meta `env` contains both entries (order: stored then appended).

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
	assertProbeHasKV(t, probe, "FOO", "bar")
	assertProbeHasKV(t, probe, "NEW", "1")

	meta := readMetaJSON(t, req.Home, req.SessionID)
	assertMetaStringSliceEquals(t, meta, "env", []string{"FOO=bar", "NEW=1"})
}
```
