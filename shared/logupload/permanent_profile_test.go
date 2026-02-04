package logupload

import (
	"testing"

	"github.com/rdkcentral/xconfwebconfig/shared"
)

// TestNewEmptyPermanentTelemetryProfile ensures defaults applied.
func TestNewEmptyPermanentTelemetryProfile(t *testing.T) {
	prof := NewEmptyPermanentTelemetryProfile()
	if prof.Type != PermanentTelemetryProfileConst {
		t.Fatalf("expected type %s got %s", PermanentTelemetryProfileConst, prof.Type)
	}
	if prof.ApplicationType != shared.STB {
		t.Fatalf("expected application type %s got %s", shared.STB, prof.ApplicationType)
	}
}
