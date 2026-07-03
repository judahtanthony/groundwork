package scheduler

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"groundwork/internal/encoding"
	"groundwork/internal/resume"
	"groundwork/internal/runtime"
)

// buildPrompt assembles the task instruction for an agent run from the node's
// durable context (ADR 0051/0047): title, ancestor contract, acceptance criteria,
// open blockers, and the recommended next step. It is handed to the runtime as
// spec.Prompt (e.g. `codex exec "<prompt>"`).
func (s *Scheduler) buildPrompt(ticketID string) string {
	p, err := resume.Assemble(s.db, ticketID)
	if err != nil || p == nil {
		return "Implement the assigned work item in the current directory (an isolated git worktree)."
	}
	var b strings.Builder
	b.WriteString("You are an autonomous coding agent working in an isolated git worktree (the current directory).\n\n")
	fmt.Fprintf(&b, "Ticket %s: %s\n", p.TicketID, p.Title)
	if p.WorkType != "" {
		fmt.Fprintf(&b, "Work type: %s\n", p.WorkType)
	}
	if p.AncestorContract != "" {
		fmt.Fprintf(&b, "\nParent contract (the boundary you implement within):\n%s\n", p.AncestorContract)
	}
	if len(p.Acceptance) > 0 {
		b.WriteString("\nAcceptance criteria:\n")
		for _, a := range p.Acceptance {
			fmt.Fprintf(&b, "- %s\n", a)
		}
	}
	if len(p.PendingBlockers) > 0 {
		b.WriteString("\nOpen questions/blockers recorded on this ticket:\n")
		for _, r := range p.PendingBlockers {
			msg := r.Statement
			if msg == "" {
				msg = r.HandoffSummary
			}
			fmt.Fprintf(&b, "- %s\n", msg)
		}
	}
	if p.NextAction != "" {
		fmt.Fprintf(&b, "\nRecommended next step: %s\n", p.NextAction)
	}
	b.WriteString("\nMake the necessary changes in this directory to satisfy the acceptance criteria. " +
		"Keep changes minimal and within scope. Do not run git commit — the coordinator checkpoints and lands your work.\n")
	return b.String()
}

// runEventLine is one canonical JSON line in a run's events.ndjson (ADR 0027).
// The JSONL log is tier-1 ignored runtime evidence under .groundwork/runs; SQLite
// holds the queryable projection (run_events).
type runEventLine struct {
	Time    string         `json:"time"`
	Type    string         `json:"type"`
	Message string         `json:"message,omitempty"`
	Payload map[string]any `json:"payload,omitempty"`
}

// writeRunDiff persists a run's full unified diff as evidence under
// <dir>/<runID>/diff.patch (ADR 0059). A blank dir or diff is a no-op.
func writeRunDiff(dir, runID, diff string) error {
	if dir == "" || diff == "" {
		return nil
	}
	runDir := filepath.Join(dir, runID)
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(runDir, "diff.patch"), []byte(diff), 0o644)
}

// appendRunEventLog appends one event to <dir>/<runID>/events.ndjson, creating
// the run directory as needed. A blank dir disables the local log (e.g. tests).
func appendRunEventLog(dir, runID string, ev runtime.Event) error {
	if dir == "" {
		return nil
	}
	runDir := filepath.Join(dir, runID)
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Join(runDir, "events.ndjson"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	line, err := json.Marshal(runEventLine{
		Time: encoding.Now(), Type: ev.Type, Message: ev.Message, Payload: ev.Payload,
	})
	if err != nil {
		return err
	}
	if _, err := f.Write(append(line, '\n')); err != nil {
		return err
	}
	return nil
}
