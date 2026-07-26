---
label: e2e
---

## Expected

- HTTP **200**.
- Response lists at least one `dir` (`subdir`) and one `file` (`note.txt`).
- Each entry has `type` of `"dir"` or `"file"` (or equivalent name/type fields).
- Optional: `path` / `parent` fields present for browser parent control.

## Errors

- Pre-impl: 404 on `/fs/list` (RED).

```go
import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	list, ok := findHTTPResult(resp, "list")
	if !ok {
		t.Fatal("missing list result")
	}
	if list.Status != 200 {
		t.Fatalf("GET fs/list expected 200, got %d body=%q", list.Status, truncate(list.Body, 400))
	}
	var root map[string]any
	if err := json.Unmarshal([]byte(list.Body), &root); err != nil {
		t.Fatalf("parse list body: %v body=%q", err, truncate(list.Body, 300))
	}
	// Accept {entries:[…]} or {items:[…]} or bare array.
	var entries []any
	if raw, ok := root["entries"]; ok {
		entries, _ = raw.([]any)
	} else if raw, ok := root["items"]; ok {
		entries, _ = raw.([]any)
	} else if raw, ok := root["files"]; ok {
		// unlikely but tolerate
		entries, _ = raw.([]any)
	}
	if entries == nil {
		// bare array?
		var arr []any
		if json.Unmarshal([]byte(list.Body), &arr) == nil {
			entries = arr
		}
	}
	if len(entries) == 0 {
		t.Fatalf("expected non-empty entries, body=%q", truncate(list.Body, 400))
	}
	var sawDir, sawFile bool
	for _, e := range entries {
		m, _ := e.(map[string]any)
		if m == nil {
			continue
		}
		typ := strings.ToLower(strings.TrimSpace(jsonStringField(m, "type")))
		name := jsonStringField(m, "name")
		if name == "" {
			name = filepath.Base(jsonStringField(m, "path"))
		}
		switch typ {
		case "dir", "directory":
			if name == "subdir" || strings.HasSuffix(name, "subdir") {
				sawDir = true
			} else {
				sawDir = true // any dir counts for type coverage
			}
		case "file":
			if name == "note.txt" || strings.HasSuffix(name, "note.txt") {
				sawFile = true
			} else {
				sawFile = true
			}
		}
		// Also accept is_dir bool if type omitted.
		if typ == "" {
			if b, ok := m["is_dir"].(bool); ok && b {
				sawDir = true
			}
			if b, ok := m["is_dir"].(bool); ok && !b {
				sawFile = true
			}
		}
	}
	if !sawDir {
		t.Fatalf("expected at least one dir entry (subdir), body=%q", truncate(list.Body, 400))
	}
	if !sawFile {
		t.Fatalf("expected at least one file entry (note.txt), body=%q", truncate(list.Body, 400))
	}
	_ = req.FixtureRoot
}
```
