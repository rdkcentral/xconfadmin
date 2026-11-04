package http

import (
	"testing"

	"github.com/rdkcentral/xconfadmin/common"
)

func TestNormalizeApplicationType(t *testing.T) {
	previous := common.ApplicationTypes
	common.ApplicationTypes = []string{"stb", "sky", "rdkcloud", "xhome"}
	t.Cleanup(func() { common.ApplicationTypes = previous })

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "configured", input: "xhome", want: "xhome"},
		{name: "unknown", input: "attacker-controlled-value", want: "unknown"},
		{name: "empty", input: "", want: "unknown"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := normalizeApplicationType(test.input); got != test.want {
				t.Fatalf("normalizeApplicationType(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}
