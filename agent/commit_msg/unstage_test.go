package commit_msg

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// candidateSummary renders candidates for order-sensitive comparisons.
func candidateSummary(candidates []UnstageCandidate) []string {
	summary := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		summary = append(summary, candidate.Kind+" "+candidate.Path+" submodule-dir="+candidate.SubmoduleDir)
	}
	return summary
}

func TestDetectUnstageCandidates_BinaryAndSubmodule(t *testing.T) {
	dir := initGitRepo(t)
	t.Chdir(dir)

	writeFile(t, "icon.png", []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00})
	writeFile(t, "README.md", []byte("hello\n"))
	mustRun(t, "mkdir", "-p", filepath.Join(dir, "vendor", "libfoo", "src"))
	os.MkdirAll(filepath.Join(dir, "vendor", "libfoo", ".git"), 0755)
	writeFile(t, "vendor/libfoo/src/main.c", []byte("int main() { return 0; }\n"))
	mustRun(t, "git", "add", "icon.png", "README.md", "vendor/libfoo/src/main.c")

	candidates, err := DetectUnstageCandidates(dir, nil)
	if err != nil {
		t.Fatalf("DetectUnstageCandidates: %v", err)
	}

	want := []string{
		"binary icon.png submodule-dir=",
		"submodule vendor/libfoo/src/main.c submodule-dir=vendor/libfoo",
	}
	if got := candidateSummary(candidates); !reflect.DeepEqual(got, want) {
		t.Fatalf("candidates = %#v, want %#v", got, want)
	}
	if candidates[0].Desc == "" {
		t.Errorf("binary candidate should carry a file type description")
	}
	for _, candidate := range candidates {
		if candidate.Path == "README.md" {
			t.Errorf("README.md is text and must not be a candidate")
		}
	}
}

func TestDetectUnstageCandidates_ExplicitPathsSkipList(t *testing.T) {
	dir := initGitRepo(t)
	t.Chdir(dir)

	mustRun(t, "mkdir", "-p", filepath.Join(dir, "vendor", "libfoo", "src"))
	os.MkdirAll(filepath.Join(dir, "vendor", "libfoo", ".git"), 0755)
	writeFile(t, "vendor/libfoo/src/main.c", []byte("int main() { return 0; }\n"))
	writeFile(t, "notes.txt", []byte("plain text\n"))

	// Explicit paths are classified without consulting the index, which is how
	// a caller pre-filters a staging batch before knowing the final staged set.
	candidates, err := DetectUnstageCandidates(dir, []string{"vendor/libfoo/src/main.c", "notes.txt"})
	if err != nil {
		t.Fatalf("DetectUnstageCandidates: %v", err)
	}
	want := []string{"submodule vendor/libfoo/src/main.c submodule-dir=vendor/libfoo"}
	if got := candidateSummary(candidates); !reflect.DeepEqual(got, want) {
		t.Fatalf("candidates = %#v, want %#v", got, want)
	}
}

func TestDetectUnstageCandidates_CleanRepo(t *testing.T) {
	dir := initGitRepo(t)
	t.Chdir(dir)

	candidates, err := DetectUnstageCandidates(dir, nil)
	if err != nil {
		t.Fatalf("DetectUnstageCandidates: %v", err)
	}
	if len(candidates) != 0 {
		t.Fatalf("clean repo should have no candidates, got %#v", candidateSummary(candidates))
	}
}

func TestDetectUnstageCandidates_DoesNotMutateIndex(t *testing.T) {
	dir := initGitRepo(t)
	t.Chdir(dir)

	writeFile(t, "icon.png", []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00})
	writeFile(t, "README.md", []byte("hello\n"))
	mustRun(t, "git", "add", "icon.png", "README.md")
	before := getStagedFiles(t)

	if _, err := DetectUnstageCandidates(dir, nil); err != nil {
		t.Fatalf("DetectUnstageCandidates: %v", err)
	}

	after := getStagedFiles(t)
	if !reflect.DeepEqual(after, before) {
		t.Fatalf("detector mutated the index: before %v, after %v", before, after)
	}
}

// TestDetectAndUnstage_MatchesDetector pins the parity the staging side depends
// on: the paths written by detectAndUnstage are exactly the detector's paths.
func TestDetectAndUnstage_MatchesDetector(t *testing.T) {
	dir := initGitRepo(t)
	t.Chdir(dir)

	writeFile(t, "icon.png", []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00})
	writeFile(t, "README.md", []byte("hello\n"))
	mustRun(t, "mkdir", "-p", filepath.Join(dir, "vendor", "libfoo", "src"))
	os.MkdirAll(filepath.Join(dir, "vendor", "libfoo", ".git"), 0755)
	writeFile(t, "vendor/libfoo/src/main.c", []byte("int main() { return 0; }\n"))
	mustRun(t, "git", "add", "icon.png", "README.md", "vendor/libfoo/src/main.c")

	candidates, err := DetectUnstageCandidates(dir, nil)
	if err != nil {
		t.Fatalf("DetectUnstageCandidates: %v", err)
	}
	wantUnstaged := UnstageCandidatePaths(candidates)

	logger := &testLogger{}
	if err := detectAndUnstage(dir, logger); err != nil {
		t.Fatalf("detectAndUnstage: %v", err)
	}

	staged := getStagedFiles(t)
	for _, path := range wantUnstaged {
		if containsFile(staged, path) {
			t.Errorf("%s should have been unstaged", path)
		}
	}
	if !containsFile(staged, "README.md") {
		t.Errorf("README.md is text and must stay staged")
	}
	if !hasLog(logger.logs, "Unstaging binary/submodule entries...") {
		t.Errorf("unstage step should be logged, got %v", logger.logs)
	}

	// Fixed point: once generation has performed its cleanup, asking again must
	// find nothing left to remove, otherwise callers would keep re-unstaging.
	remaining, err := DetectUnstageCandidates(dir, nil)
	if err != nil {
		t.Fatalf("DetectUnstageCandidates after cleanup: %v", err)
	}
	if len(remaining) != 0 {
		t.Fatalf("detector should be empty after cleanup, got %#v", candidateSummary(remaining))
	}
}

func TestUnstageCandidatePaths_PreservesOrder(t *testing.T) {
	candidates := []UnstageCandidate{
		{Path: "b.png", Kind: UnstageKindBinary},
		{Path: "vendor/lib/a.c", Kind: UnstageKindSubmodule, SubmoduleDir: "vendor/lib"},
	}
	want := []string{"b.png", "vendor/lib/a.c"}
	if got := UnstageCandidatePaths(candidates); !reflect.DeepEqual(got, want) {
		t.Fatalf("paths = %v, want %v", got, want)
	}
}

func hasLog(logs []string, want string) bool {
	for _, log := range logs {
		if strings.Contains(log, want) {
			return true
		}
	}
	return false
}
