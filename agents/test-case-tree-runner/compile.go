package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xhd2015/agent-pro/agents/test-case-tree-runner/rules"
)

func DiscoverTreeCases(root string) ([]TreeCase, error) {
	return discoverTreeCasesInternal(root, nil)
}

func discoverTreeCasesVerbose(root string, w io.Writer) ([]TreeCase, error) {
	return discoverTreeCasesInternal(root, w)
}

func discoverTreeCasesInternal(root string, w io.Writer) ([]TreeCase, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("not a directory: %s", root)
	}

	var verrs []ValidationError

	rootSetupPath := filepath.Join(root, "SETUP.md")
	rootSetupContent, err := os.ReadFile(rootSetupPath)
	if err != nil {
		return nil, err
	}
	rootSetup, err := ParseSetupDocument(rootSetupPath, string(rootSetupContent))
	if err != nil {
		verrs = append(verrs, ValidationError{Path: "SETUP.md", Msg: err.Error()})
	}
	if rootSetup.GoBlock == nil {
		verrs = append(verrs, ValidationError{Path: "SETUP.md", Msg: "must have a Go code block"})
	} else {
		if v := rules.CheckRootHasRequestResponse(rootSetup.GoBlock.Types, "SETUP.md"); v != nil {
			verrs = append(verrs, ValidationError{Path: v.Path, Msg: v.Msg})
		}
		if v := rules.CheckRootHasSetupOrRun(rootSetup.GoBlock.Setup != nil, rootSetup.GoBlock.Run != nil, "SETUP.md"); v != nil {
			verrs = append(verrs, ValidationError{Path: v.Path, Msg: v.Msg})
		}
	}

	if w != nil {
		printSetupVerbose(w, rootSetup, "SETUP.md")
	}

	printedSetupDirs := make(map[string]bool)
	printedSetupDirs[root] = true
	var cases []TreeCase
	err = filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			verrs = append(verrs, ValidationError{Path: path, Msg: walkErr.Error()})
			return nil
		}
		if d.IsDir() {
			if d.Name() == "testdata" {
				return filepath.SkipDir
			}
			if path == root {
				return nil
			}
			relPath, _ := filepath.Rel(root, path)
			if w != nil {
				setupPath := filepath.Join(path, "SETUP.md")
				if !printedSetupDirs[setupPath] {
					doc, readErr := readSetup(setupPath)
					if readErr == nil && doc.GoBlock != nil {
						printedSetupDirs[setupPath] = true
						printSetupVerbose(w, doc, filepath.Join(relPath, "SETUP.md"))
					}
				}
			}
			setupPath := filepath.Join(path, "SETUP.md")
			doc, readErr := readSetup(setupPath)
			if readErr != nil {
				rel, _ := filepath.Rel(root, setupPath)
				verrs = append(verrs, ValidationError{Path: rel, Msg: readErr.Error()})
			} else if doc.GoBlock == nil {
				rel, _ := filepath.Rel(root, setupPath)
				verrs = append(verrs, ValidationError{Path: rel, Msg: "must have a Go code block"})
			} else if doc.GoBlock.Setup == nil && doc.GoBlock.Run == nil {
				rel, _ := filepath.Rel(root, setupPath)
				verrs = append(verrs, ValidationError{Path: rel, Msg: "must have func Setup or func Run"})
			}
			return nil
		}
		if d.Name() != "ASSERT.md" {
			return nil
		}
		leafDir := filepath.Dir(path)
		relLeaf, err := filepath.Rel(root, leafDir)
		if err != nil {
			verrs = append(verrs, ValidationError{Path: path, Msg: err.Error()})
			return nil
		}
		if relLeaf == "." {
			relLeaf = ""
		}
		setupDocs, chainErr := setupChain(root, leafDir)
		if chainErr != nil {
			relAssert, _ := filepath.Rel(root, path)
			verrs = append(verrs, ValidationError{Path: relAssert, Msg: chainErr.Error()})
			return nil
		}
		assertContent, err := os.ReadFile(path)
		if err != nil {
			relAssert, _ := filepath.Rel(root, path)
			verrs = append(verrs, ValidationError{Path: relAssert, Msg: err.Error()})
			return nil
		}
		relAssert, _ := filepath.Rel(root, path)
		assertDoc, err := ParseAssertDocument(relAssert, string(assertContent))
		if err != nil {
			verrs = append(verrs, ValidationError{Path: relAssert, Msg: err.Error()})
			return nil
		}
		if v := rules.CheckChainHasRun(runSource(setupDocs), relAssert); v != nil {
			verrs = append(verrs, ValidationError{Path: v.Path, Msg: v.Msg})
		}
		tc := TreeCase{
			Name:       caseName(relLeaf),
			Path:       relLeaf,
			SetupFiles: setupDocs,
			AssertFile: assertDoc,
		}
		cases = append(cases, tc)

		if w != nil {
			fmt.Fprintf(w, "  ✦ %-30s (Run: %s)\n", caseName(relLeaf), runSource(setupDocs))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(verrs) > 0 {
		return nil, joinValidationErrors(verrs)
	}
	sort.Slice(cases, func(i, j int) bool { return cases[i].Path < cases[j].Path })
	return cases, nil
}

