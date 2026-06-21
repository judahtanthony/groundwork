package cli

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// landStager is the subset of *git.Repo that resolveLandStaging needs, so the
// staging decision is unit-testable without a real repo.
type landStager interface {
	HasStagedChanges() (bool, error)
	HasUncommitted() (bool, error)
	AddAll() error
}

// resolveLandStaging decides what the landing commit will include, before the
// coordinator commits the index (ADR 0034).
//
//   - all == true: stage every change (git add -A) — the `git commit -am` ergonomic.
//   - otherwise, if the index already has staged changes, leave it: that is the
//     human's explicit ticket-scoped pathspec.
//   - otherwise, if nothing is staged but the work tree has changes, ask whether
//     to include them all (default yes); on yes, stage everything.
//
// When the index ends up empty (declined, or nothing to stage) the coordinator
// records the landing without forcing an empty commit.
func resolveLandStaging(repo landStager, all bool, in io.Reader, out io.Writer) error {
	if all {
		return repo.AddAll()
	}
	staged, err := repo.HasStagedChanges()
	if err != nil {
		return err
	}
	if staged {
		return nil
	}
	any, err := repo.HasUncommitted()
	if err != nil {
		return err
	}
	if !any {
		return nil
	}
	if promptYesDefault(in, out, "Nothing is staged for this landing. Include all changes?") {
		return repo.AddAll()
	}
	return nil
}

// promptYesDefault asks a [Y/n] question. A present human gets default-yes (empty
// Enter accepts). A non-interactive caller — EOF with no input (piped stdin, CI) —
// gets the safe default of NO: in M3 landings run against the shared working tree,
// so staging everything unattended could sweep in unrelated work. (A Phase 4
// isolated per-ticket worktree run reverses this, since the sandbox holds only the
// ticket's work; see ADR 0034.)
func promptYesDefault(in io.Reader, out io.Writer, question string) bool {
	fmt.Fprintf(out, "%s [Y/n] ", question)
	line, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && line == "" {
		return false // EOF / non-interactive: no human to answer
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "n", "no":
		return false
	default:
		return true // empty Enter or "y": a present human accepts the default
	}
}
