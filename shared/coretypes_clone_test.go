package shared

import "testing"

// TestEnvironmentClone verifies deep copy semantics for Environment.
func TestEnvironmentClone(t *testing.T) {
	env := NewEnvironment("Env_1", "Description")
	clone, err := env.Clone()
	if err != nil {
		t.Fatalf("clone failed: %v", err)
	}
	if clone.ID != env.ID || clone.Description != env.Description || clone.Updated != env.Updated {
		t.Fatalf("clone fields mismatch")
	}
	// modify original to ensure clone independent
	env.Description = "Changed"
	if clone.Description == env.Description {
		t.Fatalf("expected clone description independent from original")
	}
}