func joinValidationErrors(verrs []ValidationError) error {
	var msgs []string
	for _, ve := range verrs {
		msgs = append(msgs, fmt.Sprintf("%s: %s", ve.Path, ve.Msg))
	}
	all := strings.Join(msgs, "\n")
	return fmt.Errorf("%d validation errors:\n%s", len(verrs), all)
}

func printSetupVerbose(w io.Writer, doc SetupDocument, relPath string) {
	if doc.GoBlock == nil {
		fmt.Fprintf(w, "%s\n", relPath)
		return
	}
	block := doc.GoBlock
	var parts []string
	if block.Types["Request"] {
		parts = append(parts, "Request")
	}
	if block.Types["Response"] {
		parts = append(parts, "Response")
	}
	if block.Setup != nil {
		parts = append(parts, "Setup")
	}
	if block.Run != nil {
		parts = append(parts, "Run")
	}
	if len(block.Helpers) > 0 {
		parts = append(parts, fmt.Sprintf("%d helpers", len(block.Helpers)))
	}
	fmt.Fprintf(w, "%s — %s", relPath, strings.Join(parts, ", "))
	if block.Run != nil {
		fmt.Fprintf(w, " (defines Run)")
	}
	fmt.Fprintln(w)
}

func runSource(setupDocs []SetupDocument) string {
	for i := len(setupDocs) - 1; i >= 0; i-- {
		doc := setupDocs[i]
		if doc.GoBlock != nil && doc.GoBlock.Run != nil {
			return doc.Path
		}
	}
	return "none"
}

func readSetup(path string) (SetupDocument, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return SetupDocument{Path: path}, nil
		}
		return SetupDocument{}, err
	}
	return ParseSetupDocument(path, string(content))
}

func setupChain(root, leafDir string) ([]SetupDocument, error) {
	rel, err := filepath.Rel(root, leafDir)
	if err != nil {
		return nil, err
	}
	var parts []string
	if rel != "." {
		parts = strings.Split(rel, string(filepath.Separator))
	}
	var docs []SetupDocument
	for i := 0; i <= len(parts); i++ {
		dir := filepath.Join(append([]string{root}, parts[:i]...)...)
		path := filepath.Join(dir, "SETUP.md")
		doc, err := readSetup(path)
		if err != nil {
			return nil, err
		}
		relPath, _ := filepath.Rel(root, path)
		if doc.GoBlock != nil {
			if v := rules.CheckChildNoRedefine(doc.GoBlock.Types, relPath, i); v != nil {
				return nil, fmt.Errorf("%s: %s", v.Path, v.Msg)
			}
		}
		doc.Path = relPath
		docs = append(docs, doc)
	}
	return docs, nil
}

func caseName(path string) string {
	if path == "" {
		return "root"
	}
	return strings.NewReplacer("/", "_", string(filepath.Separator), "_", "-", "_").Replace(path)
}

func testFileName(tc TreeCase) string {
	return caseName(tc.Path) + "_test.go"
}

func testFuncName(tc TreeCase) string {
	name := caseName(tc.Path)
	var b strings.Builder
	b.WriteString("TestGeneratedCase")
	upperNext := true
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			if upperNext && r >= 'a' && r <= 'z' {
				r = r - 'a' + 'A'
			}
			b.WriteRune(r)
			upperNext = false
			continue
		}
		upperNext = true
	}
	return b.String()
}

