package actor

import "strings"

// Identity and capability matching for actor-aware policy (ADR 0029).
//
// Identity is a dotted path of authority/grouping tiers, prefix-matchable at any
// depth, bottoming out in a unique instance (`ai`, `ai.codex`,
// `ai.codex.default`). The path carries authority/grouping only — never role or
// capability. The meaning of intermediate tiers is conventional, not enforced,
// and the id root is not validated against `type`: `type` stays a separate
// matchable dimension (the bootstrap uses `ai.*` ids with `ai_agent`/`ai_judge`
// types). Capabilities are an orthogonal property set, matched independently.

// MatchID reports whether id matches the dotted-path pattern: a pattern matches
// itself or any deeper-tier descendant. "ai" matches "ai" and "ai.codex.default"
// but not "ainsley"; an empty pattern matches nothing.
func MatchID(pattern, id string) bool {
	if pattern == "" {
		return false
	}
	if pattern == id {
		return true
	}
	return strings.HasPrefix(id, pattern+".")
}

// Match returns the actors whose id matches the dotted-path pattern, in registry
// order.
func (r *Registry) Match(pattern string) []*Actor {
	var out []*Actor
	for i := range r.Actors {
		if MatchID(pattern, r.Actors[i].ID) {
			out = append(out, &r.Actors[i])
		}
	}
	return out
}

// Resolve maps a requested actor (a class request: an exact id or a dotted
// prefix) to a concrete instance. An exact id wins; otherwise the first actor
// matching the prefix is returned (one instance per class in v1; pools are a
// later generalization, ADR 0029). An empty request resolves to nothing.
func (r *Registry) Resolve(requested string) (*Actor, bool) {
	if requested == "" {
		return nil, false
	}
	if a, ok := r.Get(requested); ok {
		return a, true
	}
	if matches := r.Match(requested); len(matches) > 0 {
		return matches[0], true
	}
	return nil, false
}

// AnyIDMatch reports whether id matches any of the dotted-path patterns.
func AnyIDMatch(patterns []string, id string) bool {
	for _, p := range patterns {
		if MatchID(p, id) {
			return true
		}
	}
	return false
}

// HasAnyRole reports whether the actor holds any of the given roles.
func (a *Actor) HasAnyRole(roles []string) bool {
	for _, r := range roles {
		if a.HasRole(r) {
			return true
		}
	}
	return false
}

// CanClaim reports whether the actor may claim the given work type. A "*" entry
// in capabilities.work_types is a wildcard.
func (a *Actor) CanClaim(workType string) bool {
	return contains(a.Capabilities.WorkTypes, workType)
}

// CanReview reports whether the actor may review the given work type.
func (a *Actor) CanReview(workType string) bool {
	return contains(a.Capabilities.Review, workType)
}

// CanApprove reports whether the actor may approve the given work type.
func (a *Actor) CanApprove(workType string) bool {
	return contains(a.Capabilities.Approve, workType)
}

// HasRole reports whether the actor holds the given role.
func (a *Actor) HasRole(role string) bool {
	for _, r := range a.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// contains reports whether want is in set, treating "*" as a wildcard match.
func contains(set []string, want string) bool {
	for _, v := range set {
		if v == "*" || v == want {
			return true
		}
	}
	return false
}
