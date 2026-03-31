package shared

import "testing"

// TestCalculatePercentEmptyAndDifference ensures empty string still yields bounded percent
// and different inputs generally produce different percents.
func TestCalculatePercentEmptyAndDifference(t *testing.T) {
    pEmpty := CalculatePercent("")
    if pEmpty < 0 || pEmpty > 100 {
        t.Fatalf("empty percent out of bounds: %d", pEmpty)
    }
    pA := CalculatePercent("AAAA")
    pB := CalculatePercent("BBBB")
    if pA == pB { // extremely unlikely; flag if happens
        t.Fatalf("expected different percents for different inputs: %d == %d", pA, pB)
    }
}

// TestCalculateHashAndPercentDeterminism verifies same input gives same outputs.
func TestCalculateHashAndPercentDeterminism(t *testing.T) {
    h1, pct1 := CalculateHashAndPercent("deterministic-value")
    h2, pct2 := CalculateHashAndPercent("deterministic-value")
    if h1 != h2 || pct1 != pct2 {
        t.Fatalf("expected deterministic results: (%f,%f) vs (%f,%f)", h1, pct1, h2, pct2)
    }
}