func RunTree(root string, opts CompileOptions) error {
	w := opts.Stderr
	if w == nil {
		w = os.Stderr
	}
	cases, err := DiscoverTreeCases(root)
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

	modRoot, modPath, hasMod := findModuleRoot(root)
	if err := writeGoMod(tmp, modRoot, modPath, hasMod); err != nil {
		return err
	}

	pkgName := "testcase"
	if srcDir, origPkg, ok := resolvePkgUnderTest(root); ok {
		newPkg, err := copySourceFiles(tmp, srcDir, origPkg)
		if err != nil {
			return fmt.Errorf("copy source files: %w", err)
		}
		pkgName = newPkg
	}

	absRoot, _ := filepath.Abs(root)
	if _, err := writeGeneratedCases(tmp, cases, false, nil, pkgName, absRoot); err != nil {
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

func CompileTree(root string) error {
	return CompileTreeWithOptions(root, CompileOptions{})
}

func CompileTreeWithOptions(root string, opts CompileOptions) error {
	w := opts.Stderr
	if w == nil {
		w = os.Stderr
	}

	var cases []TreeCase
	var err error
	if opts.Verbose {
		fmt.Fprintf(w, "test-case-tree-runner: %s\n\n", root)
		cases, err = discoverTreeCasesVerbose(root, w)
		if err != nil {
			return err
		}
		fmt.Fprintf(w, "─── %d test cases\n\n", len(cases))
	} else {
		cases, err = DiscoverTreeCases(root)
		if err != nil {
			return err
		}
		fmt.Fprintf(w, "test-case-tree-runner: %s\n", root)
		fmt.Fprintf(w, "─── %d test cases\n", len(cases))
	}

	if len(cases) == 0 {
		return fmt.Errorf("%s: no runnable test cases found", root)
	}

	tmp := opts.GenDir
	removeTmp := false
	if tmp == "" {
		tmp, err = os.MkdirTemp("", "test-case-tree-runner-*")
		if err != nil {
			return err
		}
		removeTmp = opts.RemoveTemp
	} else if err := os.MkdirAll(tmp, 0755); err != nil {
		return err
	}

	fmt.Fprintf(w, "→ %s\n\n", tmp)

	if removeTmp {
		defer os.RemoveAll(tmp)
	}

	modRoot, modPath, hasMod := findModuleRoot(root)
	if err := writeGoMod(tmp, modRoot, modPath, hasMod); err != nil {
		return err
	}
	if opts.Verbose {
		fmt.Fprintf(w, "→ %s\n", filepath.Join(tmp, "go.mod"))
	}

	pkgName := "testcase"
	if srcDir, origPkg, ok := resolvePkgUnderTest(root); ok {
		newPkg, err := copySourceFiles(tmp, srcDir, origPkg)
		if err != nil {
			return fmt.Errorf("copy source files: %w", err)
		}
		pkgName = newPkg
		if opts.Verbose {
			fmt.Fprintf(w, "→ %s (copied from %s, package %s)\n", srcDir, srcDir, newPkg)
		}
	}

	absRoot, _ := filepath.Abs(root)
	_, err = writeGeneratedCases(tmp, cases, true, w, pkgName, absRoot)
	if err != nil {
		return err
	}

	goBuildArgs := []string{"build", "-mod=mod"}
	if opts.Verbose {
		goBuildArgs = append(goBuildArgs, "-v")
	}
	goBuildArgs = append(goBuildArgs, "./...")

	fmt.Fprintf(w, "cd %s && go %s\n\n", tmp, strings.Join(goBuildArgs, " "))

	goBuildCmd := exec.Command("go", goBuildArgs...)
	goBuildCmd.Dir = tmp

	if opts.Verbose {
		goBuildCmd.Stdout = w
		goBuildCmd.Stderr = w
		if err := goBuildCmd.Run(); err != nil {
			return fmt.Errorf("go build failed: %v", err)
		}
	} else {
		if out, err := goBuildCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("go build failed: %v\n%s", err, string(out))
		}
	}
	return nil
}

func writeGeneratedCases(dir string, cases []TreeCase, compileOnly bool, w io.Writer, pkgName string, originalDir string) ([]string, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	var testPaths []string
	var testFiles []string
	first := true
	for _, tc := range cases {
		src, err := assembleTestSource(tc, compileOnly, pkgName)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", tc.Path, err)
		}
		if first && originalDir != "" {
			src = insertInitChdir(src, originalDir)
			first = false
		}
		testFile := testFileName(tc)
		testPath := filepath.Join(dir, testFile)
		if err := os.WriteFile(testPath, []byte(src), 0644); err != nil {
			return nil, err
		}
		if w != nil {
			fmt.Fprintf(w, "→ %s\n", testPath)
		}
		testPaths = append(testPaths, testPath)
		testFiles = append(testFiles, testFile)
	}
	args := append([]string{"-w"}, testPaths...)
	goimportsCmd := exec.Command("goimports", args...)
	goimportsCmd.Dir = dir
	if out, err := goimportsCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("goimports failed: %v\n%s", err, string(out))
	}
	return testFiles, nil
}

