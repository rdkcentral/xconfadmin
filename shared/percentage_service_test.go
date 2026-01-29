package shared

import "testing"

func TestCalculatePercentBounds(t *testing.T) {
	p := CalculatePercent("sample")
	if p < 0 || p > 100 {
		t.Fatalf("percent out of bounds: %d", p)
	}
	// deterministic for same input
	p2 := CalculatePercent("sample")
	if p != p2 {
		t.Fatalf("expected deterministic percent")
	}
}

func TestCalculateHashAndPercentRelationship(t *testing.T) {
	hash, pct := CalculateHashAndPercent("another")
	if hash <= 0 {
		t.Fatalf("expected positive hash")
	}
	if pct < 0 || pct > 100 {
		t.Fatalf("percent out of range: %f", pct)
	}
}
