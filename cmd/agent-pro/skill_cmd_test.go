package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/skills/skillcmd"
)

func TestDebugWithUserSkillIsRegistered(t *testing.T) {
	names := agentProSkillNames()
	if !containsString(names, "debug-with-user") {
		t.Fatalf("agentProSkillNames missing debug-with-user: %v", names)
	}
	sk, ok := agentProSkill("debug-with-user")
	if !ok {
		t.Fatal("agentProSkill missing debug-with-user")
	}
	if !strings.Contains(sk.Content, "name: debug-with-user") {
		t.Fatalf("skill content missing frontmatter name:\n%s", sk.Content)
	}
}

func TestConsolidateCodeSkillIsRegistered(t *testing.T) {
	names := agentProSkillNames()
	if !containsString(names, "consolidate-code") {
		t.Fatalf("agentProSkillNames missing consolidate-code: %v", names)
	}
	sk, ok := agentProSkill("consolidate-code")
	if !ok {
		t.Fatal("agentProSkill missing consolidate-code")
	}
	if !strings.Contains(sk.Content, "name: consolidate-code") {
		t.Fatalf("skill content missing frontmatter name:\n%s", sk.Content)
	}
	if skillcmd.ExtractDescription(sk.Content) == "" {
		t.Fatal("consolidate-code skill missing description")
	}
	if !strings.Contains(sk.Content, "doctest-tdd") {
		t.Fatalf("skill content missing doctest-tdd pairing:\n%s", sk.Content)
	}
}

func TestVerifyWithPrototypeSkillIsRegistered(t *testing.T) {
	names := agentProSkillNames()
	if !containsString(names, "verify-with-prototype") {
		t.Fatalf("agentProSkillNames missing verify-with-prototype: %v", names)
	}
	sk, ok := agentProSkill("verify-with-prototype")
	if !ok {
		t.Fatal("agentProSkill missing verify-with-prototype")
	}
	if !strings.Contains(sk.Content, "name: verify-with-prototype") {
		t.Fatalf("skill content missing frontmatter name:\n%s", sk.Content)
	}
	if skillcmd.ExtractDescription(sk.Content) == "" {
		t.Fatal("verify-with-prototype skill missing description")
	}
}

func TestVerifyOnBehalfOfUserSkillIsRegistered(t *testing.T) {
	names := agentProSkillNames()
	if !containsString(names, "verify-on-behalf-of-user") {
		t.Fatalf("agentProSkillNames missing verify-on-behalf-of-user: %v", names)
	}
	sk, ok := agentProSkill("verify-on-behalf-of-user")
	if !ok {
		t.Fatal("agentProSkill missing verify-on-behalf-of-user")
	}
	if !strings.Contains(sk.Content, "name: verify-on-behalf-of-user") {
		t.Fatalf("skill content missing frontmatter name:\n%s", sk.Content)
	}
	if skillcmd.ExtractDescription(sk.Content) == "" {
		t.Fatal("verify-on-behalf-of-user skill missing description")
	}
	if !strings.Contains(sk.Content, "## Topics") {
		t.Fatal("verify-on-behalf-of-user skill missing topic index")
	}
	if sk.TreeFS == nil {
		t.Fatal("verify-on-behalf-of-user skill missing TreeFS")
	}
	if len(sk.ExtraFiles) == 0 {
		t.Fatal("verify-on-behalf-of-user skill missing ExtraFiles for utility scripts")
	}
	var hasEnterSandbox, hasTranscriptTopic bool
	for _, f := range sk.ExtraFiles {
		if f.Path == "scripts/enter-sandbox.sh" && len(f.Content) > 0 {
			hasEnterSandbox = true
		}
		if f.Path == "transcript/TOPIC.md" && strings.Contains(string(f.Content), "Transcript format rules") {
			hasTranscriptTopic = true
		}
	}
	if !hasEnterSandbox {
		t.Fatal("verify-on-behalf-of-user ExtraFiles missing scripts/enter-sandbox.sh")
	}
	if !hasTranscriptTopic {
		t.Fatal("verify-on-behalf-of-user ExtraFiles missing transcript/TOPIC.md")
	}
}

