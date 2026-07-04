package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/xhd2015/agent-pro/agent/grok/tty"
)

func main() {
	var verbose bool
	flag.BoolVar(&verbose, "debug", false, "log grok command and state transitions to stderr")
	flag.BoolVar(&verbose, "v", false, "log grok command and state transitions to stderr")
	flag.Parse()

	timeout := 60
	if v := strings.TrimSpace(os.Getenv("GROK_SHOW_USAGE_TIMEOUT")); v != "" {
		if sec, err := strconv.Atoi(v); err == nil && sec > 0 {
			timeout = sec
		}
	}
	if !verbose {
		v := strings.TrimSpace(os.Getenv("GROK_SHOW_USAGE_DEBUG"))
		verbose = v == "1" || strings.EqualFold(v, "true") || strings.EqualFold(v, "yes")
	}

	maxAttempts := 0
	if v := strings.TrimSpace(os.Getenv("GROK_SHOW_USAGE_RETRIES")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxAttempts = n
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	info, err := tty.FetchUsageWithOptions(ctx, tty.Options{
		Debug:       verbose,
		MaxAttempts: maxAttempts,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Print(tty.FormatUsageLines(info))
}