package server

// nearestInChain returns the first non-nil result of lookup applied to nodeID
// and then its ancestors (nearest parent first), or (nil, nil) when none match.
// It is the shared "self, then closest ancestor" resolution used to find the
// governing active envelope (ADR 0056) and the integration target (ADR 0058).
func nearestInChain[T any](s *Server, nodeID string, lookup func(id string) (*T, error)) (*T, error) {
	if v, err := lookup(nodeID); err != nil || v != nil {
		return v, err
	}
	ancestors, err := s.db.Ancestors(nodeID)
	if err != nil {
		return nil, err
	}
	for i := len(ancestors) - 1; i >= 0; i-- { // nearest parent first
		if v, err := lookup(ancestors[i].ID); err != nil || v != nil {
			return v, err
		}
	}
	return nil, nil
}
