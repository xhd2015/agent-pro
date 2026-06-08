package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/xhd2015/less-flags"
)

const help = `Usage: go run ./test/integrations/ [options]

Options:
  -short           skip LLM-backed tests
  -run FILTER      run only tests matching substring
  -h, --help       show this help
`

var (
	shortMode bool
	runFilter string
)

func main() {
	_, err := lessflags.Bool("-short", &shortMode).
		String("-run", &runFilter).
		Help("-h,--help", help).
		Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	start := time.Now()
	passed := 0
	failed := 0
	skipped := 0

	run := func(name string, fn func(t *T)) {
		if runFilter != "" && !strings.Contains(name, runFilter) {
			return
		}
		t := &T{name: name, start: time.Now()}
		fmt.Printf("=== RUN   %s\n", name)
		func() {
			defer func() {
				if r := recover(); r != nil {
					if _, ok := r.(*fatalPanic); !ok {
						panic(r)
					}
					t.failed = true
				}
			}()
			fn(t)
		}()
		t.elapsed = time.Since(t.start)

		switch {
		case t.skipped:
			skipped++
			fmt.Printf("--- SKIP: %s (%s)\n", name, t.elapsed)
		case t.failed:
			failed++
			fmt.Printf("--- FAIL: %s (%s)\n", name, t.elapsed)
		default:
			passed++
			fmt.Printf("--- PASS: %s (%s)\n", name, t.elapsed)
		}
	}

	run("TestAgentRootResolves", TestAgentRootResolves)
	run("TestBuild", TestBuild)
	run("TestNoArgs", TestNoArgs)
	run("TestHelpFlag", TestHelpFlag)
	run("TestWriteCLI", TestWriteCLI)
	run("TestWriteLibrary", TestWriteLibrary)

	elapsed := time.Since(start)
	fmt.Println()
	if failed > 0 {
		fmt.Printf("FAIL\n")
	} else {
		fmt.Printf("PASS\n")
	}
	fmt.Printf("ok  \ttest/integrations\t%.3fs\n", elapsed.Seconds())

	if failed > 0 {
		os.Exit(1)
	}
}
