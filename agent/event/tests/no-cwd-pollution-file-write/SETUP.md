# Scenario

**Feature**: The `fakeagent` package generates file write events during probe simulation

## Preconditions
- The `fakeagent` package generates file write events during probe simulation.
- The prompt contains patterns that trigger `KindFileWrite` (e.g., "create output.txt").
- File writes land under the package temp `probeWriteDir`, not process cwd.
- Probe grep/bash use `Generator.WorkDir` (no `os.Chdir`).

## Steps
1. Create a fresh temporary WorkDir for probe relative ops.
2. Call `fakeagent.GenerateSession` with seed 100 and a prompt that triggers file writes.
3. List process cwd for new entries (must remain empty of probe pollution).

```go
import (
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	fakeagent "github.com/xhd2015/agent-pro/pkgs/fake-agent"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	work := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getwd: %w", err)
	}
	before, err := listDirNames(cwd)
	if err != nil {
		return err
	}

	g := fakeagent.NewGenerator(100)
	g.WorkDir = work
	g.GenerateSession("create output.txt and write result with content")

	after, err := listDirNames(cwd)
	if err != nil {
		return err
	}

	beforeSet := map[string]struct{}{}
	for _, n := range before {
		beforeSet[n] = struct{}{}
	}
	var polluted []string
	for _, n := range after {
		if _, ok := beforeSet[n]; !ok {
			polluted = append(polluted, n)
		}
	}
	sort.Strings(polluted)

	var sb strings.Builder
	for _, f := range polluted {
		fmt.Fprintf(&sb, "FILE: %s\n", f)
	}
	fmt.Fprint(&sb, "DONE")
	req.Output = sb.String()
	return nil
}

func listDirNames(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("readdir %s: %w", dir, err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names, nil
}
```
