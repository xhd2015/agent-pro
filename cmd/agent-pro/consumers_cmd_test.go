package main

import (
	"strings"
	"testing"
)

// Every level of --help prints usage and exits successfully; a command that
// answers -h with an error breaks scripts and the CLI's own conventions.
func TestConsumersMigrateHelpSucceeds(t *testing.T) {
	for _, args := range [][]string{
		{"migrate", "--help"},
		{"migrate", "-h"},
	} {
		if err := handleConsumers(args); err != nil {
			t.Fatalf("%v: %v", args, err)
		}
	}
}

// The two cache layouts are not interchangeable: a pinned backend keeps records
// beside its home, the default backend in the user cache. A run that names
// neither must stop instead of emptying the wrong one.
func TestConsumersMigrateRequiresOneTarget(t *testing.T) {
	t.Setenv("DSH_HOME", "")
	if err := handleConsumersMigrate([]string{"--consumer", "permission"}); err == nil {
		t.Fatal("a run with no target was accepted")
	}

	// $DSH_HOME supplies the pinned case, so its layout becomes the default.
	t.Setenv("DSH_HOME", t.TempDir())
	if err := handleConsumersMigrate([]string{"--consumer", "permission", "--dry-run"}); err != nil {
		t.Fatalf("pinned default: %v", err)
	}
}

// A bare migrate must not rebuild every fold on the machine: the scope is named
// either as one consumer or as the explicit --all.
func TestConsumersMigrateRequiresAnExplicitScope(t *testing.T) {
	home := t.TempDir()
	err := handleConsumersMigrate([]string{"--home", home})
	if err == nil || !strings.Contains(err.Error(), "--consumer") {
		t.Fatalf("a bare migrate was accepted: %v", err)
	}
	if err := handleConsumersMigrate([]string{"--home", home, "--all", "--consumer", "permission"}); err == nil {
		t.Fatal("--consumer with --all was accepted")
	}
	if err := handleConsumersMigrate([]string{"--home", home, "--all", "--dry-run"}); err != nil {
		t.Fatalf("--all: %v", err)
	}
}

func TestConsumersMigrateRejectsTwoTargets(t *testing.T) {
	err := handleConsumersMigrate([]string{
		"--home", t.TempDir(), "--cache-dir", t.TempDir(), "--consumer", "permission",
	})
	if err == nil {
		t.Fatal("two targets were accepted")
	}
}
