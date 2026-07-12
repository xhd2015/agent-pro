// open-bind-parallel-flake: reproduce / verify parallel doctest flakes in
// status-resume/run-open-background-bind.
//
// Modes:
//
//	(default) repro  — expect at least one of the known flake leaves to fail under
//	                   parallel package test with:
//	                     agent-run: error: grok session id not resolved
//	                   Exit 1 + REPRO: lines when symptom present (loop RED).
//	--expect=healthy — expect isolation + parallel group fully green.
//	                   Exit 0 only when all pass (VERIFY gate after fix).
//
// Usage:
//
//	go run ./script/debug/open-bind-parallel-flake
//	go run ./script/debug/open-bind-parallel-flake --expect=healthy
//	go run ./script/debug/open-bind-parallel-flake --scope=full-suite
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	// Known flake leaves (pass isolation, fail under parallel load).
	leafHardRequire     = "hard-require-without-grok-home-env"
	leafPromptFallback  = "prompt-fallback-cwd-mismatch"
	symptomNotResolved  = "grok session id not resolved"
	groupRel            = "cmd/agent-run/tests/status-resume/run-open-background-bind"
	suiteRel            = "cmd/agent-run/tests/status-resume"
)

func main() {
	expect := flag.String("expect", "repro", "repro | healthy")
	scope := flag.String("scope", "group", "group (open-background-bind only) | full-suite (all status-resume)")
	skipIsolation := flag.Bool("skip-isolation", false, "skip isolation preflight (faster repro-only)")
	flag.Parse()

	root, err := gitRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: git root: %v\n", err)
		os.Exit(2)
	}

	mode := strings.ToLower(strings.TrimSpace(*expect))
	if mode != "repro" && mode != "healthy" {
		fmt.Fprintf(os.Stderr, "FAIL: --expect must be repro or healthy\n")
		os.Exit(2)
	}

	// Isolation: known leaves must PASS alone (establishes "parallel-only" nature).
	if !*skipIsolation {
		for _, leaf := range []string{leafHardRequire, leafPromptFallback} {
			path := filepath.Join(root, groupRel, leaf)
			out, code, dur := runDoctest(path, 120*time.Second)
			fmt.Printf("isolation %s: exit=%d elapsed=%s\n", leaf, code, dur.Round(time.Millisecond))
			if code != 0 {
				fmt.Fprintf(os.Stderr, "FAIL: isolation must pass for %s (not a parallel flake if isolation fails)\n%s\n", leaf, trimOut(out, 40))
				os.Exit(2)
			}
		}
	}

	// Parallel trigger.
	var target string
	switch strings.ToLower(strings.TrimSpace(*scope)) {
	case "group", "":
		target = filepath.Join(root, groupRel)
	case "full-suite", "suite":
		target = filepath.Join(root, suiteRel)
	default:
		fmt.Fprintf(os.Stderr, "FAIL: --scope must be group or full-suite\n")
		os.Exit(2)
	}

	out, code, dur := runDoctest(target, 180*time.Second)
	fmt.Printf("parallel %s: exit=%d elapsed=%s\n", *scope, code, dur.Round(time.Millisecond))

	hits := findSymptomHits(out)
	passCount, failCount := parsePassFail(out)

	switch mode {
	case "repro":
		// Symptom present when parallel fails with known not-resolved errors
		// on either flake leaf (or generic not-resolved from those tests).
		if code != 0 && len(hits) > 0 {
			fmt.Println("REPRO: parallel open-bind flake reproduced")
			for _, h := range hits {
				fmt.Printf("REPRO: %s\n", h)
			}
			fmt.Printf("REPRO: summary pass=%d fail=%d (doctest exit %d)\n", passCount, failCount, code)
			// Evidence tail
			fmt.Println("--- doctest tail ---")
			fmt.Print(trimOut(out, 50))
			os.Exit(1) // non-zero = symptom present (bug-repro inspect RED)
		}
		if code != 0 {
			fmt.Fprintf(os.Stderr, "FAIL: parallel suite failed but symptom %q not found\n%s\n", symptomNotResolved, trimOut(out, 40))
			os.Exit(2)
		}
		fmt.Fprintf(os.Stderr, "FAIL: no flake this run (parallel green). Re-run or use --scope=full-suite; flake is usually high under group parallel.\n")
		fmt.Print(trimOut(out, 20))
		os.Exit(2)

	case "healthy":
		if code == 0 && len(hits) == 0 {
			fmt.Println("VERIFY: parallel open-bind group/suite fully green")
			fmt.Printf("VERIFY: pass=%d fail=%d\n", passCount, failCount)
			os.Exit(0)
		}
		fmt.Println("FAIL: expected healthy parallel suite")
		for _, h := range hits {
			fmt.Printf("FAIL: residual symptom: %s\n", h)
		}
		fmt.Print(trimOut(out, 50))
		os.Exit(1)
	}
}

func gitRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		// fallback: cwd
		wd, werr := os.Getwd()
		if werr != nil {
			return "", err
		}
		return wd, nil
	}
	return strings.TrimSpace(buf.String()), nil
}

func runDoctest(path string, timeout time.Duration) (output string, exitCode int, elapsed time.Duration) {
	start := time.Now()
	ctxCmd := exec.Command("doctest", "test", "-count=1", path)
	var buf bytes.Buffer
	ctxCmd.Stdout = &buf
	ctxCmd.Stderr = &buf
	// Soft timeout via process kill if needed.
	done := make(chan error, 1)
	go func() { done <- ctxCmd.Run() }()
	select {
	case err := <-done:
		elapsed = time.Since(start)
		output = buf.String()
		if err == nil {
			return output, 0, elapsed
		}
		if ee, ok := err.(*exec.ExitError); ok {
			return output, ee.ExitCode(), elapsed
		}
		return output, 2, elapsed
	case <-time.After(timeout):
		_ = ctxCmd.Process.Kill()
		<-done
		elapsed = time.Since(start)
		return buf.String() + "\nFAIL: doctest timeout\n", 124, elapsed
	}
}

var (
	reFailPkg = regexp.MustCompile(`FAIL\s+testcase/agent-run/tests/status-resume/run-open-background-bind/([^\s]+)`)
	rePassFail = regexp.MustCompile(`\((\d+)\s+Run,\s+(\d+)\s+Pass,\s+(\d+)\s+Fail`)
)

func findSymptomHits(out string) []string {
	var hits []string
	// Package-level FAIL lines for flake leaves
	for _, m := range reFailPkg.FindAllStringSubmatch(out, -1) {
		leaf := m[1]
		if leaf == leafHardRequire || leaf == leafPromptFallback ||
			strings.Contains(leaf, leafHardRequire) || strings.Contains(leaf, leafPromptFallback) {
			hits = append(hits, "FAIL package leaf="+leaf)
		}
	}
	// Symptom text
	if strings.Contains(out, symptomNotResolved) {
		// Count occurrences roughly
		n := strings.Count(out, symptomNotResolved)
		hits = append(hits, fmt.Sprintf("%q x%d", symptomNotResolved, n))
	}
	// Test name fails
	if strings.Contains(out, "TestGeneratedCaseHardRequireWithoutGrokHomeEnv") && strings.Contains(out, "--- FAIL") {
		hits = append(hits, "TestGeneratedCaseHardRequireWithoutGrokHomeEnv FAIL")
	}
	if strings.Contains(out, "TestGeneratedCasePromptFallbackCwdMismatch") && strings.Contains(out, "--- FAIL") {
		hits = append(hits, "TestGeneratedCasePromptFallbackCwdMismatch FAIL")
	}
	return unique(hits)
}

func parsePassFail(out string) (pass, fail int) {
	// Prefer last summary line
	ms := rePassFail.FindAllStringSubmatch(out, -1)
	if len(ms) == 0 {
		return 0, 0
	}
	last := ms[len(ms)-1]
	// groups: run, pass, fail
	var p, f int
	fmt.Sscanf(last[2], "%d", &p)
	fmt.Sscanf(last[3], "%d", &f)
	return p, f
}

func unique(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}

func trimOut(s string, maxLines int) string {
	lines := strings.Split(s, "\n")
	if len(lines) <= maxLines {
		return s
	}
	return strings.Join(lines[len(lines)-maxLines:], "\n")
}
