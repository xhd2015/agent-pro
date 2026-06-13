## Preconditions
- The `fakeagent` package generates file write events during probe simulation.
- The prompt contains patterns that trigger `KindFileWrite` (e.g., "create output.txt").

## Steps
1. Create a fresh temporary directory and chdir into it.
2. Call `fakeagent.GenerateSession` with seed 100 and a prompt that triggers file writes.
3. List the temporary directory contents to detect pollution.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.MainGo = `package main

import (
	"fmt"
	"os"
	"sort"

	fakeagent "github.com/xhd2015/agent-pro/pkgs/fake-agent"
)

func main() {
	dir, err := os.MkdirTemp("", "fake-agent-probe-*")
	if err != nil {
		fmt.Println("ERROR:", err)
		return
	}
	defer os.RemoveAll(dir)

	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		fmt.Println("ERROR:", err)
		return
	}
	defer os.Chdir(origDir)

	g := fakeagent.NewGenerator(100)
	g.GenerateSession("create output.txt and write result with content")

	entries, err := os.ReadDir(".")
	if err != nil {
		fmt.Println("ERROR:", err)
		return
	}

	var files []string
	for _, e := range entries {
		files = append(files, e.Name())
	}
	sort.Strings(files)
	for _, f := range files {
		fmt.Println("FILE:", f)
	}
	fmt.Println("DONE")
}
`
	return nil
}
```
