## Expected

- Raw `Tools` still includes both tools with correct Counts (numeric asserts preserved).
- Output contains section `Tool handler time`.
- Header row mentions `NAME` and `N` (table columns).
- In output text, `read_file` appears **before** `bash` (N=3 before N=1).

## Errors

- None.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if resp.Stats == nil {
		t.Fatal("stats is nil")
	}
	st := resp.Stats

	var nRead, nBash int
	for _, tool := range st.Tools {
		switch tool.Name {
		case "read_file":
			nRead = tool.Count
		case "bash":
			nBash = tool.Count
		}
	}
	if nRead != 3 || nBash != 1 {
		t.Fatalf("tool counts read_file=%d bash=%d, want 3 and 1", nRead, nBash)
	}

	out := resp.Output
	assertContains(t, out, "Tool handler time")
	// Header columns (table form, not free-form "n=" lines alone).
	if !strings.Contains(out, "NAME") {
		t.Fatalf("tool table missing NAME header column:\n%s", out)
	}
	// Standalone N column header (word boundary-ish): require "N" near SUCCESS/ERROR or after NAME.
	if !strings.Contains(out, "SUCCESS") && !strings.Contains(out, " N ") && !strings.Contains(out, "\tN") {
		// Accept header line that lists N as a column among NAME ... SUCCESS ...
		headerOK := false
		for _, line := range strings.Split(out, "\n") {
			u := strings.ToUpper(line)
			if strings.Contains(u, "NAME") && strings.Contains(u, "N") &&
				(strings.Contains(u, "SUCCESS") || strings.Contains(u, "ERROR") || strings.Contains(u, "AVG")) {
				headerOK = true
				break
			}
		}
		if !headerOK {
			t.Fatalf("tool table missing NAME/N (and SUCCESS|ERROR|AVG) header:\n%s", out)
		}
	}

	idxRead := strings.Index(out, "read_file")
	idxBash := strings.Index(out, "bash")
	if idxRead < 0 || idxBash < 0 {
		t.Fatalf("output missing tool names read_file/bash:\n%s", out)
	}
	if idxRead > idxBash {
		t.Fatalf("tools not sorted by N desc: bash (n=1) appears before read_file (n=3):\n%s", out)
	}
}
```
