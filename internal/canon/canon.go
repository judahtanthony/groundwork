// Package canon implements the read/write plumbing for canon-as-memory
// (ADR 0013): the per-node journal (tier-1 ephemeral, ignored) and the
// ratification hook that records when durable design becomes binding (a
// decomposition accepted, a node landed, a policy/SOP change approved). In M2
// this is the plumbing and the forward parent-contract channel; the agent-
// authored distillation *content* arrives with the Phase 4 runtime (ADR 0030).
//
// Writes are serialized through the coordinator (single process), keeping canon
// conflict-free; this package performs the file I/O.
package canon

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// JournalEntry is one append-only decision note for a node.
type JournalEntry struct {
	At    string `json:"at"`
	Gate  string `json:"gate,omitempty"` // set for ratification entries
	Kind  string `json:"kind,omitempty"` // set for typed entries (e.g. "context_miss")
	Entry string `json:"entry"`
}

// journalPath is the per-node journal file (one file per node so parallel runs
// never conflict, ADR 0013).
func journalPath(journalDir, nodeID string) string {
	return filepath.Join(journalDir, nodeID+".ndjson")
}

// Append adds a decision note to a node's journal, creating the journal
// directory if needed.
func Append(journalDir, nodeID, entry string) error {
	return appendEntry(journalDir, nodeID, JournalEntry{At: now(), Entry: entry})
}

// Ratify records that durable design was ratified at a gate for a node (the
// ratification hook). M2 records the event; promoting distilled content into
// canonical documents is the Phase 4 runtime's job.
func Ratify(journalDir, nodeID, gate, note string) error {
	return appendEntry(journalDir, nodeID, JournalEntry{At: now(), Gate: gate, Entry: note})
}

// Miss records a context-miss for a node: something the worker needed that the
// context brief did not provide (ADR 0035, ADR 0013). Misses are ignored runtime
// state; the review step promotes recurring ones into canon (SOPs, docs, brief).
func Miss(journalDir, nodeID, note string) error {
	return appendEntry(journalDir, nodeID, JournalEntry{At: now(), Kind: "context_miss", Entry: note})
}

// Misses returns a node's recorded context-misses, oldest first.
func Misses(journalDir, nodeID string) ([]JournalEntry, error) {
	entries, err := Read(journalDir, nodeID)
	if err != nil {
		return nil, err
	}
	var out []JournalEntry
	for _, e := range entries {
		if e.Kind == "context_miss" {
			out = append(out, e)
		}
	}
	return out, nil
}

func appendEntry(journalDir, nodeID string, e JournalEntry) error {
	if err := os.MkdirAll(journalDir, 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(journalPath(journalDir, nodeID), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	line, err := json.Marshal(e)
	if err != nil {
		return err
	}
	_, err = f.Write(append(line, '\n'))
	return err
}

// Read returns a node's journal entries, oldest first. A missing journal is an
// empty slice (not an error): journals are ephemeral.
func Read(journalDir, nodeID string) ([]JournalEntry, error) {
	f, err := os.Open(journalPath(journalDir, nodeID))
	if err != nil {
		if os.IsNotExist(err) {
			return []JournalEntry{}, nil
		}
		return nil, err
	}
	defer f.Close()
	var out []JournalEntry
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var e JournalEntry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, sc.Err()
}

// Reconcile merges a composite parent's existing canon contribution with new
// contributions from its children, de-duplicating identical lines so canon stays
// coherent and non-redundant (work-tree.md). The M2 implementation is a
// line-level union preserving order; richer semantic reconciliation is authored
// by the Phase 4 runtime.
func Reconcile(existing string, contributions []string) string {
	seen := map[string]bool{}
	var lines []string
	add := func(block string) {
		for _, ln := range strings.Split(block, "\n") {
			key := strings.TrimSpace(ln)
			if key == "" || seen[key] {
				continue
			}
			seen[key] = true
			lines = append(lines, ln)
		}
	}
	add(existing)
	for _, c := range contributions {
		add(c)
	}
	return strings.Join(lines, "\n")
}

func now() string { return time.Now().UTC().Format(time.RFC3339) }
