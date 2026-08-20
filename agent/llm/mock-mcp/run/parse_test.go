package run

import (
	"errors"
	"strings"
	"testing"
	"time"

	lessflags "github.com/xhd2015/less-flags"
)

func TestParse_flagsOnly(t *testing.T) {
	cfg, err := Parse([]string{"--name", "slow_01", "--delay-min", "1s", "--delay-max", "10s"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Name != "slow_01" {
		t.Fatalf("Name=%q", cfg.Name)
	}
	if cfg.DelayMin == nil || *cfg.DelayMin != time.Second {
		t.Fatalf("DelayMin=%v", cfg.DelayMin)
	}
	if cfg.DelayMax == nil || *cfg.DelayMax != 10*time.Second {
		t.Fatalf("DelayMax=%v", cfg.DelayMax)
	}
}

func TestParse_envOnly(t *testing.T) {
	t.Setenv(EnvName, "from-env")
	t.Setenv(EnvDelay, "50ms")
	cfg, err := Parse(nil)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Name != "from-env" {
		t.Fatalf("Name=%q", cfg.Name)
	}
	if cfg.Delay == nil || *cfg.Delay != 50*time.Millisecond {
		t.Fatalf("Delay=%v", cfg.Delay)
	}
}

func TestParse_cliWinsOverEnv(t *testing.T) {
	t.Setenv(EnvName, "env-name")
	t.Setenv(EnvDelayMin, "1s")
	t.Setenv(EnvDelayMax, "10s")
	cfg, err := Parse([]string{"--name", "cli-name", "--delay", "5s"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Name != "cli-name" {
		t.Fatalf("Name=%q", cfg.Name)
	}
	if cfg.Delay == nil || *cfg.Delay != 5*time.Second {
		t.Fatalf("Delay=%v", cfg.Delay)
	}
	if cfg.DelayMin != nil || cfg.DelayMax != nil {
		t.Fatalf("min/max should be cleared by --delay: min=%v max=%v", cfg.DelayMin, cfg.DelayMax)
	}
}

func TestParse_minGreaterThanMax(t *testing.T) {
	_, err := Parse([]string{"--delay-min", "10s", "--delay-max", "1s"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "delay-min") {
		t.Fatalf("err=%v", err)
	}
}

func TestParse_help(t *testing.T) {
	_, err := Parse([]string{"--help"})
	if !errors.Is(err, lessflags.ErrHelp) {
		t.Fatalf("err=%v want ErrHelp", err)
	}
}

func TestParse_hangExclusiveWithDelay(t *testing.T) {
	_, err := Parse([]string{"--hang", "--delay", "1s"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParse_defaultName(t *testing.T) {
	cfg, err := Parse(nil)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Name != defaultName {
		t.Fatalf("Name=%q", cfg.Name)
	}
}
