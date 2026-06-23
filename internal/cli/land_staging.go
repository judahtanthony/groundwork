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

// previewStager is the read surface `gw ticket land --preview` needs, so the
// preview is unit-testable without a real repo.
type previewStager interface {
	HasStagedChanges() (bool, error)
	StagedDiff() (string, error)
}

// previewLanding shows the staged change set a landing of id would commit,
// without mutating the index or contacting the coordinator (ADR 0041). The
// ticket's regenerated export is added by the coordinator at commit time, so it
// is noted rather than shown here.
func previewLanding(ctx *Context, id string, repo previewStager) error {
	staged, err := repo.HasStagedChanges()
	if err != nil {
		return &Error{Code: "git_error", Message: err.Error()}
	}
	if !staged {
		if ctx.JSON {
			return ctx.PrintJSON(map[string]any{"id": id, "staged": false, "diff": ""})
		}
		fmt.Fprintf(ctx.Stdout, "No staged changes for %s. Stage files (git add …) or land with --all.\n", id)
		return nil
	}
	diff, err := repo.StagedDiff()
	if err != nil {
		return &Error{Code: "git_error", Message: err.Error()}
	}
	if ctx.JSON {
		return ctx.PrintJSON(map[string]any{"id": id, "staged": true, "diff": diff})
	}
	fmt.Fprintf(ctx.Stdout, "Staged changes landing %s would commit (plus its regenerated export):\n\n", id)
	fmt.Fprint(ctx.Stdout, diff)
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
