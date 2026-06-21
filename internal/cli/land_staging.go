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

// promptYesDefault asks a [Y/n] question, defaulting to yes on empty input or EOF
// (so non-interactive callers get the documented default).
func promptYesDefault(in io.Reader, out io.Writer, question string) bool {
	fmt.Fprintf(out, "%s [Y/n] ", question)
	line, _ := bufio.NewReader(in).ReadString('\n')
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "n", "no":
		return false
	default:
		return true
	}
}
