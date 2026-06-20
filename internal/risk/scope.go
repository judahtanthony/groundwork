// Package risk classifies a gated action's scope into a risk class and a
// reversibility verdict — the two axes the gate engine composes (ADR 0014,
// ADR 0028). Risk ranks and explains; reversibility sets the floor. Both read
// the same Scope so the approval surface and the gate agree on one description.
//
// The scoring here is intentionally a coarse v1 heuristic; calibrated risk
// scoring and earned/revocable autonomy are Phase 5 refinements. Forcing
// `critical` on irreversible actions is the gate engine's job (ADR 0028), not
// this package's: ClassForScore never returns critical.
package risk

import (
	"path"
	"regexp"
	"strings"
)

// Scope describes the effects of a gated action. Files and Commands are derived
// from the change; the booleans are caller-supplied hints for effects that
// cannot be read from a diff (an action touching external/production state, a
// non-reversible migration, or explicit credential access).
type Scope struct {
	Files                 []string
	Commands              []string
	Network               bool
	External              bool
	IrreversibleMigration bool
	CredentialAccess      bool
}

// destructiveCommand matches shell commands whose effects cannot be reverted via
// git (data loss, force-push, filesystem/database destruction).
var destructiveCommand = regexp.MustCompile(`(?i)\b(rm\s+-rf|rmdir|dd\b|mkfs|drop\s+(table|database|schema)|truncate\b|git\s+push\s+.*--force|git\s+push\s+.*-f\b|--force\b|shutdown|reboot)\b`)

// HasDestructiveCommand reports whether the scope contains a destructive
// command, exposed for the policy "destructive" command category (ADR 0028).
func HasDestructiveCommand(s Scope) bool {
	return len(destructiveCommands(s.Commands)) > 0
}

// destructiveCommands returns the subset of cmds classified destructive.
func destructiveCommands(cmds []string) []string {
	var out []string
	for _, c := range cmds {
		if destructiveCommand.MatchString(c) {
			out = append(out, c)
		}
	}
	return out
}

// secretFile reports whether a path looks like a credential/secret file
// (env files or any path whose name contains "secret"). Heuristic, matched on
// the base name so directory prefixes do not matter.
func secretFile(file string) bool {
	base := strings.ToLower(path.Base(file))
	if strings.HasPrefix(base, ".env") {
		return true
	}
	return strings.Contains(base, "secret")
}

// hasSecretFile reports whether any file looks like a credential/secret file.
func hasSecretFile(files []string) bool {
	for _, f := range files {
		if secretFile(f) {
			return true
		}
	}
	return false
}
