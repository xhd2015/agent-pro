package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/xhd2015/agent-pro/agents/test-case-tree-runner/core"
)

func TestTree(root string, opts core.Options) error {
	w := opts.Stderr
	if w == nil {
		w = os.Stderr
	}
	cases, err := core.DiscoverTreeCases(root)
	if err != nil {
		return err
	}
	if len(cases) == 0 {
		return fmt.Errorf("%s: no runnable test cases found", root)
	}
	tmp, err := os.MkdirTemp("", "test-case-tree-runner-*")
	if err != nil {
		return err
	}
	if opts.RemoveTemp {
		defer os.RemoveAll(tmp)
	}

	fmt.Fprintf(w, "→ %s\n\n", tmp)

	modRoot, modPath, hasMod := core.FindModuleRoot(root)
	if err := core.WriteGoMod(tmp, modRoot, modPath, hasMod); err != nil {
		return err
	}

	pkgName := "testcase"
	if srcDir, origPkg, ok := core.ResolvePkgUnderTest(root); ok {
		newPkg, err := core.CopySourceFiles(tmp, srcDir, origPkg)
		if err != nil {
			return fmt.Errorf("copy source files: %w", err)
		}
		pkgName = newPkg
	}

	absRoot, _ := filepath.Abs(root)
	if _, err := core.WriteGeneratedCases(tmp, cases, false, nil, pkgName, absRoot); err != nil {
		return err
	}

	fmt.Fprintf(w, "cd %s && go test -mod=mod ./...\n\n", tmp)

	var errs []error
	for _, tc := range cases {
		if opts.Verbose {
			fmt.Fprintf(w, "─── %s\n", tc.Path)
		}
		if err := runGeneratedTest(tmp, tc, opts.Count, opts.Verbose, w); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", tc.Path, err))
		}
	}
	return errors.Join(errs...)
}

func runGeneratedTest(dir string, tc core.TreeCase, count int, verbose bool, w interface{ Write(p []byte) (n int, err error) }) error {
	args := []string{"test", "-mod=mod", "-run", "^" + core.TestFuncName(tc) + "$", "./..."}
	if count > 0 {
		args = append([]string{"test", fmt.Sprintf("-count=%d", count), "-mod=mod"}, args[2:]...)
	}
	if verbose {
		args = append(args, "-v")
	}
	goTestCmd := exec.Command("go", args...)
	goTestCmd.Dir = dir
	if verbose {
		goTestCmd.Stdout = w
		goTestCmd.Stderr = w
		if err := goTestCmd.Run(); err != nil {
			return fmt.Errorf("go test failed: %v", err)
		}
		return nil
	}
	if out, err := goTestCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go test failed: %v\n%s", err, string(out))
	}
	return nil
}
