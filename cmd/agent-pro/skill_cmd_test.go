package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDebugWithUserSkillIsRegistered(t *testing.T) {
	names := knownSkillNames()
	if !containsString(names, "debug-with-user") {
		t.Fatalf("knownSkillNames missing debug-with-user: %v", names)
	}
	sk, ok := knownSkills["debug-with-user"]
	if !ok {
		t.Fatal("knownSkills missing debug-with-user")
	}
	if !strings.Contains(sk.Content, "name: debug-with-user") {
		t.Fatalf("skill content missing frontmatter name:\n%s", sk.Content)
	}
}

func TestVerifyWithPrototypeSkillIsRegistered(t *testing.T) {
	names := knownSkillNames()
	if !containsString(names, "verify-with-prototype") {
		t.Fatalf("knownSkillNames missing verify-with-prototype: %v", names)
	}
	sk, ok := knownSkills["verify-with-prototype"]
	if !ok {
		t.Fatal("knownSkills missing verify-with-prototype")
	}
	if !strings.Contains(sk.Content, "name: verify-with-prototype") {
		t.Fatalf("skill content missing frontmatter name:\n%s", sk.Content)
	}
	if sk.Description == "" {
		t.Fatal("verify-with-prototype skill missing description")
	}
}

func TestInvestigateSkillIsRegistered(t *testing.T) {
	names := knownSkillNames()
	if !containsString(names, "investigate") {
		t.Fatalf("knownSkillNames missing investigate: %v", names)
	}
	sk, ok := knownSkills["investigate"]
	if !ok {
		t.Fatal("knownSkills missing investigate")
	}
	if !strings.Contains(sk.Content, "name: investigate") {
		t.Fatalf("skill content missing frontmatter name:\n%s", sk.Content)
	}
	if sk.Description == "" {
		t.Fatal("investigate skill missing description")
	}
}

func TestGitResolveConflictsSkillIsRegistered(t *testing.T) {
	names := knownSkillNames()
	if !containsString(names, "git-resolve-conflicts") {
		t.Fatalf("knownSkillNames missing git-resolve-conflicts: %v", names)
	}
	sk, ok := knownSkills["git-resolve-conflicts"]
	if !ok {
		t.Fatal("knownSkills missing git-resolve-conflicts")
	}
	if !strings.Contains(sk.Content, "git rebase --continue") {
		t.Fatalf("skill content does not describe rebase follow-up:\n%s", sk.Content)
	}
	if !strings.Contains(sk.Content, "Do not run") || !strings.Contains(sk.Content, "git add") {
		t.Fatalf("skill content does not forbid staging/continuing:\n%s", sk.Content)
	}
}

func TestHandleSkillsHelpMentionsUpdate(t *testing.T) {
	stdout := captureStdout(t, func() {
		if err := handleSkills([]string{"--help"}); err != nil {
			t.Fatalf("handleSkills(--help): %v", err)
		}
	})
	if !strings.Contains(stdout, "update") {
		t.Fatalf("skills help missing update:\n%s", stdout)
	}
}

func TestHandleSkillShowGitResolveConflicts(t *testing.T) {
	stdout := captureStdout(t, func() {
		if err := handleSkill([]string{"git-resolve-conflicts", "show"}); err != nil {
			t.Fatalf("handleSkill(show): %v", err)
		}
	})
	if !strings.Contains(stdout, "name: git-resolve-conflicts") {
		t.Fatalf("show output missing frontmatter name:\n%s", stdout)
	}
	if !strings.Contains(stdout, "git merge --continue") {
		t.Fatalf("show output missing merge follow-up:\n%s", stdout)
	}
}

func TestHandleSkillsUpdateUpdatesAlreadyInstalledSkill(t *testing.T) {
	tmp := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir temp: %v", err)
	}
	defer func() {
		if err := os.Chdir(oldWD); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	}()

	skillDir := filepath.Join(tmp, ".agents", "skills", "git-resolve-conflicts")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatalf("mkdir skill dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("old skill\n"), 0644); err != nil {
		t.Fatalf("write old skill: %v", err)
	}

	stdout := captureStdout(t, func() {
		if err := handleSkills([]string{"update"}); err != nil {
			t.Fatalf("handleSkills(update): %v", err)
		}
	})
	if !strings.Contains(stdout, "Update skill at") {
		t.Fatalf("update output missing update line:\n%s", stdout)
	}
	if strings.Contains(stdout, "Installed skill to") {
		t.Fatalf("update should not install missing skills:\n%s", stdout)
	}

	content, err := os.ReadFile(filepath.Join(skillDir, "SKILL.md"))
	if err != nil {
		t.Fatalf("read updated skill: %v", err)
	}
	if !strings.Contains(string(content), "name: git-resolve-conflicts") {
		t.Fatalf("installed skill was not updated:\n%s", content)
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	defer func() {
		os.Stdout = old
	}()

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("close stdout writer: %v", err)
	}
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Fatalf("close stdout reader: %v", err)
	}
	return string(out)
}
