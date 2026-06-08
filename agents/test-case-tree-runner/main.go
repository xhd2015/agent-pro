package main

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

const usage = `Usage: test-case-tree-runner <command> [options]

Commands:
  run <dir>                  Run executable Go snippets from a test-case-tree directory.
  compile <dir>              Validate generated snippets compile without executing behavior.
  generate-code [--dry-run] <dir>
                             Scaffold missing executable Go snippets and validate compilation.
`

type GoBlock struct {
	SourcePath string
	Code       string

	TypeDecls []string
	VarDecls  []string
	Consts    []string
	Helpers   []FuncSnippet
	Setup     *FuncSnippet
	Run       *FuncSnippet
	Assert    *FuncSnippet

	Types map[string]bool
}

type FuncSnippet struct {
	Name    string
	Params  string
	Results string
	Body    string
}

type SetupDocument struct {
	Path    string
	GoBlock *GoBlock
}

type AssertDocument struct {
	Path    string
	GoBlock GoBlock
}

type TreeCase struct {
	Name       string
	Path       string
	SetupFiles []SetupDocument
	AssertFile AssertDocument
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(usage)
		return nil
	}
	switch args[0] {
	case "run":
		if len(args) != 2 {
			return fmt.Errorf("run requires <dir>")
		}
		return RunTree(args[1])
	case "compile":
		if len(args) != 2 {
			return fmt.Errorf("compile requires <dir>")
		}
		return CompileTree(args[1])
	case "generate-code":
		var opts GenerateCodeOptions
		cmdArgs := args[1:]
		for len(cmdArgs) > 0 && strings.HasPrefix(cmdArgs[0], "-") {
			switch cmdArgs[0] {
			case "--dry-run":
				opts.DryRun = true
				cmdArgs = cmdArgs[1:]
			default:
				return fmt.Errorf("unknown generate-code option: %s", cmdArgs[0])
			}
		}
		if len(cmdArgs) != 1 {
			return fmt.Errorf("generate-code requires <dir>")
		}
		return GenerateCode(cmdArgs[0], opts)
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func ExtractFinalGoBlock(path string, content string) (GoBlock, error) {
	blocks := findGoBlocks(content)
	if len(blocks) == 0 {
		return GoBlock{}, fmt.Errorf("%s: missing go block", path)
	}
	last := blocks[len(blocks)-1]
	if strings.TrimSpace(content[last.end:]) != "" {
		return GoBlock{}, fmt.Errorf("%s: go block must be final content", path)
	}
	return GoBlock{SourcePath: path, Code: last.code}, nil
}

func ParseSetupDocument(path string, content string) (SetupDocument, error) {
	blocks := findGoBlocks(content)
	if len(blocks) == 0 {
		return SetupDocument{Path: path}, nil
	}
	if len(blocks) > 1 {
		return SetupDocument{}, fmt.Errorf("%s: multiple go blocks are not allowed", path)
	}
	block, err := ExtractFinalGoBlock(path, content)
	if err != nil {
		return SetupDocument{}, err
	}
	if err := parseGoBlock(&block); err != nil {
		return SetupDocument{}, err
	}
	if block.Setup != nil && !isSetupSignature(*block.Setup) {
		return SetupDocument{}, fmt.Errorf("%s: Setup must be func Setup(t *testing.T, req *Request) error", path)
	}
	if block.Run != nil && !isRunSignature(*block.Run) {
		return SetupDocument{}, fmt.Errorf("%s: Run must be func Run(t *testing.T, req *Request) (*Response, error)", path)
	}
	return SetupDocument{Path: path, GoBlock: &block}, nil
}

func ParseAssertDocument(path string, content string) (AssertDocument, error) {
	blocks := findGoBlocks(content)
	if len(blocks) == 0 {
		return AssertDocument{}, fmt.Errorf("%s: missing go block", path)
	}
	if len(blocks) > 1 {
		return AssertDocument{}, fmt.Errorf("%s: multiple go blocks are not allowed", path)
	}
	block, err := ExtractFinalGoBlock(path, content)
	if err != nil {
		return AssertDocument{}, err
	}
	if err := parseGoBlock(&block); err != nil {
		return AssertDocument{}, err
	}
	if block.Assert == nil {
		return AssertDocument{}, fmt.Errorf("%s: missing func Assert(t *testing.T, req *Request, resp *Response, err error)", path)
	}
	if !isAssertSignature(*block.Assert) {
		return AssertDocument{}, fmt.Errorf("%s: Assert must be func Assert(t *testing.T, req *Request, resp *Response, err error)", path)
	}
	return AssertDocument{Path: path, GoBlock: block}, nil
}

type mdBlock struct {
	code string
	end  int
}

func findGoBlocks(content string) []mdBlock {
	var blocks []mdBlock
	i := 0
	for {
		start := strings.Index(content[i:], "```go")
		if start < 0 {
			return blocks
		}
		start += i
		lineEnd := strings.IndexByte(content[start:], '\n')
		if lineEnd < 0 {
			return blocks
		}
		codeStart := start + lineEnd + 1
		close := strings.Index(content[codeStart:], "```")
		if close < 0 {
			return blocks
		}
		close += codeStart
		code := content[codeStart:close]
		end := close + len("```")
		if end < len(content) && content[end] == '\r' {
			end++
		}
		if end < len(content) && content[end] == '\n' {
			end++
		}
		blocks = append(blocks, mdBlock{code: code, end: end})
		i = end
	}
}

func parseGoBlock(block *GoBlock) error {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, block.SourcePath+".go", "package testcase\n"+block.Code, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("%s: invalid go: %w", block.SourcePath, err)
	}
	block.Types = make(map[string]bool)
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			switch d.Tok {
			case token.TYPE:
				for _, spec := range d.Specs {
					if ts, ok := spec.(*ast.TypeSpec); ok {
						block.Types[ts.Name.Name] = true
					}
				}
				block.TypeDecls = append(block.TypeDecls, nodeString(fset, d))
			case token.VAR:
				block.VarDecls = append(block.VarDecls, nodeString(fset, d))
			case token.CONST:
				block.Consts = append(block.Consts, nodeString(fset, d))
			case token.IMPORT:
				// Imports are intentionally ignored here. goimports resolves
				// the generated file from actual identifier usage.
			}
		case *ast.FuncDecl:
			fn := funcSnippet(fset, d)
			switch d.Name.Name {
			case "Setup":
				block.Setup = &fn
			case "Run":
				block.Run = &fn
			case "Assert":
				block.Assert = &fn
			default:
				block.Helpers = append(block.Helpers, fn)
			}
		}
	}
	return nil
}