func TestInvestigateSkillIsRegistered(t *testing.T) {
	names := agentProSkillNames()
	if !containsString(names, "investigate") {
		t.Fatalf("agentProSkillNames missing investigate: %v", names)
	}
	sk, ok := agentProSkill("investigate")
	if !ok {
		t.Fatal("agentProSkill missing investigate")
	}
	if !strings.Contains(sk.Content, "name: investigate") {
		t.Fatalf("skill content missing frontmatter name:\n%s", sk.Content)
	}
	if skillcmd.ExtractDescription(sk.Content) == "" {
		t.Fatal("investigate skill missing description")
	}
}

func TestBrainstormSkillIsRegistered(t *testing.T) {
	names := agentProSkillNames()
	if !containsString(names, "brainstorm") {
		t.Fatalf("agentProSkillNames missing brainstorm: %v", names)
	}
	sk, ok := agentProSkill("brainstorm")
	if !ok {
		t.Fatal("agentProSkill missing brainstorm")
	}
	if !strings.Contains(sk.Content, "name: brainstorm") {
		t.Fatalf("skill content missing frontmatter name:\n%s", sk.Content)
	}
	if skillcmd.ExtractDescription(sk.Content) == "" {
		t.Fatal("brainstorm skill missing description")
	}
	if !strings.Contains(sk.Content, "CLI examples") {
		t.Fatal("brainstorm skill missing CLI output section")
	}
	if !strings.Contains(sk.Content, "UI mockups") {
		t.Fatal("brainstorm skill missing UI mockups section")
	}
}

func TestFollowupSkillIsRegistered(t *testing.T) {
	names := agentProSkillNames()
	if !containsString(names, "followup") {
		t.Fatalf("agentProSkillNames missing followup: %v", names)
	}
	sk, ok := agentProSkill("followup")
	if !ok {
		t.Fatal("agentProSkill missing followup")
	}
	if !strings.Contains(sk.Content, "name: followup") {
		t.Fatalf("skill content missing frontmatter name:\n%s", sk.Content)
	}
	if skillcmd.ExtractDescription(sk.Content) == "" {
		t.Fatal("followup skill missing description")
	}
	if !strings.Contains(sk.Content, "clarification phase") {
		t.Fatal("followup skill missing clarification phase instruction")
	}
}

func TestEstablishALoopSkillIsRegistered(t *testing.T) {
	names := agentProSkillNames()
	if !containsString(names, "establish-a-loop") {
		t.Fatalf("agentProSkillNames missing establish-a-loop: %v", names)
	}
	sk, ok := agentProSkill("establish-a-loop")
	if !ok {
		t.Fatal("agentProSkill missing establish-a-loop")
	}
	if !strings.Contains(sk.Content, "name: establish-a-loop") {
		t.Fatalf("skill content missing frontmatter name:\n%s", sk.Content)
	}
	if skillcmd.ExtractDescription(sk.Content) == "" {
		t.Fatal("establish-a-loop skill missing description")
	}
	if !strings.Contains(sk.Content, "LOOP_<YYYY-MM-DD>_<slug>.md") {
		t.Fatalf("skill content missing LOOP filename convention:\n%s", sk.Content)
	}
	for _, needle := range []string{
		"bug-repro",
		"SYMPTOM CONFIRMED",
		"run-the-loop",
		"Derived operations",
	} {
		if !strings.Contains(sk.Content, needle) {
			t.Fatalf("establish-a-loop skill missing %q:\n%s", needle, sk.Content)
		}
	}
}

