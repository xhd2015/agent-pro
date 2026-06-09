package main

import (
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
	tmp, err := os.MkdirTemp("", "test-case-tree-runner-*")
	if err != nil {
		return err
	}
	if opts.RemoveTemp {
		defer os.RemoveAll(tmp)
	}

	fmt.Fprintf(w, "→ %s\n\n", tmp)

	cases, err := core.DiscoverTreeCases(root)
	if err != nil {
		return err
	}
	if len(cases) == 0 {
		return fmt.Errorf("%s: no runnable test cases found", root)
	}

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

	testBinPath := filepath.Join(tmp, "test.bin")

	fmt.Fprintf(w, "cd %s && go test -c -o test.bin . && cd %s && %s/test.bin\n\n", tmp, absRoot, tmp)

	goTestBuild := exec.Command("go", "test", "-c", "-mod=mod", "-o", testBinPath, ".")
	goTestBuild.Dir = tmp
	if out, err := goTestBuild.CombinedOutput(); err != nil {
		return fmt.Errorf("go test -c failed: %v\n%s", err, string(out))
	}

	args := []string{}
	if opts.Count > 0 {
		args = append(args, fmt.Sprintf("-test.count=%d", opts.Count))
	}
	if opts.Verbose {
		args = append(args, "-test.v")
	}
	runCmd := exec.Command(testBinPath, args...)
	runCmd.Dir = absRoot
	if opts.Verbose {
		runCmd.Stdout = w
		runCmd.Stderr = w
		if err := runCmd.Run(); err != nil {
			return fmt.Errorf("test binary failed: %v", err)
		}
		return nil
	}
	if out, err := runCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("test binary failed: %v\n%s", err, string(out))
	}
	return nil
}
