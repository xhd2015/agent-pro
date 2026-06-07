package ask_user

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

func Run() {
	questionFifo := os.Getenv("QUESTION_FIFO")
	answerDir := os.Getenv("ANSWER_DIR")

	if questionFifo == "" || answerDir == "" {
		fmt.Fprintf(os.Stderr, "QUESTION_FIFO and ANSWER_DIR must be set\n")
		os.Exit(1)
	}

	question := strings.Join(os.Args[1:], " ")
	if question == "" {
		fmt.Fprintf(os.Stderr, "usage: ask_user <question>\n")
		os.Exit(1)
	}

	id := nextID(filepath.Join(answerDir, ".counter"))

	answerFifo := filepath.Join(answerDir, id+".fifo")
	if err := syscall.Mkfifo(answerFifo, 0666); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create answer fifo: %v\n", err)
		os.Exit(1)
	}

	f, err := os.OpenFile(questionFifo, os.O_WRONLY, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open question fifo: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(f, "%s\t%s\n", id, question)
	f.Close()

	rf, err := os.Open(answerFifo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open answer fifo: %v\n", err)
		os.Exit(1)
	}
	data, _ := io.ReadAll(rf)
	rf.Close()

	os.Remove(answerFifo)
	fmt.Print(string(data))
}

func nextID(counterFile string) string {
	f, err := os.OpenFile(counterFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "counter file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	syscall.Flock(int(f.Fd()), syscall.LOCK_EX)
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)

	data, _ := io.ReadAll(f)
	n := 1
	if len(data) > 0 {
		fmt.Sscanf(string(data), "%d", &n)
	}
	f.Seek(0, 0)
	f.Truncate(0)
	fmt.Fprintf(f, "%d", n+1)
	return fmt.Sprintf("%d", n)
}