func nodeString(fset *token.FileSet, n any) string {
	var buf bytes.Buffer
	_ = printer.Fprint(&buf, fset, n)
	return buf.String()
}

func funcSnippet(fset *token.FileSet, d *ast.FuncDecl) FuncSnippet {
	params := ""
	if d.Type.Params != nil {
		params = fieldsString(fset, d.Type.Params)
	}
	results := ""
	if d.Type.Results != nil {
		results = resultsString(fset, d.Type.Results)
	}
	return FuncSnippet{
		Name:    d.Name.Name,
		Params:  params,
		Results: results,
		Body:    nodeString(fset, d.Body),
	}
}

func fieldsString(fset *token.FileSet, fields *ast.FieldList) string {
	if fields == nil {
		return ""
	}
	var parts []string
	for _, field := range fields.List {
		typ := nodeString(fset, field.Type)
		if len(field.Names) == 0 {
			parts = append(parts, typ)
			continue
		}
		for _, name := range field.Names {
			parts = append(parts, name.Name+" "+typ)
		}
	}
	return strings.Join(parts, ", ")
}

func resultsString(fset *token.FileSet, fields *ast.FieldList) string {
	if fields == nil || len(fields.List) == 0 {
		return ""
	}
	s := fieldsString(fset, fields)
	if len(fields.List) > 1 {
		return "(" + s + ")"
	}
	return s
}

func isSetupSignature(fn FuncSnippet) bool {
	return normalize(fn.Params) == "t*testing.T,req*Request" && normalize(fn.Results) == "error"
}

func isRunSignature(fn FuncSnippet) bool {
	return normalize(fn.Params) == "t*testing.T,req*Request" && normalize(fn.Results) == "(*Response,error)"
}

func isAssertSignature(fn FuncSnippet) bool {
	return normalize(fn.Params) == "t*testing.T,req*Request,resp*Response,errerror" && normalize(fn.Results) == ""
}

