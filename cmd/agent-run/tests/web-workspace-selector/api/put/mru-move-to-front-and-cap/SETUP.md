# Scenario

**Feature**: MRU move-to-front and cap at 12 (A7)

```
# select 13 dirs then re-select an older one
PUT ws-0 … ws-12  (13 paths)
  -> recent length <= 12; head = last PUT
PUT ws-0 again
  -> head = ws-0; no duplicate of ws-0
```

## Preconditions

- 13 distinct absolute directories under TempDir.
- Cap constant `maxRecentWorkspaces` = 12.

## Steps

1. Create 13 dirs; start web.
2. PUT each in order; then re-PUT first path; GET status.

```go
import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Scenario = "mru-move-to-front-and-cap"
	n := maxRecentWorkspaces + 1 // 13
	paths := make([]string, 0, n)
	for i := 0; i < n; i++ {
		paths = append(paths, makeSelectDir(t, req, fmt.Sprintf("mru-%02d", i)))
	}
	req.SelectPath = paths[0] // re-selected path (move-to-front)
	// Persist full list for Assert (newline-separated absolute paths).
	listPath := filepath.Join(req.TempDir, "mru-paths.txt")
	var b []byte
	for _, p := range paths {
		b = append(b, p...)
		b = append(b, '\n')
	}
	if err := os.WriteFile(listPath, b, 0644); err != nil {
		return err
	}
	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}
	var steps []HTTPStep
	for i, p := range paths {
		steps = append(steps, HTTPStep{
			Name:   fmt.Sprintf("put-%d", i),
			Method: "PUT",
			Path:   "/api/agent-run/workspace",
			Body:   putWorkspaceBody(p),
		})
	}
	steps = append(steps, HTTPStep{
		Name:   "put-reorder",
		Method: "PUT",
		Path:   "/api/agent-run/workspace",
		Body:   putWorkspaceBody(paths[0]),
	})
	steps = append(steps, HTTPStep{
		Name:   "status",
		Method: "GET",
		Path:   "/api/agent-run/status",
	})
	req.HTTPSteps = steps
	return nil
}
```
