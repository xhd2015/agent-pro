package agentui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"

	"github.com/xhd2015/agent-pro/agent/agentui/sessionstate"
	"github.com/xhd2015/agent-pro/agent/agentui/textutil"
	"github.com/xhd2015/agent-pro/agent/opencode/models"
	"github.com/xhd2015/agent-pro/agent/session"
)

type resolvedRunSession struct {
	SessionID         string
	OpencodeSessionID string
	SessionDir        string
	Feature           string
	Model             string
	InitialLogs       []string
}

func resolveRunSession(cfg Config, opts runOptions) (resolvedRunSession, error) {
	var sessionID string
	var opencodeSessionID string
	var sessionDir string
	var initialLogs []string
	feature := opts.Feature
	llmModel := opts.Model
	resumeID := opts.ResumeID

	sid, osid, dir, feat, resumeModel, logs, err := sessionstate.Resolve(cfg.AgentName, resumeID)
	if err != nil {
		return resolvedRunSession{}, err
	}
	if resumeID != "" {
		sessionID = sid
		opencodeSessionID = osid
		sessionDir = dir
		if feature == "" {
			feature = feat
		}
		if llmModel == "" {
			llmModel = resumeModel
		}
		initialLogs = logs
	} else if feature == "" {
		fmt.Fprint(os.Stderr, cfg.Usage)
		return resolvedRunSession{}, fmt.Errorf("missing feature description")
	}

	if llmModel == "" {
		_, preferred, err := models.ListFree()
		if err != nil {
			return resolvedRunSession{}, fmt.Errorf("failed to list free models: %w", err)
		}
		llmModel = preferred
	}

	if sessionID == "" {
		sessionID = sessionstate.NewID(cfg.SessionPrefix)
	}
	if sessionDir == "" {
		var err error
		sessionDir, err = session.Dir(cfg.AgentName, sessionID)
		if err != nil {
			return resolvedRunSession{}, fmt.Errorf("create session dir: %w", err)
		}
	}
	sessionstate.WriteMeta(sessionDir, sessionstate.Meta{
		SessionID: sessionID,
		Feature:   feature,
		Model:     llmModel,
	})

	return resolvedRunSession{
		SessionID:         sessionID,
		OpencodeSessionID: opencodeSessionID,
		SessionDir:        sessionDir,
		Feature:           feature,
		Model:             llmModel,
		InitialLogs:       initialLogs,
	}, nil
}

type runtimePaths struct {
	TempDir       string
	AnswerDir     string
	QuestionFIFO  string
	QuestionsFile string
}

func prepareRuntimePaths(cfg Config, sessionDir string) (runtimePaths, func(), error) {
	tempDir, err := os.MkdirTemp("", cfg.AgentName+"-*")
	if err != nil {
		return runtimePaths{}, nil, fmt.Errorf("create temp dir: %w", err)
	}
	cleanup := func() { os.RemoveAll(tempDir) }

	exe, err := os.Executable()
	if err != nil {
		return runtimePaths{}, cleanup, fmt.Errorf("get executable path: %w", err)
	}
	apqPath := filepath.Join(tempDir, "add-pending-questions")
	if out, err := exec.Command("cp", exe, apqPath).CombinedOutput(); err != nil {
		return runtimePaths{}, cleanup, fmt.Errorf("failed to copy add-pending-questions: %w\n%s", err, string(out))
	}
	for name := range cfg.Dispatch {
		dst := filepath.Join(tempDir, name)
		if out, err := exec.Command("cp", exe, dst).CombinedOutput(); err != nil {
			return runtimePaths{}, cleanup, fmt.Errorf("failed to copy %s: %w\n%s", name, err, string(out))
		}
	}

	answerDir := filepath.Join(tempDir, "answer")
	if err := os.Mkdir(answerDir, 0755); err != nil {
		return runtimePaths{}, cleanup, fmt.Errorf("create answer dir: %w", err)
	}

	questionFifo := filepath.Join(tempDir, "question.fifo")
	if err := syscall.Mkfifo(questionFifo, 0666); err != nil {
		return runtimePaths{}, cleanup, fmt.Errorf("create question fifo: %w", err)
	}

	questionsFile := filepath.Join(sessionDir, "questions.jsonl")

	return runtimePaths{
		TempDir:       tempDir,
		AnswerDir:     answerDir,
		QuestionFIFO:  questionFifo,
		QuestionsFile: questionsFile,
	}, cleanup, nil
}

func configureQuestionEnvironment(paths runtimePaths) {
	os.Setenv("QUESTION_FIFO", paths.QuestionFIFO)
	os.Setenv("ANSWER_DIR", paths.AnswerDir)
	os.Setenv("QUESTIONS_FILE", paths.QuestionsFile)
	pathEntry := paths.TempDir + string(filepath.ListSeparator)
	if !strings.Contains(os.Getenv("PATH"), pathEntry) {
		os.Setenv("PATH", pathEntry+os.Getenv("PATH"))
	}
}

type runtimeChannels struct {
	LogCh      chan string
	QuestionCh chan pendingQuestion
	DoneCh     chan llmDoneMsg
}

func newRuntimeChannels() runtimeChannels {
	return runtimeChannels{
		LogCh:      make(chan string, 64),
		QuestionCh: make(chan pendingQuestion, 16),
		DoneCh:     make(chan llmDoneMsg, 1),
	}
}

func newRuntimeModel(cfg Config, opts runOptions, resolved resolvedRunSession, paths runtimePaths, channels runtimeChannels) *model {
	ti := textinput.New()
	ti.Placeholder = "Type answer and press Enter..."
	ti.Focus()
	ti.CharLimit = 500

	vp := viewport.New(80, 20)
	initialLogs := resolved.InitialLogs
	feature := resolved.Feature

	startContent := "Starting..."
	if len(initialLogs) > 0 {
		startContent = textutil.WrapText(strings.Join(initialLogs, "\n"), vp.Width)
	} else if feature != "" {
		initialLogs = []string{"[You] " + feature}
		startContent = textutil.WrapText("[You] "+feature, vp.Width)
	}
	vp.SetContent(startContent)
	vp.GotoBottom()

	m := &model{
		feature:           feature,
		llmModel:          resolved.Model,
		agentRunner:       opts.AgentRunner,
		tempDir:           paths.TempDir,
		answerDir:         paths.AnswerDir,
		sessionID:         resolved.SessionID,
		opencodeSessionID: resolved.OpencodeSessionID,
		sessionDir:        resolved.SessionDir,
		questionsFile:     paths.QuestionsFile,
		input:             ti,
		viewport:          vp,
		logs:              initialLogs,
		logCh:             channels.LogCh,
		questionCh:        channels.QuestionCh,
		llmDoneCh:         channels.DoneCh,
	}

	if opts.ResumeID != "" {
		m.loadQuestionsFromFile()
		if m.hasUnansweredQuestions() {
			m.done = true
			m.clarificationMode = true
			m.logs = append(m.logs, "[Agent] Resumed with pending clarification questions.")
		}
	}

	return m
}