func normalize(s string) string {
	return strings.NewReplacer(" ", "", "\t", "", "\n", "").Replace(s)
}

func DiscoverTreeCases(root string) ([]TreeCase, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("not a directory: %s", root)
	}
	rootSetupPath := filepath.Join(root, "SETUP.md")
	rootSetup, err := readSetup(rootSetupPath)
	if err != nil {
		return nil, err
	}
	if rootSetup.GoBlock == nil || !rootSetup.GoBlock.Types["Request"] || !rootSetup.GoBlock.Types["Response"] {
		return nil, fmt.Errorf("%s: root SETUP.md must define type Request and type Response", root)
	}

	var cases []TreeCase
	err = filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || d.Name() != "ASSERT.md" {
			return nil
		}
		leafDir := filepath.Dir(path)
		relLeaf, err := filepath.Rel(root, leafDir)
		if err != nil {
			return err
		}
		if relLeaf == "." {
			relLeaf = ""
		}
		setupDocs, err := setupChain(root, leafDir)
		if err != nil {
			return err
		}
		assertContent, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		relAssert, _ := filepath.Rel(root, path)
		assertDoc, err := ParseAssertDocument(relAssert, string(assertContent))
		if err != nil {
			return err
		}
		cases = append(cases, TreeCase{
			Name:       caseName(relLeaf),
			Path:       relLeaf,
			SetupFiles: setupDocs,
			AssertFile: assertDoc,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(cases, func(i, j int) bool { return cases[i].Path < cases[j].Path })
	return cases, nil
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
		if doc.GoBlock != nil {
			for name := range doc.GoBlock.Types {
				if i > 0 && (name == "Request" || name == "Response") {
					relPath, _ := filepath.Rel(root, path)
					return nil, fmt.Errorf("%s: child SETUP.md cannot redefine %s", relPath, name)
				}
			}
		}
		relPath, _ := filepath.Rel(root, path)
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

func RunTree(root string) error {
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
	defer os.RemoveAll(tmp)

	var errs []error
	for _, tc := range cases {
		if err := runCase(tmp, tc); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", tc.Path, err))
		}
	}
	return errors.Join(errs...)
}

func CompileTree(root string) error {
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
	defer os.RemoveAll(tmp)

	var errs []error
	for _, tc := range cases {
		if err := compileCase(tmp, tc); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", tc.Path, err))
		}
	}
	return errors.Join(errs...)
}

func runCase(tmp string, tc TreeCase) error {
	return runGeneratedCase(tmp, tc, false)
}

func compileCase(tmp string, tc TreeCase) error {
	return runGeneratedCase(tmp, tc, true)
}

func runGeneratedCase(tmp string, tc TreeCase, compileOnly bool) error {
	dir := filepath.Join(tmp, caseName(tc.Path))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module testcase\n\ngo 1.25.0\n"), 0644); err != nil {
		return err
	}
	src, err := assembleTestSource(tc, compileOnly)
	if err != nil {
		return err
	}
	testPath := filepath.Join(dir, "generated_test.go")
	if err := os.WriteFile(testPath, []byte(src), 0644); err != nil {
		return err
	}
	goimportsCmd := exec.Command("goimports", "-w", testPath)
	goimportsCmd.Dir = dir
	if out, err := goimportsCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("goimports failed: %v\n%s", err, string(out))
	}
	goTestCmd := exec.Command("go", "test", ".")
	goTestCmd.Dir = dir
	if out, err := goTestCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go test failed: %v\n%s", err, string(out))
	}
	return nil
}

