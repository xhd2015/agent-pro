package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/xhd2015/agent-pro/agent/dsh/consumers"
)

const consumersHelp = `
Usage: agent-pro consumers <command> [ARGS]

Commands:
  migrate   drop cached consumer state so the next read replays it

The cache records belong to agent-pro; a fold whose rules changed is rebuilt by
dropping its state, not by recomputing it somewhere else.

Run agent-pro consumers migrate --help for options.
`

const consumersMigrateHelp = `
Usage: agent-pro consumers migrate --consumer ID (--home DIR | --cache-dir DIR) [--dry-run]
       agent-pro consumers migrate --all (--home DIR | --cache-dir DIR) [--dry-run]

Drops one consumer's cached state from every record in one cache directory, so
the next Sync replays that consumer from the stream origin. Consumers whose state
survived skip the rows they already folded.

Target (exactly one):
  --home DIR       a pinned backend; its records live in DIR/consumers. Defaults
                   to $DSH_HOME when that is set, which is where eluc started
                   with --dsh-data-dir keeps them. Naming --cache-dir overrides
                   this default rather than conflicting with it.
  --cache-dir DIR  a cache directory used directly: the default backend's user
                   cache (agent-pro/session-consumers/v1 under the user cache
                   dir). eluc started WITHOUT --dsh-data-dir keeps its records
                   there, because it passes no home to the bus.

Scope:
  --consumer ID    the fold to drop, e.g. user-rows
  --all            drop every consumer's state in every record; every fold
                   replays. Requires this word explicitly: a bare migrate must
                   not rebuild the whole machine by accident.

Other:
  --dry-run        report what would change without writing
`

func handleConsumers(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(strings.TrimPrefix(consumersHelp, "\n"))
		return nil
	}
	switch args[0] {
	case "migrate":
		return handleConsumersMigrate(args[1:])
	default:
		return fmt.Errorf("unknown consumers command: %s", args[0])
	}
}

func handleConsumersMigrate(args []string) error {
	fs := flag.NewFlagSet("consumers migrate", flag.ContinueOnError)
	fs.Usage = func() { fmt.Print(strings.TrimPrefix(consumersMigrateHelp, "\n")) }
	home := fs.String("home", "", "pinned backend; defaults to $DSH_HOME, records live in <home>/consumers")
	cacheDir := fs.String("cache-dir", "", "cache directory used directly (the default backend's user cache)")
	consumer := fs.String("consumer", "", "consumer id to drop")
	all := fs.Bool("all", false, "drop every consumer's state")
	dryRun := fs.Bool("dry-run", false, "report what would change without writing")
	if err := fs.Parse(args); err != nil {
		// Every level of --help prints usage and succeeds; only a real flag
		// error is an error.
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument: %s", fs.Arg(0))
	}

	dir, err := migrateCacheDir(strings.TrimSpace(*home), strings.TrimSpace(*cacheDir))
	if err != nil {
		return err
	}

	id := strings.TrimSpace(*consumer)
	if *all {
		if id != "" {
			return fmt.Errorf("--consumer and --all are mutually exclusive")
		}
	} else if id == "" {
		return fmt.Errorf("--consumer is required (e.g. --consumer user-rows); pass --all to drop every consumer's state")
	}

	scanned, changed, err := consumers.DropConsumer(dir, id, *dryRun)
	if err != nil {
		return err
	}
	verb := "dropped"
	if *dryRun {
		verb = "would drop"
	}
	fmt.Printf("scanned %d records under %s; %s state for %d record(s)\n", scanned, dir, verb, changed)
	return nil
}

// migrateCacheDir resolves the one directory this run may touch.
//
// Both cache layouts exist, and guessing wrong rebuilds the wrong machine: a
// default backend keeps records under the user cache dir, while a pinned one
// keeps them beside the home it reads. So the caller names the target, and
// $DSH_HOME only supplies the pinned case.
func migrateCacheDir(home, cacheDir string) (string, error) {
	if home != "" && cacheDir != "" {
		return "", fmt.Errorf("choose one target: --home or --cache-dir, not both")
	}
	if cacheDir != "" {
		return cacheDir, nil
	}
	if home == "" {
		// The environment only fills in the pinned target: an explicit
		// --cache-dir must not collide with a home the operator did not name.
		home = strings.TrimSpace(os.Getenv("DSH_HOME"))
	}
	if home == "" {
		return "", fmt.Errorf("choose a target: --home (a pinned backend) or --cache-dir (the default backend's user cache); DSH_HOME is unset")
	}
	return consumers.CacheDirFor(home)
}
