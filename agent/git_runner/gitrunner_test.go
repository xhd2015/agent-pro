package git_runner

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateSSHConfigRequiresNCWhenProxyEnabled(t *testing.T) {
	prev := isToolAvailable
	isToolAvailable = func(name string) bool { return false }
	defer func() { isToolAvailable = prev }()

	err := validateSSHConfig(&SSHKeyConfig{
		ProxyURL: "http://proxy.example.com:3128",
	})
	if err == nil {
		t.Fatalf("validateSSHConfig() error = nil, want missing nc error")
	}
	if !strings.Contains(err.Error(), "remote-agent exec apt install netcat") {
		t.Fatalf("validateSSHConfig() = %q, want install hint", err.Error())
	}
}

func TestValidateSSHConfigSkipsNCWithoutProxy(t *testing.T) {
	prev := isToolAvailable
	isToolAvailable = func(name string) bool { return false }
	defer func() { isToolAvailable = prev }()

	if err := validateSSHConfig(&SSHKeyConfig{}); err != nil {
		t.Fatalf("validateSSHConfig() error = %v, want nil", err)
	}
}

func TestBuildSSHCommandIncludesKeyAndHTTPProxy(t *testing.T) {
	cmd := buildSSHCommand(&SSHKeyConfig{
		KeyPath:  "/tmp/test key",
		ProxyURL: "http://proxy.example.com:3128",
	})

	for _, want := range []string{
		`"ssh"`,
		`"/tmp/test key"`,
		`"StrictHostKeyChecking=no"`,
		`"UserKnownHostsFile=/dev/null"`,
		`"BatchMode=yes"`,
		`ProxyCommand=`,
		`proxy.example.com:3128`,
		`\"connect\"`,
	} {
		if !strings.Contains(cmd, want) {
			t.Fatalf("buildSSHCommand() missing %q in %q", want, cmd)
		}
	}
}

func TestBuildSSHCommandIncludesSOCKS5ProxyWithoutKey(t *testing.T) {
	cmd := buildSSHCommand(&SSHKeyConfig{
		ProxyURL: "socks5://proxy.example.com:1080",
	})
	if !strings.Contains(cmd, `proxy.example.com:1080`) || !strings.Contains(cmd, `\"5\"`) {
		t.Fatalf("buildSSHCommand() missing socks proxy config in %q", cmd)
	}
	if strings.Contains(cmd, " -i ") {
		t.Fatalf("buildSSHCommand() unexpectedly included -i in %q", cmd)
	}
}

func TestProxyCommandForURLRejectsUnsupportedSchemes(t *testing.T) {
	if got := proxyCommandForURL("https://proxy.example.com:443"); got != "" {
		t.Fatalf("proxyCommandForURL() = %q, want empty for unsupported scheme", got)
	}
}

func TestBuildPreservesCustomAskPass(t *testing.T) {
	cmd := NewCommand("fetch").WithEnv("GIT_ASKPASS", "/tmp/git-askpass").Build()

	got := envValues(cmd.Env, "GIT_ASKPASS")
	if len(got) != 1 || got[0] != "/tmp/git-askpass" {
		t.Fatalf("GIT_ASKPASS values = %v, want only custom helper", got)
	}
}

func TestRemoveStaleIndexLock_AllowsCommit(t *testing.T) {
	dir := t.TempDir()
	runGit(t, dir, "init", "--template=")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "Test User")
	runGit(t, dir, "config", "core.hooksPath", "/dev/null")
	if err := os.WriteFile(filepath.Join(dir, "tracked.txt"), []byte("hello\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", "tracked.txt")
	runGit(t, dir, "commit", "-m", "initial")

	if err := os.WriteFile(filepath.Join(dir, "next.txt"), []byte("world\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", "next.txt")

	lockPath, err := IndexLockPath(dir)
	if err != nil {
		t.Fatalf("IndexLockPath: %v", err)
	}
	if err := os.WriteFile(lockPath, []byte("stale"), 0644); err != nil {
		t.Fatalf("create stale lock: %v", err)
	}

	if err := RemoveStaleIndexLock(dir); err != nil {
		t.Fatalf("RemoveStaleIndexLock: %v", err)
	}
	if _, err := os.Stat(lockPath); !os.IsNotExist(err) {
		t.Fatalf("stale lock still present after RemoveStaleIndexLock: %v", err)
	}

	output, err := CommitWithRetry(dir, "feat: after stale lock", 1)
	if err != nil {
		t.Fatalf("CommitWithRetry failed: %s: %v", string(output), err)
	}
	subject, err := NewCommand("log", "-1", "--format=%s").Dir(dir).Output()
	if err != nil {
		t.Fatalf("git log: %v", err)
	}
	if got := strings.TrimSpace(string(subject)); got != "feat: after stale lock" {
		t.Fatalf("commit subject = %q, want %q", got, "feat: after stale lock")
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s: %s: %v", strings.Join(args, " "), string(out), err)
	}
}

func envValues(env []string, key string) []string {
	prefix := key + "="
	var values []string
	for _, entry := range env {
		if strings.HasPrefix(entry, prefix) {
			values = append(values, strings.TrimPrefix(entry, prefix))
		}
	}
	return values
}