func assembleTestSource(tc TreeCase, compileOnly bool) (string, error) {
	var buf strings.Builder
	buf.WriteString("package testcase\n\n")
	buf.WriteString("import \"testing\"\n\n")
	buf.WriteString("func TestGeneratedCase(t *testing.T) {\n")

	var run *FuncSnippet
	for _, doc := range tc.SetupFiles {
		if doc.GoBlock == nil {
			continue
		}
		for _, decl := range doc.GoBlock.TypeDecls {
			writeIndented(&buf, decl)
		}
		for _, decl := range doc.GoBlock.Consts {
			writeIndented(&buf, decl)
		}
		for _, decl := range doc.GoBlock.VarDecls {
			writeIndented(&buf, decl)
		}
		for _, helper := range doc.GoBlock.Helpers {
			writeFuncClosure(&buf, helper.Name, helper)
		}
		if doc.GoBlock.Run != nil {
			runCopy := *doc.GoBlock.Run
			run = &runCopy
		}
	}
	for _, decl := range tc.AssertFile.GoBlock.TypeDecls {
		writeIndented(&buf, decl)
	}
	for _, decl := range tc.AssertFile.GoBlock.Consts {
		writeIndented(&buf, decl)
	}
	for _, decl := range tc.AssertFile.GoBlock.VarDecls {
		writeIndented(&buf, decl)
	}
	for _, helper := range tc.AssertFile.GoBlock.Helpers {
		writeFuncClosure(&buf, helper.Name, helper)
	}

	buf.WriteString("\treq := &Request{}\n")
	for i, doc := range tc.SetupFiles {
		if doc.GoBlock == nil || doc.GoBlock.Setup == nil {
			continue
		}
		name := fmt.Sprintf("setup%d", i)
		writeFuncClosure(&buf, name, *doc.GoBlock.Setup)
		buf.WriteString(fmt.Sprintf("\tif err := %s(t, req); err != nil {\n", name))
		buf.WriteString(fmt.Sprintf("\t\tt.Fatalf(\"%s failed: %%v\", err)\n", escapeString(doc.Path)))
		buf.WriteString("\t}\n")
	}
	if run == nil {
		return "", fmt.Errorf("missing Run(t *testing.T, req *Request) (*Response, error) in setup chain")
	}
	writeFuncClosure(&buf, "run", *run)
	writeFuncClosure(&buf, "assert", *tc.AssertFile.GoBlock.Assert)
	if compileOnly {
		buf.WriteString("\t_ = req\n")
		buf.WriteString("\t_ = run\n")
		buf.WriteString("\t_ = assert\n")
		buf.WriteString("\tvar resp *Response\n")
		buf.WriteString("\tvar runErr error\n")
		buf.WriteString("\t_ = resp\n")
		buf.WriteString("\t_ = runErr\n")
		buf.WriteString("}\n")
		return buf.String(), nil
	}
	buf.WriteString("\tresp, runErr := run(t, req)\n")
	buf.WriteString("\tassert(t, req, resp, runErr)\n")
	buf.WriteString("}\n")
	return buf.String(), nil
}

func writeFuncClosure(buf *strings.Builder, name string, fn FuncSnippet) {
	results := ""
	if strings.TrimSpace(fn.Results) != "" {
		results = " " + fn.Results
	}
	buf.WriteString(fmt.Sprintf("\t%s := func(%s)%s %s\n", name, fn.Params, results, fn.Body))
}

func writeIndented(buf *strings.Builder, s string) {
	for _, line := range strings.Split(strings.TrimSpace(s), "\n") {
		buf.WriteString("\t")
		buf.WriteString(line)
		buf.WriteString("\n")
	}
}

func escapeString(s string) string {
	return strings.ReplaceAll(s, "\\", "\\\\")
}

type GenerateCodeOptions struct {
	DryRun bool
}

type FileEdit struct {
	Path    string
	OldText string
	NewText string
	Reason  string
}

type CodegenPlan struct {
	Root      string
	FileEdits []FileEdit
}