func TestRunTheLoopSkillIsRegistered(t *testing.T) {
	names := agentProSkillNames()
	if !containsString(names, "run-the-loop") {
		t.Fatalf("agentProSkillNames missing run-the-loop: %v", names)
	}
	sk, ok := agentProSkill("run-the-loop")
	if !ok {
		t.Fatal("agentProSkill missing run-the-loop")
	}
	if !strings.Contains(sk.Content, "name: run-the-loop") {
		t.Fatalf("skill content missing frontmatter name:\n%s", sk.Content)
	}
	if skillcmd.ExtractDescription(sk.Content) == "" {
		t.Fatal("run-the-loop skill missing description")
	}
}

func TestSoundFixSkillIsRegistered(t *testing.T) {
	names := agentProSkillNames()
	if !containsString(names, "sound-fix") {
		t.Fatalf("agentProSkillNames missing sound-fix: %v", names)
	}
	sk, ok := agentProSkill("sound-fix")
	if !ok {
		t.Fatal("agentProSkill missing sound-fix")
	}
	if !strings.Contains(sk.Content, "name: sound-fix") {
		t.Fatalf("skill content missing frontmatter name:\n%s", sk.Content)
	}
	if skillcmd.ExtractDescription(sk.Content) == "" {
		t.Fatal("sound-fix skill missing description")
	}
}

func TestSplitPhasesSkillIsRegistered(t *testing.T) {
	names := agentProSkillNames()
	if !containsString(names, "split-phases") {
		t.Fatalf("agentProSkillNames missing split-phases: %v", names)
	}
	sk, ok := agentProSkill("split-phases")
	if !ok {
		t.Fatal("agentProSkill missing split-phases")
	}
	if !strings.Contains(sk.Content, "name: split-phases") {
		t.Fatalf("skill content missing frontmatter name:\n%s", sk.Content)
	}
	if skillcmd.ExtractDescription(sk.Content) == "" {
		t.Fatal("split-phases skill missing description")
	}
	if !strings.Contains(sk.Content, "independently implementable") {
		t.Fatal("split-phases skill missing independently implementable instruction")
	}
	if !strings.Contains(sk.Content, "dependency-ordered") && !strings.Contains(sk.Content, "Dependency-first") {
		t.Fatal("split-phases skill missing dependency ordering guidance")
	}
}

func TestGitResolveConflictsSkillIsRegistered(t *testing.T) {
	names := agentProSkillNames()
	if !containsString(names, "git-resolve-conflicts") {
		t.Fatalf("agentProSkillNames missing git-resolve-conflicts: %v", names)
	}
	sk, ok := agentProSkill("git-resolve-conflicts")
	if !ok {
		t.Fatal("agentProSkill missing git-resolve-conflicts")
	}
	if !strings.Contains(sk.Content, "git rebase --continue") {
		t.Fatalf("skill content does not describe rebase follow-up:\n%s", sk.Content)
	}
	if !strings.Contains(sk.Content, "Do not run") || !strings.Contains(sk.Content, "git add") {
		t.Fatalf("skill content does not forbid staging/continuing:\n%s", sk.Content)
	}
}

func TestSummarizeASkillIsRegistered(t *testing.T) {
	names := agentProSkillNames()
	if !containsString(names, "summarize-a-skill") {
		t.Fatalf("agentProSkillNames missing summarize-a-skill: %v", names)
	}
	sk, ok := agentProSkill("summarize-a-skill")
	if !ok {
		t.Fatal("agentProSkill missing summarize-a-skill")
	}
	if !strings.Contains(sk.Content, "name: summarize-a-skill") {
		t.Fatalf("skill content missing frontmatter name:\n%s", sk.Content)
	}
	if skillcmd.ExtractDescription(sk.Content) == "" {
		t.Fatal("summarize-a-skill skill missing description")
	}
	if !strings.Contains(sk.Content, "Output path resolution") {
		t.Fatal("summarize-a-skill skill missing output path resolution section")
	}
	if !strings.Contains(sk.Content, "What works") {
		t.Fatal("summarize-a-skill skill missing What works section")
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
		if err := handleSkill([]string{"git-resolve-conflicts", "--show"}); err != nil {
			t.Fatalf("handleSkill(--show): %v", err)
		}
	})
	if !strings.Contains(stdout, "name: git-resolve-conflicts") {
		t.Fatalf("show output missing frontmatter name:\n%s", stdout)
	}
	if !strings.Contains(stdout, "git merge --continue") {
		t.Fatalf("show output missing merge follow-up:\n%s", stdout)
	}
}

