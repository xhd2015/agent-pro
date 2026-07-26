---
label: e2e
---

## Expected

- Every PUT returns **200**.
- After 13 PUTs + reorder: `recent_workspaces` length **≤ 12**.
- Head of recent is re-selected path (`req.SelectPath` = first dir).
- Re-selected path appears **once** (dedupe move-to-front).
- After overflow, oldest unique path that fell off is absent when length == 12.

## Errors

- Pre-impl: PUT missing (RED).

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	// All PUT steps must succeed.
	for _, r := range resp.HTTPResults {
		if strings.HasPrefix(r.Name, "put") && r.Status != 200 {
			t.Fatalf("%s expected 200, got %d body=%q", r.Name, r.Status, truncate(r.Body, 200))
		}
	}
	st, ok := findHTTPResult(resp, "status")
	if !ok {
		t.Fatal("missing status result")
	}
	if st.Status != 200 {
		t.Fatalf("GET status expected 200, got %d body=%q", st.Status, truncate(st.Body, 300))
	}
	m := parseJSONMap(t, st.Body)
	recent := jsonStringSlice(m, "recent_workspaces")
	if len(recent) == 0 {
		// Fallback to config.json
		cfg := readHomeConfigMap(t, req.Home)
		if cfg != nil {
			recent = jsonStringSlice(cfg, "recent_workspaces")
		}
	}
	if len(recent) == 0 {
		t.Fatal("expected recent_workspaces after MRU PUTs")
	}
	if len(recent) > maxRecentWorkspaces {
		t.Fatalf("recent_workspaces length %d exceeds cap %d: %v", len(recent), maxRecentWorkspaces, recent)
	}
	if !pathsEqual(recent[0], req.SelectPath) {
		t.Fatalf("MRU head after reorder: got %q want %q", recent[0], req.SelectPath)
	}
	// Dedupe: SelectPath appears once.
	count := 0
	for _, p := range recent {
		if pathsEqual(p, req.SelectPath) {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("SelectPath should appear once in MRU, count=%d list=%v", count, recent)
	}
	// After 13 unique inserts + re-select of #0: cap 12 means one path dropped.
	// Last unique insert before reorder was path #12; head becomes #0.
	// Path that should have fallen off: the oldest among first 13 after insert of #12
	// is path #0... wait, after 13 puts head is #12 and #0 is oldest; then reorder
	// moves #0 to front. Still 12 entries. Path #1 is oldest after reorder of #0.
	// Just assert length == 12 when we inserted 13 unique.
	if len(recent) != maxRecentWorkspaces {
		t.Fatalf("after 13 unique + reorder, want len=%d got %d list=%v",
			maxRecentWorkspaces, len(recent), recent)
	}
	// Load original path list to ensure a dropped middle path is not all present.
	data, err := os.ReadFile(filepath.Join(req.TempDir, "mru-paths.txt"))
	if err != nil {
		t.Fatalf("read mru-paths: %v", err)
	}
	var all []string
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			all = append(all, line)
		}
	}
	if len(all) != maxRecentWorkspaces+1 {
		t.Fatalf("fixture path count: got %d want %d", len(all), maxRecentWorkspaces+1)
	}
	// At least one of the 13 original paths is absent from recent (cap).
	present := 0
	for _, p := range all {
		for _, r := range recent {
			if pathsEqual(p, r) {
				present++
				break
			}
		}
	}
	if present != maxRecentWorkspaces {
		t.Fatalf("expected exactly %d of %d unique paths in MRU, present=%d recent=%v",
			maxRecentWorkspaces, len(all), present, recent)
	}
}
```
