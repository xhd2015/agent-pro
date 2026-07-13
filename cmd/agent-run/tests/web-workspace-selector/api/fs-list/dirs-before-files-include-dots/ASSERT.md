## Expected

- HTTP **200**.
- Entries include `.git` and `src` as `type: "dir"`.
- Entries include `.env` as `type: "file"`.
- **All directories appear before any file** in `entries` order.
- Within the dir group and within the file group, names are sorted case-insensitively.

## Errors

- Pre-impl: `.git` / `.env` omitted (hidden skip); order may interleave by readdir.

```go
import (
	"encoding/json"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

type listEntry struct {
	Name string
	Type string // "dir" | "file"
}

func parseListEntries(t *testing.T, body string) []listEntry {
	t.Helper()
	var root map[string]any
	if err := json.Unmarshal([]byte(body), &root); err != nil {
		t.Fatalf("parse list body: %v body=%q", err, truncate(body, 300))
	}
	var raw []any
	if v, ok := root["entries"]; ok {
		raw, _ = v.([]any)
	} else if v, ok := root["items"]; ok {
		raw, _ = v.([]any)
	}
	if raw == nil {
		var arr []any
		if json.Unmarshal([]byte(body), &arr) == nil {
			raw = arr
		}
	}
	if len(raw) == 0 {
		t.Fatalf("expected non-empty entries, body=%q", truncate(body, 400))
	}
	out := make([]listEntry, 0, len(raw))
	for _, e := range raw {
		m, _ := e.(map[string]any)
		if m == nil {
			continue
		}
		typ := strings.ToLower(strings.TrimSpace(jsonStringField(m, "type")))
		name := jsonStringField(m, "name")
		if name == "" {
			name = filepath.Base(jsonStringField(m, "path"))
		}
		if typ == "directory" {
			typ = "dir"
		}
		if typ == "" {
			if b, ok := m["is_dir"].(bool); ok {
				if b {
					typ = "dir"
				} else {
					typ = "file"
				}
			}
		}
		if typ != "dir" && typ != "file" {
			// treat unknown as file for ordering checks
			typ = "file"
		}
		out = append(out, listEntry{Name: name, Type: typ})
	}
	return out
}

func Assert(t *testing.T, req *Request, resp *Response, err error) {
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
	entries := parseListEntries(t, list.Body)

	var sawGit, sawSrc, sawEnv bool
	var firstFileIdx = -1
	for i, e := range entries {
		switch e.Type {
		case "dir":
			if firstFileIdx >= 0 {
				t.Fatalf("dirs must appear before any file: entry[%d]=%q type=dir after file at %d; body=%q",
					i, e.Name, firstFileIdx, truncate(list.Body, 400))
			}
			if e.Name == ".git" {
				sawGit = true
			}
			if e.Name == "src" {
				sawSrc = true
			}
		case "file":
			if firstFileIdx < 0 {
				firstFileIdx = i
			}
			if e.Name == ".env" {
				sawEnv = true
			}
		}
	}
	if !sawGit {
		t.Fatalf("expected dir entry .git (dot dirs included), body=%q", truncate(list.Body, 400))
	}
	if !sawSrc {
		t.Fatalf("expected dir entry src, body=%q", truncate(list.Body, 400))
	}
	if !sawEnv {
		t.Fatalf("expected file entry .env (dot files included), body=%q", truncate(list.Body, 400))
	}
	if firstFileIdx < 0 {
		t.Fatalf("expected at least one file entry, body=%q", truncate(list.Body, 400))
	}

	// Case-insensitive name order within dir group and within file group.
	var dirNames, fileNames []string
	for _, e := range entries {
		if e.Type == "dir" {
			dirNames = append(dirNames, e.Name)
		} else {
			fileNames = append(fileNames, e.Name)
		}
	}
	assertCaseInsensitiveSorted(t, "dirs", dirNames)
	assertCaseInsensitiveSorted(t, "files", fileNames)
	_ = req.FixtureRoot
}

func assertCaseInsensitiveSorted(t *testing.T, label string, names []string) {
	t.Helper()
	if len(names) < 2 {
		return
	}
	want := append([]string(nil), names...)
	sort.SliceStable(want, func(i, j int) bool {
		return strings.ToLower(want[i]) < strings.ToLower(want[j])
	})
	for i := range names {
		if names[i] != want[i] {
			t.Fatalf("%s not case-insensitive sorted: got %v want %v", label, names, want)
		}
	}
}
```
