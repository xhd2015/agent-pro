package web

import (
	"context"
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// Service owns DSH Web session lifecycle. It never starts a terminal, agent-run,
// dsh CLI, or readiness probe. Client transport and Store are independently
// injectable; separate services using the same Store share durable checkpoints.
type Service struct {
	Client Client
	Store  Store
}

// OpenSession creates/reuses a caller-keyed DSH conversation, admits its first
// input once, and opens the selected browser unless NoOpen or BrowserNone. Repeated calls do
// not resubmit an admitted initial prompt. Use Send for a subsequent message.
func (s Service) OpenSession(ctx context.Context, opts SessionOptions) (Result, error) {
	return s.deliver(ctx, opts, true)
}

// Send admits a follow-up to an existing binding and never opens a browser.
// An interrupted initial input is settled first with its original request ID.
func (s Service) Send(ctx context.Context, opts SendOptions) (Result, error) {
	if strings.TrimSpace(opts.Prompt) == "" {
		return Result{}, fmt.Errorf("DSH prompt is required")
	}
	return s.deliver(ctx, SessionOptions{Key: opts.Key, Prompt: opts.Prompt, NoSubmit: opts.NoSubmit, NoOpen: true}, false)
}

func (s Service) deliver(ctx context.Context, opts SessionOptions, initial bool) (result Result, err error) {
	if s.Store == nil {
		return result, fmt.Errorf("DSH session store is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	err = s.Store.WithSession(ctx, opts.Key, func(record Record) error {
		state, found, err := record.Load()
		if err != nil {
			return err
		}
		if !found {
			if !initial {
				return fmt.Errorf("DSH session binding is missing; open the conversation first")
			}
			browser, err := NormalizeBrowser(opts.Browser)
			if err != nil {
				return err
			}
			workspace, err := filepath.Abs(opts.Workspace)
			if err != nil {
				return err
			}
			selected, err := normalizeModelSelection(opts.ModelSelection)
			if err != nil {
				return err
			}
			state = State{Runner: Runner, Browser: browser, Workspace: workspace, DSHSessionID: reserveDSHSessionID(), InitialModelSelection: selected}
			if err := record.Save(state); err != nil {
				return err
			}
		}
		if err := validateState(state); err != nil {
			return err
		}
		browser, err := NormalizeBrowser(state.Browser)
		if err != nil {
			return fmt.Errorf("invalid DSH binding: %w", err)
		}
		// Legacy not-yet-created records may predate reservation of provider IDs.
		if state.DSHSessionID == "" {
			state.DSHSessionID = reserveDSHSessionID()
			if err := record.Save(state); err != nil {
				return err
			}
		}
		result.SessionID = state.DSHSessionID
		noSubmit := opts.NoSubmit
		hash := promptHash(opts.Prompt, noSubmit)
		shouldDeliver := strings.TrimSpace(opts.Prompt) != "" && (!initial || !state.InitialDelivered)
		if shouldDeliver && state.PendingRequestID == "" {
			setPending(&state, opts.Prompt, noSubmit, hash)
			if err := record.Save(state); err != nil {
				return err
			}
		}
		if !state.Created {
			created, err := s.Client.CreateWithOptions(ctx, CreateOptions{Workspace: state.Workspace, SessionID: state.DSHSessionID, ModelSelection: state.InitialModelSelection})
			if err != nil {
				return err
			}
			state.DSHSessionID, state.Created = created.SessionID, true
			state.InitialModelSelection = created.ModelSelection
			result.SessionID, result.Created = created.SessionID, true
			if err := record.Save(state); err != nil {
				return err
			}
		}
		pendingWasCurrent := state.PendingHash == hash
		if state.PendingRequestID != "" {
			if err := s.flush(ctx, record, &state); err != nil {
				return err
			}
		}
		if shouldDeliver && !pendingWasCurrent {
			setPending(&state, opts.Prompt, noSubmit, hash)
			if err := record.Save(state); err != nil {
				return err
			}
			if err := s.flush(ctx, record, &state); err != nil {
				return err
			}
		}
		result.Accepted = state.InitialDelivered
		if !opts.NoOpen && browser != BrowserNone && !state.BrowserOpened {
			if err := s.Client.Open(state.DSHSessionID, browser); err != nil {
				result.Warnings = append(result.Warnings, Warning{Code: "browser_open_failed", Message: err.Error()})
				return nil // admission already committed; don't request redelivery
			}
			result.BrowserOpened = true
			state.BrowserOpened = true
			if err := record.Save(state); err != nil {
				result.Warnings = append(result.Warnings, Warning{Code: "browser_checkpoint_failed", Message: "could not record DSH browser opening; input admission is already confirmed"})
			}
		}
		return nil
	})
	return result, err
}

// reserveDSHSessionID mints the same Host shape as DSH Web New Session
// (createMinted): session-<uuid>.
func reserveDSHSessionID() string {
	return "session-" + uuid.NewString()
}

func setPending(state *State, prompt string, noSubmit bool, hash string) {
	state.PendingRequestID, state.PendingHash = uuid.NewString(), hash
	state.PendingPrompt, state.PendingNoSubmit = prompt, noSubmit
}

func promptHash(prompt string, noSubmit bool) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%t\x00%s", noSubmit, prompt))))
}

func validateState(state State) error {
	if _, err := normalizeModelSelection(state.InitialModelSelection); err != nil {
		return fmt.Errorf("invalid DSH binding: %w", err)
	}
	if !IsRunner(state.Runner) {
		return fmt.Errorf("session binding belongs to another runner, not DSH Web")
	}
	if state.Created && state.DSHSessionID == "" {
		return fmt.Errorf("invalid DSH binding: created session has no provider ID")
	}
	if state.PendingRequestID != "" && strings.TrimSpace(state.PendingPrompt) == "" {
		return fmt.Errorf("invalid DSH binding: pending request has no prompt")
	}
	return nil
}

func (s Service) flush(ctx context.Context, record Record, state *State) error {
	if err := s.Client.Send(ctx, state.DSHSessionID, state.PendingRequestID, state.PendingPrompt, state.PendingNoSubmit); err != nil {
		return err
	}
	state.PendingRequestID, state.PendingHash, state.PendingPrompt = "", "", ""
	state.PendingNoSubmit = false
	state.InitialDelivered = true
	return record.Save(*state)
}

// Open only opens an existing, created DSH session. A BrowserNone binding uses
// the system default browser for this explicit action without changing its pin.
// It never creates a provider session, submits/retries pending input, or consults
// a terminal runner.
func (s Service) Open(ctx context.Context, key string) (result Result, err error) {
	if s.Store == nil {
		return result, fmt.Errorf("DSH session store is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	err = s.Store.WithSession(ctx, key, func(record Record) error {
		state, found, err := record.Load()
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("DSH session binding is missing")
		}
		if err := validateState(state); err != nil {
			return err
		}
		browser, err := NormalizeBrowser(state.Browser)
		if err != nil {
			return fmt.Errorf("invalid DSH binding: %w", err)
		}
		if !state.Created {
			return fmt.Errorf("DSH session has not been created; retry its initial delivery")
		}
		result.SessionID = state.DSHSessionID
		if browser == BrowserNone {
			browser = "" // This explicit action does not change automatic opening policy.
		}
		if err := s.Client.Open(state.DSHSessionID, browser); err != nil {
			return err
		}
		result.BrowserOpened = true
		state.BrowserOpened = true
		if err := record.Save(state); err != nil {
			result.Warnings = append(result.Warnings, Warning{Code: "browser_checkpoint_failed", Message: "could not record DSH browser opening"})
		}
		return nil
	})
	return result, err
}
