package validate

import (
	"fmt"
	"os"
	"path/filepath"
)

func Run() error {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: validate_test_case_tree <dir>")
		os.Exit(1)
	}

	dir := os.Args[1]
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory not found: %s", dir)
		}
		return fmt.Errorf("cannot stat directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", dir)
	}

	var errs []string
	errs = validateRoot(dir, errs)
	errs = validateSubdirs(dir, errs)

	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, e)
		}
		os.Exit(1)
	}

	return nil
}

func validateRoot(dir string, errs []string) []string {
	readmePath := filepath.Join(dir, "README.md")
	info, err := os.Stat(readmePath)
	if err != nil || info.IsDir() {
		errs = append(errs, fmt.Sprintf("%s: root must contain README.md", dir))
	}
	return errs
}

func validateSubdirs(root string, errs []string) []string {
	filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			errs = append(errs, fmt.Sprintf("%s: error accessing: %v", path, walkErr))
			return nil
		}
		if !d.IsDir() {
			return nil
		}

		isRoot := path == root

		entries, err := os.ReadDir(path)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: cannot read directory: %v", path, err))
			return nil
		}

		if !isRoot && len(entries) == 0 {
			errs = append(errs, fmt.Sprintf("%s: empty directory (must have at least one child)", path))
			return nil
		}

		hasAssert := false
		hasSetup := false
		for _, e := range entries {
			if !e.IsDir() {
				switch e.Name() {
				case "ASSERT.md":
					hasAssert = true
				case "SETUP.md":
					hasSetup = true
				}
			}
		}

		if hasAssert && !hasSetup {
			errs = append(errs, fmt.Sprintf("%s: ASSERT.md found but SETUP.md missing", path))
		}

		return nil
	})

	return errs
}