func insertInitChdir(src string, dir string) string {
	escaped := strings.ReplaceAll(dir, "\\", "\\\\")
	initCode := "\nimport \"os\"\n\nfunc init() { os.Chdir(`" + escaped + "`) }\n"
	idx := strings.Index(src, ")\n\nfunc ")
	if idx >= 0 {
		return src[:idx+2] + initCode + src[idx+2:]
	}
	i := strings.Index(src, "\n")
	if i < 0 {
		return src
	}
	return src[:i+1] + initCode + src[i+1:]
}

func findModuleRoot(dir string) (modRoot string, modPath string, ok bool) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", "", false
	}
	for {
		modFile := filepath.Join(dir, "go.mod")
		data, err := os.ReadFile(modFile)
		if err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "module ") {
					modPath = strings.TrimSpace(strings.TrimPrefix(line, "module "))
					return dir, modPath, true
				}
			}
			return "", "", false
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", "", false
		}
		dir = parent
	}
}

func writeGoMod(genDir, modRoot, modPath string, hasMod bool) error {
	var content string
	if hasMod {
		content = fmt.Sprintf("module testcase\n\ngo 1.21\n\nrequire %s v0.0.0\n\nreplace %s => %s\n", modPath, modPath, modRoot)
	} else {
		content = "module testcase\n\ngo 1.21\n"
	}
	if err := os.WriteFile(filepath.Join(genDir, "go.mod"), []byte(content), 0644); err != nil {
		return err
	}
	if hasMod {
		srcGoSum := filepath.Join(modRoot, "go.sum")
		if data, err := os.ReadFile(srcGoSum); err == nil {
			if err := os.WriteFile(filepath.Join(genDir, "go.sum"), data, 0644); err != nil {
				return err
			}
		}
	}
	return nil
}

func resolvePkgUnderTest(root string) (srcDir string, origPkgName string, ok bool) {
	rootSetupPath := filepath.Join(root, "SETUP.md")
	content, err := os.ReadFile(rootSetupPath)
	if err != nil {
		return "", "", false
	}
	pkgName := ""
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "- Package under test:") {
			pkgName = strings.Trim(line[len("- Package under test:"):], " `")
		}
	}
	if pkgName == "" {
		return "", "", false
	}
	modRoot, _, hasMod := findModuleRoot(root)
	if !hasMod {
		return "", "", false
	}
	absDir := filepath.Join(modRoot, pkgName)
	entries, err := os.ReadDir(absDir)
	if err != nil {
		return "", "", false
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(absDir, name))
		if err != nil {
			continue
		}
		text := string(data)
		var pkgLine string
		if strings.HasPrefix(text, "package ") {
			pkgLine = text[len("package "):]
		} else {
			i := strings.Index(text, "\npackage ")
			if i < 0 {
				continue
			}
			pkgLine = text[i+len("\npackage "):]
		}
		j := strings.IndexAny(pkgLine, " \t\n\r;")
		if j < 0 {
			origPkgName = strings.TrimSpace(pkgLine)
		} else {
			origPkgName = pkgLine[:j]
		}
		if origPkgName != "" {
			return absDir, origPkgName, true
		}
	}
	return "", "", false
}

func copySourceFiles(genDir, srcDir, origPkgName string) (string, error) {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return "", err
	}
	newPkgName := origPkgName + "_tc"
	newPkgDecl := "package " + newPkgName
	oldPkgDecl := "package " + origPkgName
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(srcDir, name))
		if err != nil {
			return "", err
		}
		content := strings.Replace(string(data), oldPkgDecl, newPkgDecl, 1)
		dst := filepath.Join(genDir, name)
		if err := os.WriteFile(dst, []byte(content), 0644); err != nil {
			return "", err
		}
	}
	return newPkgName, nil
}

func runGeneratedTest(dir string, tc TreeCase, count int, verbose bool, w io.Writer) error {
	args := []string{"test", "-mod=mod", "-run", "^" + testFuncName(tc) + "$", "./..."}
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
