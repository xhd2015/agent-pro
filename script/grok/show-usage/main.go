package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	var verbose bool
	flag.BoolVar(&verbose, "debug", false, "log grok command and state transitions to stderr")
	flag.BoolVar(&verbose, "v", false, "log grok command and state transitions to stderr")
	flag.Parse()

	timeout := defaultTimeoutSeconds
	if v := strings.TrimSpace(os.Getenv(envShowUsageTimeout)); v != "" {
		if sec, err := strconv.Atoi(v); err == nil && sec > 0 {
			timeout = sec
		}
	}
	if !verbose {
		verbose = debugEnabled()
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	info, err := FetchUsageWithOptions(ctx, Options{
		Debug:       verbose,
		MaxAttempts: maxAttemptsFromEnv(),
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("Weekly limit: %s\nNext reset: %s\n", info.WeeklyLimit, info.NextReset)
}