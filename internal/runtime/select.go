package runtime

import "fmt"

// Select returns the Runtime named by the project config (config.runtime). It is
// the single place the coordinator chooses between the records-only stub and the
// Codex adapter (ADR 0027). An unknown name is an error rather than a silent
// fallback, so a misconfigured runtime fails loudly at boot.
func Select(name string, cfg Config) (Runtime, error) {
	switch name {
	case "", "stub":
		return Stub{}, nil
	case "codex":
		return NewCodex(cfg), nil
	default:
		return nil, fmt.Errorf("unknown runtime %q (want stub or codex)", name)
	}
}
