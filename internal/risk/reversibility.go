package risk

// Reversible reports whether an action's scope is cheaply reversible (revertible
// via git, no external side effects) and, when not, the reasons it is
// irreversible (ADR 0014). An irreversible action is forced to `critical` by the
// gate engine regardless of its risk score; this function only classifies, it
// does not gate.
func Reversible(s Scope) (bool, []string) {
	var reasons []string
	if s.External {
		reasons = append(reasons, "touches external or production state")
	}
	if s.IrreversibleMigration {
		reasons = append(reasons, "non-reversible migration without a tested rollback")
	}
	if s.CredentialAccess || hasSecretFile(s.Files) {
		reasons = append(reasons, "credential or secret access")
	}
	if cmds := destructiveCommands(s.Commands); len(cmds) > 0 {
		reasons = append(reasons, "destructive command: "+cmds[0])
	}
	return len(reasons) == 0, reasons
}
