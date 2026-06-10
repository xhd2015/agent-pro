package runner

import (
	"os"
	"path/filepath"
)

func ResolveRoot(dir string) (root string, ok bool) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", false
	}
	if _, err := os.Stat(absDir); err != nil {
		return "", false
	}

	var parents []string
	parents = append(parents, absDir)

	if !hasFile(absDir, "go.mod") {
		cur := absDir
		for {
			parent := filepath.Dir(cur)
			if parent == cur {
				break
			}
			if hasFile(parent, "go.mod") {
				parents = append(parents, parent)
				break
			}
			if filepath.Base(parent) == "testdata" {
				parents = append(parents, parent)
				break
			}
			parents = append(parents, parent)
			cur = parent
		}
	}

	for _, p := range parents {
		if hasFile(p, "DOCTEST.md") {
			return p, true
		}
	}

	for i := len(parents) - 1; i >= 0; i-- {
		if hasFile(parents[i], "SETUP.md") {
			return parents[i], true
		}
	}

	return "", false
}

func hasFile(dir, name string) bool {
	_, err := os.Stat(filepath.Join(dir, name))
	return err == nil
}
