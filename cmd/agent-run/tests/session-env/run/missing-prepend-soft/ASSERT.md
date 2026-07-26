---
label: e2e
---

## Expected

- Exit code 0 (soft allow; missing dir is not a hard error).
- Env probe `PATH` starts with the absolute missing path.
- Meta `prepend_paths` contains that abs path.

## Exit Code

0

```go
import (
	"os"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)

	if _, statErr := os.Stat(req.PrependPathDir); !os.IsNotExist(statErr) {
		t.Fatalf("precondition: prepend dir should be missing; stat=%v", statErr)
	}

	probe := readEnvProbe(t, req.EnvProbePath)
	assertProbePATHPrefixed(t, probe, req.PrependPathDir)

	meta := readMetaJSON(t, req.Home, req.SessionID)
	assertMetaStringSliceEquals(t, meta, "prepend_paths", []string{req.PrependPathDir})
}
```
