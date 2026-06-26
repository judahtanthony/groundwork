package approval

import "testing"

// The authority-elevation action types are recognized approval types (ADR 0038).
func TestAuthorityElevationTypesValid(t *testing.T) {
	for _, ty := range []Type{TypeAmendPolicy, TypeElevateAutonomy} {
		if !ty.Valid() {
			t.Errorf("type %q should be valid", ty)
		}
	}
	if Type("bogus").Valid() {
		t.Error("bogus type should be invalid")
	}
}
