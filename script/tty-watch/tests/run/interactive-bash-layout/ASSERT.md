## Expected

- PTY capture contains `LAYOUT_OK` after the harness sends `echo LAYOUT_OK`.
- Every line containing `bash:` has `bash:` at column 0–7 (no 8+ leading spaces).
- Output must **not** match the smeared anti-pattern ` {8,}bash:`.

```go
import (
	"regexp"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	combined := resp.Combined
	if combined == "" {
		combined = resp.Stdout
	}
	if !strings.Contains(combined, "LAYOUT_OK") {
		t.Fatalf("expected LAYOUT_OK in output, got %q", combined)
	}
	smeared := regexp.MustCompile(` {8,}bash:`)
	if smeared.MatchString(combined) {
		t.Fatalf("smeared bash error line (8+ spaces before bash:), got %q", combined)
	}
	for _, line := range strings.Split(combined, "\n") {
		if idx := strings.Index(line, "bash:"); idx >= 0 && idx >= 8 {
			t.Fatalf("bash: error padded right (column %d), line %q", idx, line)
		}
	}
}
```