func GenerateCode(root string, opts GenerateCodeOptions) error {
	plan, err := BuildCodegenPlan(root)
	if err != nil {
		return err
	}
	if opts.DryRun {
		return nil
	}
	for _, edit := range plan.FileEdits {
		if err := os.MkdirAll(filepath.Dir(edit.Path), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(edit.Path, []byte(edit.NewText), 0644); err != nil {
			return err
		}
	}
	return CompileTree(root)
}

func BuildCodegenPlan(root string) (CodegenPlan, error) {
	info, err := os.Stat(root)
	if err != nil {
		return CodegenPlan{}, err
	}
	if !info.IsDir() {
		return CodegenPlan{}, fmt.Errorf("not a directory: %s", root)
	}
	var plan CodegenPlan
	plan.Root = root

	rootSetupPath := filepath.Join(root, "SETUP.md")
	rootText, err := readTextIfExists(rootSetupPath)
	if err != nil {
		return CodegenPlan{}, err
	}
	rootNew, rootChanged, err := scaffoldSetup(rootSetupPath, rootText, true)
	if err != nil {
		return CodegenPlan{}, err
	}
	if rootChanged {
		plan.FileEdits = append(plan.FileEdits, FileEdit{
			Path:    rootSetupPath,
			OldText: rootText,
			NewText: rootNew,
			Reason:  "ensure root Request/Response/Run scaffold",
		})
	}

	leafDirs, err := discoverAssertDirs(root)
	if err != nil {
		return CodegenPlan{}, err
	}
	for _, leafDir := range leafDirs {
		setupPath := filepath.Join(leafDir, "SETUP.md")
		setupText, err := readTextIfExists(setupPath)
		if err != nil {
			return CodegenPlan{}, err
		}
		setupNew, setupChanged, err := scaffoldSetup(setupPath, setupText, false)
		if err != nil {
			return CodegenPlan{}, err
		}
		if setupChanged {
			plan.FileEdits = append(plan.FileEdits, FileEdit{
				Path:    setupPath,
				OldText: setupText,
				NewText: setupNew,
				Reason:  "ensure leaf Setup scaffold",
			})
		}

		assertPath := filepath.Join(leafDir, "ASSERT.md")
		assertText, err := readTextIfExists(assertPath)
		if err != nil {
			return CodegenPlan{}, err
		}
		assertNew, assertChanged, err := scaffoldAssert(assertPath, assertText)
		if err != nil {
			return CodegenPlan{}, err
		}
		if assertChanged {
			plan.FileEdits = append(plan.FileEdits, FileEdit{
				Path:    assertPath,
				OldText: assertText,
				NewText: assertNew,
				Reason:  "ensure Assert scaffold",
			})
		}
	}
	return plan, nil
}

func readTextIfExists(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

func discoverAssertDirs(root string) ([]string, error) {
	var dirs []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || d.Name() != "ASSERT.md" {
			return nil
		}
		dirs = append(dirs, filepath.Dir(path))
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(dirs)
	return dirs, nil
}

func scaffoldSetup(path string, text string, root bool) (string, bool, error) {
	blocks := findGoBlocks(text)
	if len(blocks) > 1 {
		return "", false, fmt.Errorf("%s: multiple go blocks are not allowed", path)
	}
	if len(blocks) == 1 {
		if strings.TrimSpace(text[blocks[0].end:]) != "" {
			return "", false, fmt.Errorf("%s: go block must be final content", path)
		}
		doc, err := ParseSetupDocument(path, text)
		if err != nil {
			return "", false, err
		}
		if doc.GoBlock == nil {
			return text, false, nil
		}
		if root && (!doc.GoBlock.Types["Request"] || !doc.GoBlock.Types["Response"] || doc.GoBlock.Run == nil) {
			return "", false, fmt.Errorf("%s: existing root SETUP.md go block must define Request, Response, and Run", path)
		}
		return text, false, nil
	}
	if root {
		return appendGoBlock(text, rootSetupScaffold()), true, nil
	}
	return appendGoBlock(text, leafSetupScaffold()), true, nil
}

func scaffoldAssert(path string, text string) (string, bool, error) {
	blocks := findGoBlocks(text)
	if len(blocks) > 1 {
		return "", false, fmt.Errorf("%s: multiple go blocks are not allowed", path)
	}
	if len(blocks) == 1 {
		if strings.TrimSpace(text[blocks[0].end:]) != "" {
			return "", false, fmt.Errorf("%s: go block must be final content", path)
		}
		if _, err := ParseAssertDocument(path, text); err != nil {
			return "", false, err
		}
		return text, false, nil
	}
	return appendGoBlock(text, assertScaffold()), true, nil
}

func appendGoBlock(text string, code string) string {
	text = strings.TrimRight(text, " \t\r\n")
	if text != "" {
		text += "\n\n"
	}
	return text + "```go\n" + strings.TrimSpace(code) + "\n```\n"
}

func rootSetupScaffold() string {
	return `
import "fmt"

type Request struct {
}

type Response struct {
}

func Run(t *testing.T, req *Request) (*Response, error) {
	return nil, fmt.Errorf("not implemented yet")
}
`
}

func leafSetupScaffold() string {
	return `
func Setup(t *testing.T, req *Request) error {
	return nil
}
`
}

func assertScaffold() string {
	return `
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
}
`
}