func TestHandleSkillShowFlagBeforeName(t *testing.T) {
	stdout := captureStdout(t, func() {
		if err := handleSkill([]string{"--show", "summarize-a-skill"}); err != nil {
			t.Fatalf("handleSkill(--show name): %v", err)
		}
	})
	if !strings.Contains(stdout, "name: summarize-a-skill") {
		t.Fatalf("show output missing frontmatter name:\n%s", stdout)
	}
}

func TestHandleSkillsInstallFlagOrder(t *testing.T) {
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

	// skills <name> --install (the form that previously failed with --install as command)
	stdout := captureStdout(t, func() {
		if err := handleSkills([]string{"summarize-a-skill", "--install", "--dry-run"}); err != nil {
			t.Fatalf("handleSkills(name --install): %v", err)
		}
	})
	if !strings.Contains(stdout, "summarize-a-skill") {
		t.Fatalf("install dry-run missing skill name:\n%s", stdout)
	}

	// skill --install <name>
	stdout2 := captureStdout(t, func() {
		if err := handleSkill([]string{"--install", "summarize-a-skill", "--dry-run"}); err != nil {
			t.Fatalf("handleSkill(--install name): %v", err)
		}
	})
	if !strings.Contains(stdout2, "summarize-a-skill") {
		t.Fatalf("install dry-run missing skill name:\n%s", stdout2)
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
	// HandleUpdateMany prints a per-skill table + summary, not the older
	// single-skill "Update skill at <dir>" / "Installed skill to" lines.
	if !strings.Contains(stdout, "git-resolve-conflicts") || !strings.Contains(stdout, "updated") {
		t.Fatalf("update output missing updated skill row:\n%s", stdout)
	}
	if !strings.Contains(stdout, "SKILL.md") {
		t.Fatalf("update output missing SKILL.md path:\n%s", stdout)
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

func agentProSkillNames() []string {
	skills := agentProSkills()
	names := make([]string, 0, len(skills))
	for _, sk := range skills {
		names = append(names, sk.Name)
	}
	return names
}

func agentProSkill(name string) (skillcmd.RegisteredSkill, bool) {
	for _, sk := range agentProSkills() {
		if sk.Name == name {
			return sk, true
		}
	}
	return skillcmd.RegisteredSkill{}, false
}

func TestHandleSkillShowTopicBothForms(t *testing.T) {
	for _, args := range [][]string{
		{"--show", "verify-on-behalf-of-user/transcript"},
		{"--show", "verify-on-behalf-of-user", "transcript"},
	} {
		stdout := captureStdout(t, func() {
			if err := handleSkill(args); err != nil {
				t.Fatalf("handleSkill(%v): %v", args, err)
			}
		})
		if !strings.Contains(stdout, "Transcript format rules") {
			t.Fatalf("handleSkill(%v) missing topic content:\n%s", args, stdout)
		}
	}
}

func TestHandleSkillListShowsEverySkill(t *testing.T) {
	stdout := captureStdout(t, func() {
		if err := handleSkill([]string{"--list"}); err != nil {
			t.Fatalf("--list: %v", err)
		}
	})
	for _, name := range agentProSkillNames() {
		if !strings.Contains(stdout, name) {
			t.Fatalf("--list missing %q:\n%s", name, stdout)
		}
	}
}
