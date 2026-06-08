package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

var (
	shortMode bool
	runFilter string
)

func main() {
	flag.BoolVar(&shortMode, "short", false, "skip LLM-backed tests")
	flag.StringVar(&runFilter, "run", "", "run only tests matching regex (substring match)")
	flag.Parse()

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

	run("TestAddPendingQuestionsBuild", TestAddPendingQuestionsBuild)
	run("TestAddPendingQuestionsWritesQuestion", TestAddPendingQuestionsWritesQuestion)
	run("TestAddPendingQuestionsMultipleQuestions", TestAddPendingQuestionsMultipleQuestions)
	run("TestAddPendingQuestionsWritesJSONL", TestAddPendingQuestionsWritesJSONL)
	run("TestAddPendingQuestionsNonBlocking", TestAddPendingQuestionsNonBlocking)
	run("TestMainBuild", TestMainBuild)
	run("TestNoArgs", TestNoArgs)
	run("TestRunSelfContained", TestRunSelfContained)
	run("TestRunWithQuestions", TestRunWithQuestions)
	run("TestModelFlag", TestModelFlag)

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
