# Scenario

**Feature**: The `fakeagent` package generates file write events during probe simulation

## Preconditions
- The `fakeagent` package generates file write events during probe simulation.
- The prompt contains patterns that trigger `KindFileWrite` (e.g., "create output.txt").

## Steps
1. Create a fresh temporary directory and chdir into it.
2. Call `fakeagent.GenerateSession` with seed 100 and a prompt that triggers file writes.
3. List the temporary directory contents to detect pollution.

```go
import (
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	fakeagent "github.com/xhd2015/agent-pro/pkgs/fake-agent"
)

func Setup(t *testing.T, req *Request) error {
	dir, err := os.MkdirTemp("", "fake-agent-probe-*")
	if err != nil {
		return fmt.Errorf("mkdirtemp: %w", err)
	}
	defer os.RemoveAll(dir)

	origDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getwd: %w", err)
	}
	if err := os.Chdir(dir); err != nil {
		return fmt.Errorf("chdir: %w", err)
	}
	defer os.Chdir(origDir)

	g := fakeagent.NewGenerator(100)
	g.GenerateSession("create output.txt and write result with content")

	entries, err := os.ReadDir(".")
	if err != nil {
		return fmt.Errorf("readdir: %w", err)
	}

	var files []string
	for _, e := range entries {
		files = append(files, e.Name())
	}
	sort.Strings(files)

	var sb strings.Builder
	for _, f := range files {
		fmt.Fprintf(&sb, "FILE: %s\n", f)
	}
	fmt.Fprint(&sb, "DONE")
	req.Output = sb.String()
	return nil
}
```
