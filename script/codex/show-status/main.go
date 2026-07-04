package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/xhd2015/agent-pro/agent/codex/tty"
)

func main() {
	var verbose bool
	flag.BoolVar(&verbose, "debug", false, "log phase timing and snapshot polls to stderr")
	flag.BoolVar(&verbose, "v", false, "log phase timing and snapshot polls to stderr")
	flag.Parse()

	timeout := 60
	if v := strings.TrimSpace(os.Getenv("CODEX_SHOW_STATUS_TIMEOUT")); v != "" {
		if sec, err := strconv.Atoi(v); err == nil && sec > 0 {
			timeout = sec
		}
	}
	if !verbose {
		v := strings.TrimSpace(os.Getenv("CODEX_SHOW_STATUS_DEBUG"))
		verbose = v == "1" || strings.EqualFold(v, "true") || strings.EqualFold(v, "yes")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	info, err := tty.FetchStatusWithOptions(ctx, tty.Options{Debug: verbose})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Print(tty.FormatStatusLines(info))
}