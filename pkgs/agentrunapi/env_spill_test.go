package agentrunapi

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestShouldSpillEnvThreshold(t *testing.T) {
	t.Parallel()
	short := "A=" + strings.Repeat("x", EnvFileSpillMinRunes-2) // "A=" + 62 = 64 runes
	if utf8.RuneCountInString(short) != EnvFileSpillMinRunes {
		t.Fatalf("fixture: got %d want %d", utf8.RuneCountInString(short), EnvFileSpillMinRunes)
	}
	if ShouldSpillEnv([]string{short}) {
		t.Fatalf("exactly %d runes must not spill", EnvFileSpillMinRunes)
	}
	long := short + "y"
	if !ShouldSpillEnv([]string{long}) {
		t.Fatalf("%d runes must spill", utf8.RuneCountInString(long))
	}
	if ShouldSpillEnv([]string{"  ", ""}) {
		t.Fatal("empty entries must not spill")
	}
	if !ShouldSpillEnv([]string{"SHORT=1", long}) {
		t.Fatal("any long entry must spill the list")
	}
}

func TestMaybeSpillEnvUnderThresholdNoFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path, spilled, err := MaybeSpillEnv([]string{"NO_COLOR=1"}, EnvSpillOpts{SpillDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	if spilled || path != "" {
		t.Fatalf("short env: spilled=%v path=%q", spilled, path)
	}
	ents, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(ents) != 0 {
		t.Fatalf("spill dir should be empty; got %v", ents)
	}
}

func TestMaybeSpillEnvOverThresholdWritesAllEntries(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	long := "PATH=" + strings.Repeat("a", EnvFileSpillMinRunes) // PATH= (5) + 64 = 69
	entries := []string{"NO_COLOR=1", long, "FOO=bar"}
	path, spilled, err := MaybeSpillEnv(entries, EnvSpillOpts{
		SpillDir:  dir,
		SessionID: "sess/env",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !spilled {
		t.Fatal("expected spill")
	}
	if !filepath.IsAbs(path) {
		t.Fatalf("path must be abs; got %q", path)
	}
	rel, err := filepath.Rel(dir, path)
	if err != nil || strings.HasPrefix(rel, "..") {
		t.Fatalf("path %q not under %q", path, dir)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := "NO_COLOR=1\n" + long + "\nFOO=bar\n"
	if string(got) != want {
		t.Fatalf("spill body:\n got %q\nwant %q", got, want)
	}
	base := filepath.Base(path)
	if !strings.HasPrefix(base, "env-sess_env-") {
		t.Fatalf("session id should be sanitized into name; got %q", base)
	}
}

func TestMaybeSpillEnvForceUnderThreshold(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path, spilled, err := MaybeSpillEnv([]string{"A=1"}, EnvSpillOpts{SpillDir: dir, Force: true})
	if err != nil {
		t.Fatal(err)
	}
	if !spilled {
		t.Fatal("Force must write")
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "A=1\n" {
		t.Fatalf("got %q", got)
	}
}

func TestBuildFollowUpCommand_EnvSpillOmitsInlineE(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	long := "PATH=" + strings.Repeat("p", EnvFileSpillMinRunes)
	line, err := BuildFollowUpCommand(FollowUpOpts{
		SessionID:   "sess-env-spill",
		Prompt:      "hi",
		AgentRunner: "grok-tty",
		Open:        true,
		Env:         []string{"NO_COLOR=1", long},
		EnvSpillDir: dir,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(line, "--env-file=") {
		t.Fatalf("want --env-file; got %q", line)
	}
	if strings.Contains(line, "-e") && strings.Contains(line, long) {
		t.Fatalf("must not embed long -e on line; got %q", line)
	}
	if strings.Contains(line, long) {
		t.Fatalf("long PATH must not appear on follow-up line; len=%d", len(line))
	}
	// Short companion must be in the file, not as -e.
	ents, err := os.ReadDir(dir)
	if err != nil || len(ents) == 0 {
		t.Fatalf("expected spill file; ents=%v err=%v", ents, err)
	}
}

func TestBuildFollowUpCommand_ShortEnvStaysInline(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	line, err := BuildFollowUpCommand(FollowUpOpts{
		SessionID:   "sess-env-short",
		Prompt:      "hi",
		AgentRunner: "grok-tty",
		Open:        true,
		Env:         []string{"NO_COLOR=1"},
		EnvSpillDir: dir,
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(line, "--env-file") {
		t.Fatalf("short env must not spill; got %q", line)
	}
	if !strings.Contains(line, "NO_COLOR=1") {
		t.Fatalf("want inline -e NO_COLOR=1; got %q", line)
	}
	ents, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(ents) != 0 {
		t.Fatalf("spill dir must stay empty; got %v", ents)
	}
}

func TestBuildFollowUpCommand_ExplicitEnvFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	given := filepath.Join(dir, "given.env")
	if err := os.WriteFile(given, []byte("A=1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	spill := filepath.Join(dir, "spill")
	if err := os.MkdirAll(spill, 0o755); err != nil {
		t.Fatal(err)
	}
	line, err := BuildFollowUpCommand(FollowUpOpts{
		SessionID:   "sess-env-explicit",
		Prompt:      "hi",
		Open:        true,
		Env:         []string{"PATH=" + strings.Repeat("x", 200)},
		EnvFile:     given,
		EnvSpillDir: spill,
	})
	if err != nil {
		t.Fatal(err)
	}
	abs, _ := filepath.Abs(given)
	if !strings.Contains(line, "--env-file="+abs) && !strings.Contains(line, abs) {
		t.Fatalf("want explicit env-file path; got %q", line)
	}
	ents, err := os.ReadDir(spill)
	if err != nil {
		t.Fatal(err)
	}
	if len(ents) != 0 {
		t.Fatalf("explicit EnvFile must not auto-spill; got %v", ents)
	}
}
