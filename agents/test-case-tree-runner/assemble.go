package main

import (
	"fmt"
	"sort"
	"strings"
)

func assembleTestSource(tc TreeCase, compileOnly bool, pkgName string) (string, error) {
	var buf strings.Builder
	buf.WriteString("package ")
	buf.WriteString(pkgName)
	buf.WriteString("\n\n")

	imports := collectImports(tc.SetupFiles, tc.AssertFile.GoBlock)
	if !imports["testing"] {
		imports["testing"] = true
	}
	if len(imports) > 0 {
		buf.WriteString("import (\n")
		importList := make([]string, 0, len(imports))
		for pkg := range imports {
			importList = append(importList, pkg)
		}
		sort.Strings(importList)
		for _, pkg := range importList {
			buf.WriteString("\t\"" + pkg + "\"\n")
		}
		buf.WriteString(")\n\n")
	}

	buf.WriteString("func ")
	buf.WriteString(testFuncName(tc))
	buf.WriteString("(t *testing.T) {\n")

	run := writeSetupDecls(&buf, tc.SetupFiles)
	writeAssertDecls(&buf, tc.AssertFile.GoBlock)

	buf.WriteString("\treq := &Request{}\n")
	writeSetupCalls(&buf, tc.SetupFiles)
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

func collectImports(setupFiles []SetupDocument, assertBlock GoBlock) map[string]bool {
	imports := make(map[string]bool)
	for _, doc := range setupFiles {
		if doc.GoBlock == nil {
			continue
		}
		for _, pkg := range doc.GoBlock.Imports {
			if pkg != "" {
				imports[pkg] = true
			}
		}
	}
	for _, pkg := range assertBlock.Imports {
		if pkg != "" {
			imports[pkg] = true
		}
	}
	return imports
}

func writeSetupDecls(buf *strings.Builder, setupFiles []SetupDocument) *FuncSnippet {
	var run *FuncSnippet
	for _, doc := range setupFiles {
		if doc.GoBlock == nil {
			continue
		}
		writeGoBlockDecls(buf, *doc.GoBlock)
		if doc.GoBlock.Run != nil {
			runCopy := *doc.GoBlock.Run
			run = &runCopy
		}
	}
	return run
}

func writeAssertDecls(buf *strings.Builder, block GoBlock) {
	writeGoBlockDecls(buf, block)
}

func writeGoBlockDecls(buf *strings.Builder, block GoBlock) {
	for _, decl := range block.TypeDecls {
		writeIndented(buf, decl)
	}
	for _, decl := range block.Consts {
		writeIndented(buf, decl)
	}
	for _, decl := range block.VarDecls {
		writeIndented(buf, decl)
	}
	for _, helper := range block.Helpers {
		writeFuncClosure(buf, helper.Name, helper)
	}
}

func writeSetupCalls(buf *strings.Builder, setupFiles []SetupDocument) {
	for i, doc := range setupFiles {
		if doc.GoBlock == nil || doc.GoBlock.Setup == nil {
			continue
		}
		name := fmt.Sprintf("setup%d", i)
		writeFuncClosure(buf, name, *doc.GoBlock.Setup)
		buf.WriteString(fmt.Sprintf("\tif err := %s(t, req); err != nil {\n", name))
		buf.WriteString(fmt.Sprintf("\t\tt.Fatalf(\"%s failed: %%v\", err)\n", escapeString(doc.Path)))
		buf.WriteString("\t}\n")
	}
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
